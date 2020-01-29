// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"net/rpc"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

/*
 * This file contains the user model and its defnitions
 */

//User is used by the application to authenticate the users
type User struct {
	//UID is the unique id of the user
	UID uuid.UUID
	//AccessToken is the token with which the user is authenticated
	AccessToken string
	//AuthAgent has the name of the agent who authenticated the user. Eg. GOOGLE
	AuthAgent string
	//Email is the name of the user
	Email string
}

var users = make(map[string]*User)

//GetUser returns the user given the user id
func GetUser(id string) *User {
	return users[id]
}

//NewUser returns a new user
func NewUser(id uuid.UUID) *User {
	user := &User{UID: id}
	users[user.UID.String()] = user
	return user
}

var authenticatedUsers = &AuthenticatedUsers{users: make(map[string]User)}

//AuthenticatedUsers stores the users that are authenticated in the system
type AuthenticatedUsers struct {
	users map[string]User
	lock  sync.Mutex
}

//GetAutenticatedUser will return the autenticated user for a given accesstoken
//It will return the user if existing. ok parameter will be false if the user is not
//authenticated for a given access token
func GetAutenticatedUser(accessToken string) (user User, ok bool) {
	authenticatedUsers.lock.Lock()
	user, ok = authenticatedUsers.users[accessToken]
	authenticatedUsers.lock.Unlock()
	return
}

//SetAuthenticatedUser will set a user as an authenticated user
func (a *AuthenticatedUsers) SetAuthenticatedUser(user User) {
	a.lock.Lock()
	a.users[user.AccessToken] = user
	a.lock.Unlock()
}

//DeleteAuthenticatedUser will delete an user as an authenticated user
func (a *AuthenticatedUsers) DeleteAuthenticatedUser(user User) {
	a.lock.Lock()
	delete(a.users, user.AccessToken)
	a.lock.Unlock()
}

//RPCAuth has the handler for the auth rpc api of the application
type RPCAuth struct{}

//Authenticate will inform the service that the provided user is authenticated
func (r *RPCAuth) Authenticate(u User, ok *bool) error {
	authenticatedUsers.SetAuthenticatedUser(u)
	*ok = true
	return nil
}

//Unauthenticate will invalide the user auth with the given access token
func (r *RPCAuth) Unauthenticate(u User, ok *bool) error {
	authenticatedUsers.DeleteAuthenticatedUser(u)
	*ok = true
	return nil
}

//GetAllAutheticatedUsers will return the list of all the authenticated users
func (r *RPCAuth) GetAllAutheticatedUsers(ok bool, users *map[string]User) error {
	/*
	 * We will iterate through the authenticated users list
	 * and addd each one to the result list
	 */
	*users = authenticatedUsers.users
	return nil
}

//InformAuth will info other services that the user has been authenticated
func (u User) InformAuth(appCtx AppContext, loggedIn bool) {
	/*
	 * We will communicate with the consul client
	 * Then we will get all the services.
	 * We will then use the auth rpc call to authenticate them.
	 */
	//Registering the db with the discovery api
	// Get a new client
	dConfig := api.DefaultConfig()
	dConfig.Address = DiscoveryURL
	dConfig.Token = DiscoveryToken
	client, err := api.NewClient(dConfig)
	if err != nil {
		appCtx.Log.Error("Error while initing the discovery service client", err.Error())
		return
	}

	//getting the services list
	services, err := client.Agent().Services()
	if err != nil {
		appCtx.Log.Error("Error while getting the list fo services registered with the application")
		//We needn't panic since it's not affecting any service
		return
	}

	//going to use rpc call to authenticate the user
	for _, v := range services {
		if _, ok := v.Meta["RPCService"]; ok && v.ID != AuthServiceID {
			u.rpcAuth(appCtx, v, loggedIn)
		}
	}
}

