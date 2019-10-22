// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Package oauth contains the oauth utilities for the platform
package oauth

import (
	"context"
	"errors"

	"github.com/cuttle-ai/auth-service/config"
)

/*
 * This file contains the defnition of the oauth Agent interface
 */

const (
	//GOOGLE is google oauth agent string
	GOOGLE = "GOOGLE"
)

//Agent is the oauth agent interface to be implmented. OAuth agents are google, facebook etc.
type Agent interface {
	//Info returns the user info for a given app context
	Info(context.Context, *config.AppContext) (*config.UserInfo, error)
	//Name is the name of the auth agent
	Name() string
}

//Info returns the userinfo of a given user from the oauth agent
func Info(ctx context.Context, appCtx *config.AppContext, agent Agent) (*config.UserInfo, error) {
	/*
	 * Will check whether the user/oauth is nil or not
	 * Based on the oauth agent we will fetch the user info
	 */
	//nil check for the user
	if appCtx.Session.User == nil {
		return nil, errors.New("No user found in the session")
	}
	if agent == nil {
		return nil, errors.New("No auth agent found for the user")
	}

	//returning the information based on the auth agent
	return agent.Info(ctx, appCtx)
}
