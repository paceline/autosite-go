# autosite-go

### Introduction
Package autosite provides a simple infrastructure for running a
personal website (off of the Google App Engine)

### TODOs
* Support more OAuth providers (planned are: Facebook, LinkedIn, XING)
* Add RSS/Atom feed
* Validations and proper error handling
* Move all datastore code to ds_ext.go
* More documentation

### Required libraries
* Gorilla web toolkit ([mux](http://github.com/gorilla/mux), [schema](github.com/gorilla/schema), and [sessions](github.com/gorilla/sessions))
* [go-oauth](github.com/garyburd/go-oauth/oauth)
* [goauth2 (custom)](http://github.com/paceline/goauth2/oauth)

### Credits
Created by Ulf Möhring <ulf@moehring.me>
