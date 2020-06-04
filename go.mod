module github.com/cuttle-ai/auth-service

go 1.13

replace github.com/cuttle-ai/configs => ../configs/

require (
	github.com/cuttle-ai/configs v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.1.1
	github.com/hashicorp/consul/api v1.2.0
	github.com/inconshreveable/log15 v0.0.0-20180818164646-67afb5ed74ec // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/xeonx/timeago v1.0.0-rc4 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/fsnotify/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/stack.v0 v0.0.0-20141108040640-9b43fcefddd0 // indirect
)
