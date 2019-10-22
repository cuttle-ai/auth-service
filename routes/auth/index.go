// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"html/template"

	"github.com/cuttle-ai/auth-service/config"
	"github.com/cuttle-ai/auth-service/routes/response"
)

/*
 * This file contains the template for the html response of index page
 */

var headerText = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Cuttle.ai</title>
</head>
<body>`

var footerText = `
</body>
</html>`

var indexTemplateString = headerText + `<span>{{.Email}}</span> <a href="/logout">Logout</a>` + footerText

func indexPage(appCtx *config.AppContext) response.Template {
	tem, err := template.New("index-page").Parse(indexTemplateString)
	if err != nil {
		appCtx.Log.Error("Error while initializing the index page template in routes/auth/index", err.Error())
	}
	return response.Template{T: tem, Name: "index-page"}
}

var indexRedirectTemplateString = headerText +
	`<h1>You will be redirected in <span id="timer">3</span>s</h1>` +
	`<script>
let timer = 3;
function redirect() {
	timer--;
	document.getElementById('timer').innerText = timer;
	if(timer === 0) {
		location.href = '{{.}}';
	}
}
setInterval(redirect, 1000);
</script>` +
	footerText

func indexRedirectPage(appCtx *config.AppContext) response.Template {
	tem, err := template.New("index-redirect-page").Parse(indexRedirectTemplateString)
	if err != nil {
		appCtx.Log.Error("Error while initializing the index redirect page template in routes/auth/index", err.Error())
	}
	return response.Template{T: tem, Name: "index-redirect-page"}
}

var indexErrorTemplateString = headerText + `
<span>Authenticate your self
{{range $key, $value := .}} 
<a href="{{$value}}">{{$key}}</a> {{end}}
</span>` + footerText

func indexErrorPage(appCtx *config.AppContext) response.Template {
	tem, err := template.New("index-error-page").Parse(indexErrorTemplateString)
	if err != nil {
		appCtx.Log.Error("Error while initializing the index error page template in routes/auth/index", err.Error())
	}
	return response.Template{T: tem, Name: "index-error-page"}
}
