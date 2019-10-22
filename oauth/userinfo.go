// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package oauth

/*
 * This file contains the user info map required by the oauth
 */

//UserInfoMap contains the string of the keywords to be used to retrieve the
//user info from the api respose of the auth agent like Google, Facebook etc.
type UserInfoMap struct {
	Email   string //Email key of the user info model
	Name    string //Name key of the user info model
	Picture string //Picture url key of the user info model
}
