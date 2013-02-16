# autosite-go

### Introduction
Package autosite provides a simple infrastructure for running a
personal website (off of the Google App Engine)

### TODOs
* Support more OAuth providers (planned are: Facebook, Github, LinkedIn, XING)
* Add RSS/Atom feed
* Validations and proper error handling
* Move all datastore code to ds_ext.go
* More documentation

### Required libraries
* Gorilla web toolkit (github.com/gorilla/mux, github.com/gorilla/schema, github.com/gorilla/sessions)
* go-oauth (github.com/garyburd/go-oauth/oauth)
* goauth2 (code.google.com/p/goauth2/oauth)

### Credits
Created by Ulf Möhring <ulf@moehring.me>
