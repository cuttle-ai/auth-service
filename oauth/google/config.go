// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Package google has the implmentations for google oauth
// and google user profile fetch
package google

/*
 * This file contains the google oauth utils
 */

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/cuttle-ai/auth-service/config"
	"github.com/cuttle-ai/auth-service/oauth"
	"github.com/revel/revel"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

//Config is the google oauth2 config
var Config *oauth2.Config

//UserInfoMap has the user info map required for the google use profile
var UserInfoMap oauth.UserInfoMap

//UserInfoURL is the url to hit to get the user info from google
var UserInfoURL string

//UserInfoTimeout is the timeout for the fetching the user info from google auth agent
var UserInfoTimeout = time.Duration(200 * time.Millisecond)

const (
	//RedirectURLKey is the key storing for the google oauth redirect url
	RedirectURLKey = "OAUTH2_GOOGLE_REDIRECT_URL"
	//ClientIDKey is the key storing for the google oauth client id
	ClientIDKey = "OAUTH2_GOOGLE_CLIENT_ID"
	//ClientSecretKey is the key storing for the google oauth client secret
	ClientSecretKey = "OAUTH2_GOOGLE_CLIENT_SECRET"
	//ProfileScopeKey is the key storing for the google user profile scope
	ProfileScopeKey = "OAUTH2_GOOGLE_USER_PROFILE_SCOPE"
	//EmailScopeKey is the key storing for the google user email scope
	EmailScopeKey = "OAUTH2_GOOGLE_USER_EMAIL_SCOPE"
	//UserInfoURLKey is the key storing the url which gives google user info
	UserInfoURLKey = "OAUTH2_GOOGLE_USER_INFO_URL"
	//UserInfoNameKey is the key storing the name key in the google user info
	UserInfoNameKey = "OAUTH2_GOOGLE_USER_INFO_NAME"
	//UserInfoEmailKey is the key storing the email key in the google user info
	UserInfoEmailKey = "OAUTH2_GOOGLE_USER_INFO_EMAIL"
	//UserInfoPictureKey is the key storing the picture key in the google user info
	UserInfoPictureKey = "OAUTH2_GOOGLE_USER_INFO_PICTURE"
	//UserInfoTimeoutKey is the key storing the user info google fetch timeout
	UserInfoTimeoutKey = "OAUTH2_GOOGLE_USER_INFO_TIMEOUT"
)

//LoadConfig will load the config required for the google oauth
func LoadConfig() error {
	/*
	 * We will set the redirect url
	 * Then client id
	 * Then client secret
	 * Then user profile scope
	 * Then user email scope
	 * Then will initialize the google config
	 * Then we will set the configs for user info
	 * 		URL
	 * 		Name
	 *		Email
	 *		Picture
	 */
	//setting the oauth2 config
	redirectURL := os.Getenv(RedirectURLKey)
	if len(redirectURL) == 0 {
		return errors.New("Google OAuth2 redirect url not found")
	}
	clientID := os.Getenv(ClientIDKey)
	if len(clientID) == 0 {
		return errors.New("Google OAuth2 client id not found")
	}
	clientsecret := os.Getenv(ClientSecretKey)
	if len(clientsecret) == 0 {
		return errors.New("Google OAuth2 client secret not found")
	}
	profileScope := os.Getenv(ProfileScopeKey)
	if len(profileScope) == 0 {
		return errors.New("Google OAuth2 user profile scope not found")
	}
	emailScope := os.Getenv(EmailScopeKey)
	if len(emailScope) == 0 {
		return errors.New("Google OAuth2 user email scope not found")
	}

	// //setting the user info configs
	userInfoURL := os.Getenv(UserInfoURLKey)
	if len(userInfoURL) == 0 {
		return errors.New("Google OAuth2 user info url not found")
	}
	userInfoName := os.Getenv(UserInfoNameKey)
	if len(userInfoName) == 0 {
		return errors.New("Google OAuth2 user info name key not found")
	}
	userInfoEmail := os.Getenv(UserInfoEmailKey)
	if len(userInfoEmail) == 0 {
		return errors.New("Google OAuth2 user info email key not found")
	}
	userInfoPicture := os.Getenv(UserInfoPictureKey)
	if len(userInfoPicture) == 0 {
		return errors.New("Google OAuth2 user info picture key not found")
	}
	userInfoTimeout := os.Getenv(UserInfoTimeoutKey)
	if len(userInfoTimeout) != 0 {
		//if successful convert timeout
		if t, err := strconv.ParseInt(os.Getenv(UserInfoTimeoutKey), 10, 64); err == nil {
			UserInfoTimeout = time.Duration(t * int64(time.Millisecond))
		}
	}

	Config = &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientsecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			profileScope,
			emailScope,
		},
	}

	UserInfoMap = oauth.UserInfoMap{
		Name:    userInfoName,
		Email:   userInfoEmail,
		Picture: userInfoPicture,
	}

	UserInfoURL = userInfoURL

	return nil
}

func init() {
	err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
}

//Agent is the google oauth agent
type Agent struct{}

//Info returns the returns google user's info
func (a *Agent) Info(ctx context.Context, appCtx *config.AppContext) (*config.UserInfo, error) {
	/*
	 * We will set the context first since we have a network call
	 * We will hit the google api for user info
	 * If any error happens we will return after logging it
	 * Then we will parse the api response
	 * Then will set the properties based on the google user info api's response
	 * We will set the user model email with the info email
	 */
	//setting the conext since we have network call
	u := appCtx.Session.User
	newCtx, cancel := context.WithTimeout(ctx, UserInfoTimeout)
	defer cancel()

	//hitting the google apis
	req, err := http.NewRequestWithContext(newCtx, http.MethodGet, UserInfoURL+url.QueryEscape(u.AccessToken), nil)
	if err != nil {
		//handling the error while setting the request with context
		appCtx.Log.Error("Error while setting the request to fetch the userinfo from google")
		return nil, err
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		//handling the error while getting the info from the google
		appCtx.Log.Error("Error while getting the userinfo from the google")
		appCtx.Log.Error("Status code is ", req.Response.StatusCode)
		return nil, err
	}
	defer res.Body.Close()

	//parsing the api response
	ui := map[string]interface{}{}
	//handling the error while parsing the api response
	if err := json.NewDecoder(res.Body).Decode(&ui); err != nil {
		appCtx.Log.Error("Error while parsing the userinfo api response from google")
		return nil, err
	}

	//setting the info from the api response in the userinfo model
	info := &config.UserInfo{}
	//setting the name
	v, ok := ui[UserInfoMap.Name]
	if !ok {
		revel.AppLog.Error("Error while getting the user's name from the google user info api")
		return nil, errors.New("Key not found " + UserInfoMap.Name)
	}
	info.Name = v.(string)
	//setting the email
	v, ok = ui[UserInfoMap.Email]
	if !ok {
		revel.AppLog.Error("Error while getting the user's email from the google user info api")
		return nil, errors.New("Key not found " + UserInfoMap.Email)
	}
	info.Email = v.(string)
	//setting the picture
	v, ok = ui[UserInfoMap.Picture]
	if !ok {
		revel.AppLog.Error("Error while getting the user's pciture from the google user info api")
		return nil, errors.New("Key not found " + UserInfoMap.Picture)
	}
	info.Picture = v.(string)

	//setting the email of the user with info email
	u.Email = info.Email
	return info, nil
}

//Name returns the google string as the name of the auth agent
func (a *Agent) Name() string {
	return oauth.GOOGLE
}
