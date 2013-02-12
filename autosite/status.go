/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite

import (
	"appengine/datastore"
	"reflect"
	"strings"
	"time"
)

type Status struct {
	Kind string
	OriginalId int64
	Heading string
	Content string
	Link string
	Created time.Time
}

func (u *Status) Type() string {
	return strings.ToLower(strings.Replace(reflect.TypeOf(u).String(), "*autosite.", "", 1))
}

func (u *Status) Save() (string, error) {
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, u.Type(), nil), u)
	return key.String(), err
}
