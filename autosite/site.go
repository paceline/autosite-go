/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/
package autosite

import (
	"appengine"
    "appengine/datastore"
    "net/http"
    "reflect"
    "strings"
)

// Set up data structure
type Site struct {
	SiteTitle	string
	HomepageTitle	string
	Footer	string
	WebRoot	string
	BitlyUserName	string
	BitlyAccessKey	string
	TrackerCode	string
}

// Return own type as String
func (s *Site) Type() string {
	return strings.ToLower(strings.Replace(reflect.TypeOf(s).String(), "*autosite.", "", 1))
}

// Set and save attrs
func (s *Site) Save(r *http.Request) (string, error) {
	c := appengine.NewContext(r)
	s.SiteTitle = r.FormValue(FormKey(s,"SiteTitle"))
	s.HomepageTitle = r.FormValue(FormKey(s,"HomepageTitle"))
	s.Footer = r.FormValue(FormKey(s,"Footer"))
	s.WebRoot = r.FormValue(FormKey(s,"WebRoot"))
	s.BitlyUserName = r.FormValue(FormKey(s,"BitlyUserName"))
	s.BitlyAccessKey = r.FormValue(FormKey(s,"BitlyAccessKey"))
	s.TrackerCode = r.FormValue(FormKey(s,"TrackerCode"))
	key, err := datastore.Put(c, datastore.NewKey(c, s.Type(), "MySite", 0, nil), s)
	return key.String(), err
}

// Load from datastore
func (s *Site) Load(r *http.Request) error {
	c := appengine.NewContext(r)
	err := datastore.Get(c, datastore.NewKey(c, s.Type(), "MySite", 0, nil), s)
	return err
}