//rpcAuth will do a rpc call to the given service for the provided user
func (u User) rpcAuth(appCtx AppContext, service *api.AgentService, loggedIn bool) {
	/*
	* Will dial the rpc server
	* On defer will close the client
	* if no error will make the rpc call
	 */
	//dailing the rpc server
	client, err := rpc.DialHTTP("tcp", service.Address+":"+strconv.Itoa(service.Port))
	if err != nil {
		//error whiole connecting to the client
		appCtx.Log.Error("Error while connecting to the rpc client for auth of", service)
		appCtx.Log.Error(err.Error())
		return
	}

	//close the client on defer
	defer func() {
		client.Close()
	}()

	//make rpc call
	var ok = false
	procedure := "RPCAuth.Authenticate"
	if !loggedIn {
		procedure = "RPCAuth.Unauthenticate"
	}
	err = client.Call(procedure, u, &ok)
	if err != nil {
		//error while making the rpc
		appCtx.Log.Error("Error while informing the user authentication to the service", service)
		appCtx.Log.Error(err.Error())
		return
	}
}

//InitAuthState will init the authentication state of the microservice.
//It will fetch all the authentitcated users from the auth service service
func InitAuthState(l Logger) error {
	/*
	 * We will initialize the client required for getting the consul service
	 * We will get all the services that are registered with the consul
	 * We will iterate through the services to find the brain auth service we will initiate the rpc to get all the users
	 * If we couldn't find the service it's fine may be the auth service is not yet up
	 * We will create a rpc client
	 * Then we will call the get all authenticated users of RPCAuth
	 */
	//initing the client
	dConfig := api.DefaultConfig()
	dConfig.Address = DiscoveryURL
	dConfig.Token = DiscoveryToken
	client, err := api.NewClient(dConfig)
	if err != nil {
		//error while initializing the client
		l.Error("Error while initializing the client for initing the auth state")
		return err
	}

	//getting all the services
	services, err := client.Agent().Services()
	if err != nil {
		//Error while getting all the services list
		l.Error("Error while getting the list of services registered while fetching the authenticated users list")
		return err
	}

	//iterating through the services to find the list of services
	var service *api.AgentService
	for _, v := range services {
		if _, ok := v.Meta["RPCService"]; ok && v.ID == AuthServiceRPCID {
			service = v
		}
	}

	//checking whether we could find a service
	if service == nil {
		return nil
	}

	//creating a rpc client to do a rpc call
	rClient, errC := rpc.DialHTTP("tcp", service.Address+":"+strconv.Itoa(service.Port))
	if errC != nil {
		l.Error("Error while getting the rpc client for fethcing the list of authenticated users", service.Address+":"+strconv.Itoa(service.Port))
		return errC
	}

	//closing the client on defer
	defer func() {
		rClient.Close()
	}()

	//making the rpc call
	err = rClient.Call("RPCAuth.GetAllAutheticatedUsers", true, &authenticatedUsers.users)
	if err != nil {
		//Error while getting the list
		l.Error("Error while fetching the list of authenticated users from auth service")
		return err
	}

	return nil
}

//UserInfo is the model used for storing the profile info of the user
type UserInfo struct {
	//Email of the user
	Email string `db:"email"`
	//Name of the user
	Name string `db:"name"`
	//Picture of the user
	Picture string `db:"picture"`
	//Registered indicates whether the user has registered with the application
	Registered bool `db:"registered"`
	//Subscribed indicates that the user is subscribed to the platform newsletter
	Subscribed bool `db:"subscribed"`
	//LastSeen is the last activity by the user
	LastSeen time.Time `db:"last_seen"`
	//LastSynced is the last time the profile was updated from the oauth agent
	LastSynced time.Time `db:"last_synced"`
}

//Get returns the userinfo model from the database
//If doesn't exist in the db, the method will return nil
func (u UserInfo) Get(ctx AppContext) (result *UserInfo) {
	results := []UserInfo{}
	ctx.Db.Where("email = ?", u.Email).Find(&results)
	if len(results) != 0 {
		result = &results[0]
	}
	return
}

//Insert inserts the user info record to the database
func (u UserInfo) Insert(ctx AppContext) error {
	return ctx.Db.Create(&u).Error
}

//Update updates the userinfo model based on the email
func (u UserInfo) Update(ctx AppContext) error {
	return ctx.Db.Model(&u).Where("email = ?", u.Email).Updates(map[string]interface{}{
		"name":        u.Name,
		"picture":     u.Picture,
		"registered":  u.Registered,
		"last_seen":   u.LastSeen,
		"last_synced": u.LastSynced,
	}).Error
}
