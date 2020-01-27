// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/hashicorp/consul/api"
)

/*
 * This file contains the discovery service init
 */

//AuthServiceID is the auth service id to be used with the discovery service
var AuthServiceID = "Brain-Auth-Service"

//AuthServiceRPCID is the rpc auth service id tpo be used with the discovery service
var AuthServiceRPCID = "Brain-Auth-Service-RPC"

func init() {
	/*
	 * We will communicate with the consul client
	 * Will prepare the service instance for the http and rpc service
	 * Then will register the application with consul
	 */
	//Registering the db with the discovery api
	// Get a new client
	log.Println("Going to register with the discovery service")
	dConfig := api.DefaultConfig()
	dConfig.Address = DiscoveryURL
	dConfig.Token = DiscoveryToken
	client, err := api.NewClient(dConfig)
	if err != nil {
		log.Fatal("Error while initing the discovery service client", err.Error())
		return
	}

	//service instances for the http service
	log.Println("Connected with discovery service")
	appInstance := &api.AgentServiceRegistration{
		Name: AuthServiceID,
		Port: IntPort,
		Tags: []string{AuthServiceID},
	}

	//registering the service with the agent
	log.Println("Going to register the web service with the discovery service")
	err = client.Agent().ServiceRegister(appInstance)
	if err != nil {
		log.Fatal("Error while registering the web service with the discovery agent", err.Error())
	}

	//service instance for rpc service
	rpcInstance := &api.AgentServiceRegistration{
		Name: AuthServiceRPCID,
		Port: RPCIntPort,
		Tags: []string{AuthServiceRPCID},
		Meta: map[string]string{"RPCService": "yes"},
	}
	log.Println("Going to register the rpc service with the discovery service")
	err = client.Agent().ServiceRegister(rpcInstance)
	if err != nil {
		log.Fatal("Error while registering the rpc service with the discovery agent", err.Error())
	}

	log.Println("Successfully registered with the discovery service")
}

//StartRPC service will start the rpc service. It helps the services to communicate between each other
func StartRPC() {
	/*
	 * Will register the user auth rpc with rpc package
	 * We will listen to the http with rpc of auth module
	 * Then we will start listening to the rpc port
	 */
	//Registering the auth model with the rpc package
	rpc.Register(new(RPCAuth))

	//registering the handler with http
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+RPCPort)
	if e != nil {
		log.Fatal("Error while listening to the rpc port", e.Error())
	}
	go http.Serve(l, nil)
}
