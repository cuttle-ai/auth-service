// Copyright 2019 Cuttle.ai. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"log"
	"os"

	"github.com/cuttle-ai/brain/appctx"
	bLog "github.com/cuttle-ai/brain/log"
	"github.com/cuttle-ai/db-toolkit/datastores/services"
	"github.com/cuttle-ai/go-sdk/services/datastores"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

/* This file contains the definition of AppContext */

const (
	//DbHost is the environment variable storing the database access url
	DbHost = "DB_HOST"
	//DbPort is the environment variable storing the database access port
	DbPort = "DB_PORT"
	//DbDatabaseName is the environment variable storing the database name
	DbDatabaseName = "DB_DATABASE_NAME"
	//DbUsername is the environment variable storing the database username
	DbUsername = "DB_USERNAME"
	//DbPassword is the environment variable storing the database password
	DbPassword = "DB_PASSWORD"
	//EnabledDB is the environment variable stating whether the db is enabled or not
	EnabledDB = "ENABLE_DB"
	//DevelopmentDatastoreUser is the username that has ssh access to the development datastore
	DevelopmentDatastoreUser = "DEVELOPMENT_DATSTORE_USERID"
)

//MasterAppDetails details has the master app details
var MasterAppDetails *AppInfo

//DbConfig is the database configuration to connect to it
type DbConfig struct {
	//Host to be used to connect to the database
	Host string
	//Port with which the database can be accessed
	Port string
	//Database to connect
	Database string
	//Username to access the connection
	Username string
	//Password to access the connection
	Password string
}

//NewDbConfig will read the db config from the os environment variables and set it in the config
func NewDbConfig() *DbConfig {
	dbC := &DbConfig{
		Host:     os.Getenv(DbHost),
		Port:     os.Getenv(DbPort),
		Database: os.Getenv(DbDatabaseName),
		Username: os.Getenv(DbUsername),
		Password: os.Getenv(DbPassword),
	}
	return dbC
}

//Connect will connect the database. Will return an error if anything comes up else nil
func (d DbConfig) Connect() (*gorm.DB, error) {
	/*
	 * We will build the connection string
	 * Then will connect to the database
	 */
	cStr := fmt.Sprintf("host=%s port=%s dbname=%s  user=%s password=%s sslmode=disable",
		d.Host, d.Port, d.Database, d.Username, d.Password)

	return gorm.Open("postgres", cStr)
}

//AppContext contains the
type AppContext struct {
	//Db is the database connection
	Db *gorm.DB
	//Log for logging purposes
	Log Logger
	//Session is the session associated with the request
	Session Session
}

var rootAppContext *AppContext

func init() {
	/*
	 * We will initialize the context
	 * We will connect to the database
	 * Then we will get all the authenticated apps from the database
	 * Then load them up into the authentication map
	 */
	//initializing the context
	rootAppContext = &AppContext{}

	//connecting the database
	log.Println("Going to connect with Database")
	err := rootAppContext.ConnectToDB()
	if err != nil {
		log.Fatal("Error while creating the root app context. Connecting to DB failed. ", err)
	}

	//getting all the authenticated apps from the database
	apps := GetAllApps(*rootAppContext)
	appsMap := map[string]App{}
	for _, v := range apps {
		appsMap[v.AccessToken] = v.ToApp()
	}

	//storing the authenticated apps in the authentication map
	authenticatedUsers.apps = appsMap
}

//AddAsSuperAdmin will add the given user as a super admin
func AddAsSuperAdmin(ctx *AppContext, userID uint, email string) error {
	/*
	 * We will add the user as super admin
	 * Then will create the master token for the same
	 * Then we will inform the auth
	 * Then we will initialize the datastore
	 */
	iu := UserInfo{Model: gorm.Model{ID: userID}}
	err := iu.AddAsSuperAdmin(*ctx)
	if err != nil {
		return err
	}

	err = initializeAppToken(ctx, userID, email)
	if err != nil {
		return err
	}

	user := MasterAppDetails.ToApp().ToUser()
	user.InformAuth(*ctx, true)

	err = initializeDatastore(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func initializeDatastore(ctx *AppContext, userID uint) error {
	/*
	 * If the environment is prod we will return
	 * First we will create the data store in db
	 * Then will register it with the datastore service
	 */
	if PRODUCTION != 0 {
		return nil
	}

	initDatastoreName := "development-datastore"
	ctx.Log.Info("creating the init datastore since in non-prod mode", initDatastoreName)
	err := CreateInitDatastore(*ctx, initDatastoreName)
	if err != nil {
		ctx.Log.Error("error while creating the init datastore", initDatastoreName)
		return err
	}

	_, err = datastores.CreateDatastore(appctx.WithAccessToken(ctx, MasterAppDetails.AccessToken), services.Service{
		URL:           os.Getenv(DbHost),
		Port:          os.Getenv(DbPort),
		Username:      os.Getenv(DbUsername),
		Password:      os.Getenv(DbPassword),
		Name:          initDatastoreName,
		Group:         "development",
		Datasets:      0,
		DatastoreType: services.POSTGRES,
		DataDirectory: os.Getenv(DevelopmentDatastoreUser) + "@" + os.Getenv(DbHost) + ":/home/",
	})
	if err != nil {
		//error while registering the development datastores with datastores service
		ctx.Log.Error("error while registering the development datastores with datastores service", initDatastoreName)
		return err
	}
	return nil
}

//initializeAppToken will initialize the master app token
func initializeAppToken(ctx *AppContext, userID uint, email string) error {
	/*
	 * We can get the master app from db
	 * If not created, we will create one
	 * Then store the token in the config
	 */
	ap, err := GetMasterApp(*ctx)
	if err == nil {
		//app is available
		ap = MasterAppDetails
		return nil
	}

	//not created. Need to create one
	//this adds the first user came to the platform as the master user. Generally it should be the admin
	ap = &AppInfo{
		UserID:      userID,
		UID:         uuid.New(),
		Email:       email,
		AccessToken: uuid.New().String(),
		Description: "Master Cuttle App",
		Name:        "Cuttle Master",
		IsMasterApp: true,
	}
	err = ap.Insert(*ctx)
	if err != nil {
		//error while inserting the master app details
		return err
	}

	//putting in the details
	MasterAppDetails = ap
	return nil
}

//NewAppContext returns an initlized app context
func NewAppContext(l Logger) *AppContext {
	return &AppContext{Log: l, Db: rootAppContext.Db}
}

//ConnectToDB connects the database and updates the Db property of the context as new connection
//If any error happens in between , it will be returned and connection won't be set in the context
func (a *AppContext) ConnectToDB() error {
	/*
	 * We will enable db only if the enable db env is true
	 * We will get the db config
	 * Connect to it
	 * If no error then set the database connection
	 */
	if os.Getenv(EnabledDB) != "true" {
		return nil
	}
	c := NewDbConfig()
	d, err := c.Connect()
	if err == nil {
		a.Db = d
	}
	a.Db.AutoMigrate(&UserInfo{})
	a.Db.AutoMigrate(&AppInfo{})
	return err
}

//Logger returns the logger of the app context
func (a AppContext) Logger() bLog.Log {
	return a.Log
}

//AccessToken of the app
func (a AppContext) AccessToken() string {
	return a.Session.ID
}

//DiscoveryAddress of thedisocvery service
func (a AppContext) DiscoveryAddress() string {
	return DiscoveryURL
}

//DiscoveryToken of the discovery service
func (a AppContext) DiscoveryToken() string {
	return DiscoveryToken
}
