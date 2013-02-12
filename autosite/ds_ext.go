/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite

import (
    "appengine/datastore"
    "github.com/garyburd/go-oauth/oauth"
    "net/http"
)


/*
 * Interface definition and shared functions 
 */

type Model interface {
	Type() string
}

// Generic get by function
func Get(m Model) string {
	q := datastore.NewQuery(m.Type()).Limit(1)
	for t := q.Run(c); ; {
		key, err := t.Next(m)
        if err != nil {
			session.AddFlash("An error occured while loading: %s", err.Error())
			break
        }
        return key.Encode()
	}
	return ""
}

// Generic get by function
func GetByName(m Model, name string) string {
	q := datastore.NewQuery(m.Type()).Filter("Name =", name).Limit(1)
	for t := q.Run(c); ; {
		key, err := t.Next(m)
        if err != nil {
			session.AddFlash("An error occured while loading: %s", err.Error())
			break
        }
        return key.Encode()
	}
	return ""
}

// Generic count function
func Count(m Model) (int, error) {
	q := datastore.NewQuery(m.Type())
	return q.Count(c)
}

// Initialize model from form
func Build(m Model, r *http.Request) {
	r.ParseForm()
	decoder.Decode(m, r.Form)
}

// Save new model (random key)
func Save(m Model) string {
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, m.Type(), nil), m)
	if err != nil {
		session.AddFlash("An error occured while saving: %s", err.Error())
		return ""
    }
    session.AddFlash(m.Type() + " has been saved successfully")
	return key.Encode()
}

// Save new model (pre-defined key)
func Update(m Model, k string) string {
	key, _ := datastore.DecodeKey(k)
	_, err := datastore.Put(c, key, m)
	if err != nil {
		session.AddFlash("An error occured while saving: %s", err.Error())
    }
    session.AddFlash(m.Type() + " has been saved successfully")
	return key.Encode()
}

// Generic delete function
func Delete(k string) {
	key, _ := datastore.DecodeKey(k)
	err := datastore.Delete(c, key)
	if err != nil {
		session.AddFlash("An error occured while deleting: %s", err.Error())
    }
}


/*
 * Site struct for storing main site info
 */

type Site struct {
	SiteTitle	string
	HomepageTitle	string
	Footer	string
	WebRoot	string
	TrackerCode	[]byte
}

// Return own type as String
func (s *Site) Type() string {
	return "Site"
}

// Return tracker code as string
func (s *Site) TrackerCodeString() string {
	return string(s.TrackerCode)
}


/*
 * Page struct and non-shared methods 
 */
 
type Page struct {
	Position	int
	Title	string
	Body	[]byte
	Name	string
	Published	bool
}

// Return own name
func (p *Page) Type() string {
	return "Page"
}

// Return own name
func (p *Page) BodyString() string {
	return string(p.Body)
}

// Sort pages
func SortPages(slugs []string) {
	pages := make([]Page, 0)
	q := datastore.NewQuery("Page")
	keys, _ := q.GetAll(c, &pages)
	for i := 0; i < len(keys); i++ {
		pages[i].Position = LookFor(slugs, pages[i].Name) + 1
	}
	datastore.PutMulti(c, keys, pages)
}


/*
 * Account struct for storing credentials
 */

type Account struct {
	Name string
	ConsumerKey string
	ConsumerSecret string
	Token string
	Secret string
}

func (a *Account) Type() string {
	return "Account"
}

func (a *Account) Creds() oauth.Credentials {
	return oauth.Credentials{Token: a.Token, Secret: a.Secret}
}
