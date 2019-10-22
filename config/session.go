// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

/* this file contains the model definition of http user session */

//Session denotes an existing user session
type Session struct {
	//ID is the id of te session
	ID string
	//Authenticated denotes whether the session is authenticated or not
	Authenticated bool
	//User with which the app context is associated with
	User *User
}
