// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package routes

import (
	"time"

	"github.com/cuttle-ai/auth-service/config"
	"github.com/cuttle-ai/auth-service/log"
	"github.com/google/uuid"
)

/*
 * this file contains the defintions of the rate limiter.
 * Basically the server cater the no. of requests at a given point of time as per specs.
 * When requests overflows it become very easy to scale if it is tracked.
 */

//RequestType is the type of the AppContext Request
type RequestType int

const (
	//Get is to get an app context
	Get RequestType = 0
	//Finished is to return an app context
	Finished RequestType = 1
	//CleanUp is to clean up the non-returned app context
	CleanUp RequestType = 2
	//SetSession sets the user session info
	SetSession RequestType = 3
)

//AppContextRequest is the request to get, return or try clean up app contexts
type AppContextRequest struct {
	//AppContext is the appcontext being requested
	AppContext *config.AppContext
	//Type is the type of request
	Type RequestType
	//Out is the ouput channel for get requests
	Out chan AppContextRequest
	//Exhausted flag states whether the app context exhausted
	Exhausted bool
	//Session is  the user session
	Session config.Session
}

//AppContextRequestChan is the common channel through which the requests for app context come
var AppContextRequestChan = make(chan AppContextRequest)

//SendRequest is to send request to the channel. When this function used as go routines
//the blocking quenes can be solved
func SendRequest(ch chan AppContextRequest, req AppContextRequest) {
	ch <- req
}

//AppContext is the app context go routine running to
func AppContext(in chan AppContextRequest) {
	/*
	 * We will keep two maps for storing busy requests and free requests
	 * Will make a map for storing the user sessions
	 * First we will generate the id pool and store it in
	 * We will start inifinite loop waiting for the requests
	 */
	//maps for storing the free and used requests
	freeMaps := make([]int, config.MaxRequests)
	usedMaps := make(map[int]time.Time, config.MaxRequests)
	userSession := make(map[string]config.Session)

	//generate the request pool
	for i := 1; i <= config.MaxRequests; i++ {
		freeMaps = append(freeMaps, i)
	}

	//starting the infinite loop waiting for the requests
	for {
		req := <-in
		switch req.Type {
		case Get:
			//If it is a get request we will try to get get a app context from the store
			//first check whether exist free session maps
			if len(freeMaps) == 0 {
				req.Exhausted = true
				go SendRequest(req.Out, req)
				return
			}

			//if exist create an app context
			//get the session into it using the session id provided
			id := freeMaps[0]
			freeMaps = freeMaps[1:]
			usedMaps[id] = time.Now()
			req.AppContext = config.NewAppContext(log.NewLogger(id))
			req.Exhausted = false

			//we will also set the session
			if sess, ok := userSession[req.Session.ID]; ok && len(req.Session.ID) != 0 {
				req.Session = sess
			} else {
				sess = config.Session{ID: uuid.New().String(), Authenticated: false}
				userSession[sess.ID] = sess
				req.Session = sess
			}
			req.AppContext.Session = req.Session
			go SendRequest(req.Out, req)
		case SetSession:
			//we will set the user session in given from
			userSession[req.Session.ID] = req.Session
		case Finished:
			//we will return the rewwuest ids
			delete(usedMaps, req.AppContext.Log.GetID())
			freeMaps = append(freeMaps, req.AppContext.Log.GetID())
		case CleanUp:
			//clean up the timed out requests
			n := time.Now()
			tot := config.RequestRTimeout + config.ResponseTimeout + config.ResponseWTimeout
			toBeAdded := []int{}
			for k, v := range usedMaps {
				if v.Add(tot).Before(n) {
					toBeAdded = append(toBeAdded, k)
					delete(usedMaps, k)
				}
			}
			freeMaps = append(freeMaps, toBeAdded...)
		}
	}
}

//CleanUpCheck is the cleanup check to be used as a go routine which periodically sends cleanup
//requests to the AppContext go routines
func CleanUpCheck(in chan AppContextRequest) {
	/*
	 * We will go into a infinte for loop
	 * Will send the requests of type clean up
	 */
	for {
		time.Sleep(config.RequestCleanUpCheck)
		go SendRequest(in, AppContextRequest{Type: CleanUp})
	}
}

func init() {
	go AppContext(AppContextRequestChan)
	go CleanUpCheck(AppContextRequestChan)
}
