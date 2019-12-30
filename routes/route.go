// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routes

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cuttle-ai/auth-service/log"
	"github.com/cuttle-ai/auth-service/routes/response"

	"github.com/cuttle-ai/auth-service/version"

	"github.com/cuttle-ai/auth-service/config"
)

/*
 * This file has the definition of route data structure
 */

//HandlerFunc is the Handler func with the context
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

//Route is a route with explicit versions
type Route struct {
	//Version is the version of the route
	Version string
	//Pattern is the url pattern of the route
	Pattern string
	//HandlerFunc is the handler func of the route
	HandlerFunc HandlerFunc
}

type key string

//AppContextKey is the key with which the application is saved in the request context
const AppContextKey key = "app-context"

//Register registers the route with the default http handler func
func (r Route) Register(s *http.ServeMux) {
	/*
	 * If the route version is default version then will register it without version string to http handler
	 * Will register the router with the http handler
	 */
	if r.Version == version.Default.API {
		s.Handle(r.Pattern, http.TimeoutHandler(r, config.ResponseTimeout, "timeout"))
	}
	s.Handle("/"+r.Version+r.Pattern, http.TimeoutHandler(r, config.ResponseTimeout, "timeout"))
}

//ServeHTTP implements HandlerFunc of http package. It makes use of the context of request
func (r Route) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	/*
	 * We will set the cors
	 * Will get the context
	 * Will parse the form
	 * Will get the auth token from the request header
	 * We will fetch the app context for the request
	 * If app contexts have exhausted, we will reject the request
	 * Then we will set the app context in request
	 * Execute request handler func
	 * After execution return the app context
	 */
	//setting the cors
	res.Header().Set("Access-Control-Allow-Origin", config.FrontendURL)
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	res.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	if req.Method == "OPTIONS" {
		return
	}

	//getting the context
	ctx := req.Context()

	//parsing the form
	err := req.ParseForm()
	if err != nil {
		//error while parsing the form
		log.Error("Error while parsing the request form", err)
		response.WriteError(nil, res, response.Error{Err: "Couldn't parse the request form"}, http.StatusUnprocessableEntity)
		_, cancel := context.WithCancel(ctx)
		cancel()
		return
	}

	//getting the auth token from the header
	cookie, cErr := req.Cookie(config.AuthHeaderKey)
	if cErr != nil {
		log.Warn("Auth cookie not found")
		cookie = &http.Cookie{Name: config.AuthHeaderKey}
	}
	auth := cookie.Value

	//fetching the app context
	appCtxReq := AppContextRequest{
		Type:    Get,
		Out:     make(chan AppContextRequest),
		Session: config.Session{ID: auth},
	}
	go SendRequest(AppContextRequestChan, appCtxReq)
	resCtx := <-appCtxReq.Out

	//checking whether the app context exhausted or not
	if resCtx.Exhausted {
		//reject the request
		log.Error("We have exhausted the request limits")
		response.WriteError(resCtx.AppContext, res, response.Error{Err: "We have exhuasted the server request limits. Please try after some time."}, http.StatusTooManyRequests)
		_, cancel := context.WithCancel(ctx)
		cancel()
		return
	}

	//setting the app context
	newCtx := context.WithValue(ctx, AppContextKey, resCtx.AppContext)
	if resCtx.AppContext.Session.ID != auth {
		cookie.Expires = time.Now()
		cookie.Domain = strings.Split(config.FrontendURL, ":")[0]
		cookie.Path = "/"
		cookie.Value = resCtx.Session.ID
		http.SetCookie(res, cookie)
	}

	resCtx.AppContext.Log.Info("Request URL ", req.URL.RequestURI())

	//executing the request
	r.Exec(newCtx, res, req)

	//returning the app context
	appCtxReq = AppContextRequest{
		Type:       Finished,
		AppContext: resCtx.AppContext,
	}
	go SendRequest(AppContextRequestChan, appCtxReq)
}

//Exec will execute the handler func. By default it will set response content type as as json.
//It will also cancel the context at the end. So no need of explicitly invoking the same in the handler funcs
func (r Route) Exec(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	/*
	 * Will get the cancel for the context
	 * Will execute the handlerfunc
	 * Cancelling the context at the end
	 */
	//getting the context cancel
	c, cancel := context.WithCancel(ctx)

	//executing the handler
	r.HandlerFunc(c, res, req)

	//cancelling the context
	cancel()
}
