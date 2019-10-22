// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package response

import (
	"html/template"
	"net/http"

	"github.com/cuttle-ai/auth-service/config"
	"github.com/cuttle-ai/auth-service/log"
)

/*
 * This file contains the html template response implementation
 */

//Template is Template Format with templatre name
type Template struct {
	//T is template
	T *template.Template
	//Name is the template
	Name string
}

//WriteTemplate writes a given template to the response writer with content type set as html
func WriteTemplate(appCtx *config.AppContext, res http.ResponseWriter, t Template, data interface{}) {
	res.WriteHeader(http.StatusOK)
	res.Header().Set("Content-Type", "text/html")
	err := t.T.ExecuteTemplate(res, t.Name, data)
	if err != nil && appCtx != nil {
		//Error while writing the template
		appCtx.Log.Error("Error while rendering the template", err.Error())
	} else if appCtx == nil {
		//Error while writing the template
		log.Error("No app context found")
	}
}

//WriteErrorTemplate writes a given error template to the response writer with content type set as html
func WriteErrorTemplate(appCtx *config.AppContext, res http.ResponseWriter, t Template, data interface{}, code int) {
	res.WriteHeader(code)
	res.Header().Set("Content-Type", "text/html")
	err := t.T.ExecuteTemplate(res, t.Name, data)
	if err != nil && appCtx != nil {
		//Error while writing the template
		appCtx.Log.Error("Error while rendering the template", err.Error())
	} else if err != nil && appCtx == nil {
		//Error while writing the template
		log.Error("Error while rendering the template", err.Error())
	}
}
