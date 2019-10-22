// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package response

/*
 * This file contains the error codes required for sending out the response
 */

//ERROR CODES
const (
	//ErrorCodeNone denotes there is no error
	ErrorCodeNone = 0
	//ErrorCodeSessionExpired denotes the session has been expired
	ErrorCodeSessionExpired = 1
	//ErrorCodeInvalidParams denotes that the api parameters are invalid
	ErrorCodeInvalidParams = 2
)

//ErrorCodes has the map of error code mapped to the error messages
var ErrorCodes map[int]string = map[int]string{
	ErrorCodeNone:           "No Error",
	ErrorCodeSessionExpired: "Session has been expired",
	ErrorCodeInvalidParams:  "The following parameters are invalid",
}
