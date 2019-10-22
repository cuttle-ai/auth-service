// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"log"

	"github.com/hashicorp/consul/api"
)

/*
 * This file contains the discovery service init
 */

//AuthServiceID is the auth service id to be used with the discovery service
var AuthServiceID = "Brain-Auth-Service"

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
		Tags: []string{"Go", AuthServiceID},
	}

	//registering the service with the agent
	log.Println("Going to register with the discovery service")
	err = client.Agent().ServiceRegister(appInstance)
	if err != nil {
		log.Fatal("Error while registering with the discovery agent", err.Error())
	}

	log.Println("Successfully registered with the discovery service")
}
