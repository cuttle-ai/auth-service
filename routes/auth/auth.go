// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Package auth contains the handlers required for authentication purposes
package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cuttle-ai/auth-service/config"
	"github.com/cuttle-ai/auth-service/oauth"
	"github.com/cuttle-ai/auth-service/oauth/google"
	"github.com/cuttle-ai/auth-service/routes"
	"github.com/cuttle-ai/auth-service/routes/response"
	"golang.org/x/oauth2"
)

func getUserInfo(appCtx *config.AppContext) (*config.UserInfo, error) {

	//getting the user model
	if appCtx.Session.User == nil || len(appCtx.Session.User.AccessToken) == 0 {
		return nil, nil
	}
	agent := getAgent(appCtx.Session.User)

	//getting the updated user info from the auth agent
	info, err := oauth.Info(oauth2.NoContext, appCtx, agent)
	if err != nil {
		//error while getting the user info from the auth agent
		return nil, err
	}

	return info, nil
}

//Urls will return the 3party auth URLs
func Urls(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	response.Write(appCtx, w, map[string]string{
		"Google": google.Config.AuthCodeURL("state", oauth2.AccessTypeOffline),
	})
}

//GoogleAuth is the callback url for the Google OAuth
func GoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * We will get the google auth code from the request
	 * Will get the token from the google auth exchange
	 * We will initiate the user session and save the session
	 * We will get user info from the auth agent
	 * We will also info the user logged info info to all the applications
	 * Then will redirect to the index page
	 */
	//we will get the code from the request
	code := r.URL.Query().Get("code")
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)

	//we will get the token from the google exchange
	tok, err := google.Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		//error while getting the token from the google auth exchange
		appCtx.Log.Error("Error while getting the token for the code", code)
		appCtx.Log.Error(err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Sorry couldn't complete your oauth"}, http.StatusForbidden)
		return
	}

	//we will set the user
	user := &config.User{}
	user.AccessToken = tok.AccessToken
	user.AuthAgent = oauth.GOOGLE
	appCtx.Session.User = user

	//getting the user info from the auth agent
	agent := getAgent(appCtx.Session.User)
	info, err := getUserInfo(appCtx)
	if err != nil {
		//error while getting the user info from the auth agent
		appCtx.Log.Error("Error while fetching the user info from oauth agent", agent.Name())
		appCtx.Log.Error(err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Sorry couldn't complete your oauth"}, http.StatusForbidden)
		return
	}

	//if the info is nil, we know that the session is empty
	if info == nil {
		response.WriteError(appCtx, w, response.Error{Err: "Sorry couldn't complete your oauth"}, http.StatusForbidden)
		return
	}

	//will save the session
	appCtx.Session.Authenticated = true
	appCtx.Session.User.Email = info.Email
	appCtx.Session.User.AccessToken = appCtx.Session.ID
	go routes.SendRequest(routes.AppContextRequestChan, routes.AppContextRequest{
		Session: appCtx.Session,
		Type:    routes.SetSession,
	})

	//informing the user logged in info to all the applications
	go appCtx.Session.User.InformAuth(*appCtx, true)
	http.SetCookie(w, &http.Cookie{
		Name:    config.AuthHeaderKey,
		Value:   appCtx.Session.ID,
		Expires: time.Now().AddDate(0, 0, 1),
		Domain:  strings.Split(config.FrontendURL, ":")[0],
		Path:    "/",
	})

	//since we have a valid info, we will get the info from db
	//if the info is empty, we have to update the db with new info
	//if info is not empty, except the registered info we will update the existing info in db
	i := (*info).Get(*appCtx)
	if i == nil {
		(*info).Insert(*appCtx)
	} else {
		info.Registered = i.Registered
		(*info).Update(*appCtx)
	}

	//will rediect to the index page
	response.Write(appCtx, w, appCtx.Session)
}

