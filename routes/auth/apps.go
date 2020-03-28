// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cuttle-ai/auth-service/config"
	"github.com/cuttle-ai/auth-service/routes"
	"github.com/cuttle-ai/auth-service/routes/response"
	"github.com/google/uuid"
)

//GetApps api will return the list of apps registered by the user in the system
func GetApps(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * First we will get the app context
	 * Then we will get the apps the user has created
	 * Return the response
	 */
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	u := appCtx.Session.User.ToUserInfo()

	apps, err := u.GetApps(*appCtx)
	if err != nil {
		//error while getting the apps the user has registered with the platform
		appCtx.Log.Error("Error while fetching app registered by the user", u.ID)
		appCtx.Log.Error(err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Sorry fetch the apps"}, http.StatusInternalServerError)
		return
	}

	response.Write(appCtx, w, response.Message{Message: "fetched the list", Data: apps})
}

//CreateApp api will create an app registered with the platform
func CreateApp(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * First we will get the app context
	 * Then we will parse the request
	 * Then we will create the app
	 * Then will inform the authentication across the platform
	 * Return the response
	 */
	//getting the app context
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	appCtx.Log.Info("a request has come to create the app from ", appCtx.Session.User.ID)

	//parse the request param
	a := &config.App{}
	err := json.NewDecoder(r.Body).Decode(a)
	if err != nil {
		//bad request
		appCtx.Log.Error("error while parsing the app param", err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Invalid Params " + err.Error()}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//creating the app
	a.UID = uuid.New()
	a.AccessToken = uuid.New().String()
	a.UserID = appCtx.Session.User.ID
	aI := a.ToAppInfo()
	err = (&aI).Insert(*appCtx)
	if err != nil {
		//error while inserting the app to the platform
		appCtx.Log.Error("error while inserting the app into db for", aI.UserID)
		appCtx.Log.Error(err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Couldn't create the app"}, http.StatusInternalServerError)
		return
	}

	//informing the authentication across the platform
	appCtx.Log.Info("created the app for user - ", a.UserID, "with id", a.ID, "going to update the same across the platform")
	user := aI.ToApp().ToUser()
	go user.InformAuth(*appCtx, true)

	//we will write the response
	response.Write(appCtx, w, response.Message{Message: "created the app", Data: aI.ToApp()})
}

//UpdateApp api will update an app registered with the platform. Only name, email and description are updated
func UpdateApp(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * First we will get the app context
	 * Then we will parse the request
	 * Then we will update the app
	 * Return the response
	 */
	//getting the app context
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	appCtx.Log.Info("a request has come to update the app from ", appCtx.Session.User.ID)

	//parse the request param
	a := &config.App{}
	err := json.NewDecoder(r.Body).Decode(a)
	if err != nil {
		//bad request
		appCtx.Log.Error("error while parsing the app param", err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Invalid Params " + err.Error()}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//updating the app
	if appCtx.Session.User.UserType != config.AdminUser && appCtx.Session.User.UserType != config.SuperAdmin {
		a.UserID = appCtx.Session.User.ID
	}
	aI := a.ToAppInfo()
	err = (&aI).Update(*appCtx)
	if err != nil {
		//error while updating the app in the platform
		appCtx.Log.Error("error while updating the app in the db for", aI.UserID, aI.ID)
		appCtx.Log.Error(err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Couldn't update the app"}, http.StatusInternalServerError)
		return
	}

	//we will write the response
	appCtx.Log.Info("updated the app for user - ", a.UserID, "with id", a.ID)
	response.Write(appCtx, w, response.Message{Message: "updated the app", Data: aI.ToApp()})
}

//DeleteApp api will delete an app registered with the platform.
func DeleteApp(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * First we will get the app context
	 * Then we will parse the request
	 * Then we will delete the app
	 * Return the response
	 */
	//getting the app context
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	appCtx.Log.Info("a request has come to delete the app from ", appCtx.Session.User.ID)

	//parse the request param
	a := &config.App{}
	err := json.NewDecoder(r.Body).Decode(a)
	if err != nil {
		//bad request
		appCtx.Log.Error("error while parsing the app param", err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Invalid Params " + err.Error()}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//deleting the app
	aI := a.ToAppInfo()
	err = (&aI).Delete(*appCtx)
	if err != nil {
		//error while deleting the app from the platform
		appCtx.Log.Error("error while deleting the app from the db for", aI.UserID, aI.ID)
		appCtx.Log.Error(err.Error())
		response.WriteError(appCtx, w, response.Error{Err: "Couldn't delete the app"}, http.StatusInternalServerError)
		return
	}

	//we will write the response
	appCtx.Log.Info("delete the app for user - ", a.UserID, "with id", a.ID, "going to update the same across the platform")
	user := aI.ToApp().ToUser()
	go user.InformAuth(*appCtx, false)
	response.Write(appCtx, w, response.Message{Message: "deleted the app", Data: nil})
}

//GetAllApps api will return the list of all apps registered in the platform. This is intented for admin use
func GetAllApps(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	/*
	 * First we will get the app context
	 * Then we will get all the apps
	 * Return the response
	 */
	appCtx := ctx.Value(routes.AppContextKey).(*config.AppContext)
	apps := config.GetAllApps(*appCtx)
	response.Write(appCtx, w, apps)
}

func init() {
	routes.AddRoutes(
		routes.Route{
			Version:       "v1",
			Pattern:       "/auth/apps",
			HandlerFunc:   GetApps,
			Authenticated: true,
		},
		routes.Route{
			Version:       "v1",
			Pattern:       "/auth/apps/create",
			HandlerFunc:   CreateApp,
			Authenticated: true,
		},
		routes.Route{
			Version:       "v1",
			Pattern:       "/auth/apps/update",
			HandlerFunc:   UpdateApp,
			Authenticated: true,
		},
		routes.Route{
			Version:       "v1",
			Pattern:       "/auth/apps/delete",
			HandlerFunc:   DeleteApp,
			Authenticated: true,
		},
		routes.Route{
			Version:       "v1",
			Pattern:       "/auth/admin/apps",
			HandlerFunc:   GetAllApps,
			ForAdmin:      true,
			Authenticated: true,
		},
	)
}
