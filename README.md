# autosite-go

### Introduction
Package autosite provides a simple infrastructure for running a
personal website (off of the Google App Engine)

### TODOs
* Support more OAuth providers (planned are: LinkedIn, XING)
* Add RSS/Atom feed
* Validations
* Move all datastore code to ds_ext.go
* More documentation

### Required libraries
* Gorilla web toolkit ([mux](http://github.com/gorilla/mux), [schema](http://github.com/gorilla/schema), and [sessions](http://github.com/gorilla/sessions))
* [go-oauth](http://github.com/garyburd/go-oauth/)
* [goauth2 (custom)](http://github.com/paceline/goauth2)

### Credits
Created by Ulf Möhring <ulf@moehring.me>