//Register registers the user with the platform.
//One can register only if he agrees the terms and conditions
//while subscribing to the newsletter is optional
func Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * First we will get the user session
	 * If the session doesn't exist we will give session expired message
	 * Then check whether the user has agreed to the terms and conditions of the application
	 * Will parse the subscribe parameter
	 * If yes we will update the profile information
	 */
	//getting the user session
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)

	//session expired if no session
	if !appCtx.Session.Authenticated || appCtx.Session.User == nil {
		appCtx.Log.Error("User session expired for while registering by accepting the terms and conditions")
		response.WriteError(appCtx, w, response.Error{Err: response.ErrorCodes[response.ErrorCodeSessionExpired]}, http.StatusForbidden)
		return
	}

	//checking whether the user has agreed to the terms and conditions
	if r.FormValue("agree") != "true" {
		appCtx.Log.Error("User hasn't agreed to the terms and conditions")
		response.WriteError(appCtx, w, response.Error{Err: "Please agree to our terms and condition"}, http.StatusExpectationFailed)
		return
	}

	//form validation for subscribe to newletter
	unPSubs := r.FormValue("subscribe")
	if unPSubs != "true" && unPSubs != "false" {
		appCtx.Log.Error("invalid form value provided. Was expecting true or false. Got ", unPSubs)
		response.WriteError(appCtx, w, response.Error{Err: response.ErrorCodes[response.ErrorCodeInvalidParams] + " subscribe. Expecting true or false"}, http.StatusUnprocessableEntity)
		return
	}
	subscribed := false
	if unPSubs == "true" {
		subscribed = true
	}

	//updating the profile info
	pro := config.UserInfo{Email: appCtx.Session.User.Email}
	pro = *(pro.Get(*appCtx))
	pro.Registered = true
	pro.Subscribed = subscribed
	pro.Update(*appCtx)

	//sending success message
	response.Write(appCtx, w, "Successfully registered the user")
}

//Session returns the session information of the user
func Session(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	response.Write(appCtx, w, appCtx.Session)
}

//Profile returns the profile information of the user
func Profile(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	u := &config.UserInfo{}
	if appCtx.Session.User != nil {
		u.Email = appCtx.Session.User.Email
		u = u.Get(*appCtx)
	}
	response.Write(appCtx, w, u)
}

//Logout logs a user out of the platform
func Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * We will inform all the services that the user has logged out
	 * We will empty the session
	 * Then saves the session
	 */
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)

	//inform all the services that the user has logged out
	go appCtx.Session.User.InformAuth(*appCtx, false)

	//We will delete the user model from the session
	appCtx.Session.Authenticated = false
	appCtx.Session.User = nil

	//will save the session
	go routes.SendRequest(routes.AppContextRequestChan, routes.AppContextRequest{
		Session: appCtx.Session,
		Type:    routes.SetSession,
	})
	http.SetCookie(w, &http.Cookie{
		Name:    config.AuthHeaderKey,
		Expires: time.Now(),
		Domain:  strings.Split(config.FrontendURL, ":")[0],
		Path:    "/",
	})

	//send the ok response
	response.Write(appCtx, w, "you have sucessfully logged out of the system")
}

func init() {
	routes.AddRoutes(
		routes.Route{
			Version:     "v1",
			Pattern:     "/auth/urls",
			HandlerFunc: Urls,
		},
		routes.Route{
			Version:     "v1",
			Pattern:     "/auth/google",
			HandlerFunc: GoogleAuth,
		},
		routes.Route{
			Version:     "v1",
			Pattern:     "/auth/register",
			HandlerFunc: Register,
		},
		routes.Route{
			Version:     "v1",
			Pattern:     "/auth/session",
			HandlerFunc: Session,
		},
		routes.Route{
			Version:     "v1",
			Pattern:     "/auth/profile",
			HandlerFunc: Profile,
		},
		routes.Route{
			Version:     "v1",
			Pattern:     "/auth/logout",
			HandlerFunc: Logout,
		},
	)
}

func getAgent(u *config.User) oauth.Agent {
	if u == nil {
		return nil
	}
	switch u.AuthAgent {
	case oauth.GOOGLE:
		return &google.Agent{}
	}
	return nil
}
