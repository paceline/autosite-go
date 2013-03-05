/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <ulf@moehring.me>
*/

package autosite

import (
	"appengine/datastore"
    "net/http"
    "regexp"
    "strings"
    "text/template"
    "time" 
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
			session.AddFlash("An error occured while loading: " + err.Error())
			break
        }
        return key.Encode()
	}
	return ""
}

// Generic get by function
func GetByKey(m Model, key *datastore.Key) string {
	if err := datastore.Get(c, key, m); err != nil {
        session.AddFlash("An error occured while loading: " + err.Error())
		return ""
    }
	return key.Encode()
}

// Generic get by function
func GetByName(m Model, name string) string {
	q := datastore.NewQuery(m.Type()).Filter("Name =", name).Limit(1)
	for t := q.Run(c); ; {
		key, err := t.Next(m)
        if err != nil {
			session.AddFlash("An error occured while loading: " + err.Error())
			break
        }
        return key.Encode()
	}
	return ""
}

// Convert string to datastore key
func ToKey(k string) *datastore.Key {
	key, _ := datastore.DecodeKey(k)
	return key
}

// Generic count function
func Count(kind string) (int, error) {
	q := datastore.NewQuery(kind)
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
		session.AddFlash("An error occured while saving: " + err.Error())
		return ""
    }
    session.AddFlash(m.Type() + " has been saved successfully")
	return key.Encode()
}

// Save new model (pre-defined key)
func Update(m Model, k string) string {
	key, err := datastore.Put(c, ToKey(k), m)
	if err != nil {
		session.AddFlash("An error occured while saving: " + err.Error())
    }
    session.AddFlash(m.Type() + " has been saved successfully")
	return key.Encode()
}

// Generic delete function
func Delete(k string) {
	err := datastore.Delete(c, ToKey(k))
	if err != nil {
		session.AddFlash("An error occured while deleting: " + err.Error())
    }
}


/*
 * Site struct for storing main site info
 */

type Site struct {
	SiteTitle	string
	HomepageTitle	string
	Footer	string
	TrackerCode	[]byte
	TemplateKey	*datastore.Key
}

// Return own type as String
func (s *Site) Type() string {
	return "Site"
}

// Return tracker code as string
func (s *Site) TrackerCodeString() string {
	return string(s.TrackerCode)
}

// Return own type as String
func (s *Site) Style() string {
	var css Page
	GetByKey(&css, s.TemplateKey)
	return css.Name
}

// Return all templates as map
func (s *Site) Templates() []map[string]string {
	var templates []map[string]string
	q := datastore.NewQuery("Page").Order("Name")
	for t := q.Run(c); ; {
		var template Page
		key, err := t.Next(&template)
		if err == datastore.Done {
			break
        }
		if template.IsTemplate() {
			var selected string
			if s.TemplateKey != nil && key.Encode() == s.TemplateKey.Encode() {
				selected = "selected"
			}
			templates = append(templates, map[string]string{"Key": key.Encode(), "Name": template.Name, "Selected": selected})
		}
	}
	return templates
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

// Check whether page is a (CSS-)template
func (p *Page) IsTemplate() bool {
	cssmatch, _ := regexp.Compile("\\.css$")
	if cssmatch.FindString(p.Name) == "" {
		return false
	}
	return true
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
	RequestUrl string
	AuthUrl string
	AccessUrl string
	Repost bool
	Expires time.Time
}

func (a *Account) Type() string {
	return "Account"
}

func (a *Account) Version() int {
	reqmatch, _ := regexp.Compile("^http")
	if reqmatch.FindString(a.RequestUrl) == "" {
		return 2
	}
	return 1
}

func (a *Account) Verified() bool {
	return len(a.Token) > 0 && (a.Expires.IsZero() || time.Now().Before(a.Expires))
}

func (a *Account) Twitter() bool {
	return a.Name == "twitter"
}

func (a *Account) XING() bool {
	return a.Name == "xing"
}

func (a *Account) LinkedIn() bool {
	return a.Name == "linkedin"
}

func (a *Account) GitHub() bool {
	return a.Name == "github"
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	if extendMethod(r) == "GET" {
		var account Account
		q := datastore.NewQuery("Account")
		for t := q.Run(c); ; {
			_, err := t.Next(&account)
			if err == datastore.Done {
				break
			}
			if account.Verified() {
				switch account.Name {
					case "github":
						account.prepareOAuth2Connection(r)
						account.GetGithubUpdates(r)
					case "linkedin":
						account.prepareOAuth2Connection(r)
						account.GetLinkedInUpdates(r)
					case "twitter":
						account.prepareOAuthConnection(r)
						account.GetTwitterUpdates(r)
				}
			}
		}
		q = datastore.NewQuery("Status").Order("-Created").Offset(100).KeysOnly()
		keys, err := q.GetAll(c, nil)
		if err == nil && len(keys) > 0 {
			datastore.DeleteMulti(c, keys)
		}
	}
	pageTemplate, _ := template.ParseFiles("templates/manage/refresh.txt")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    pageTemplate.Execute(w, map[string]interface{}{"notice": session.Flashes()})
}


/*
 * Status struct for storing updates
 */

type Status struct {
	Name string
	OriginalId int64
	Heading string
	Content string
	Link string
	Created time.Time
	User string
	UserUrl string
}

func (s *Status) Type() string {
	return "Status"
}

func (s *Status) NameTitle() string {
	return strings.Title(s.Name)
}

func Latest(name string) Status {
	updates := make([]Status, 0)
	q := datastore.NewQuery("Status").Filter("Name =", name).Order("-Created").Limit(1)
	q.GetAll(c, &updates)
	if len(updates) > 0 {
		return updates[0]
	}
	return Status{}
}

func Timeline(page int) []*Status {
	q := datastore.NewQuery("Status").Order("-Created").Limit(9).Offset((page-1) * 9)
	var timeline []*Status
	q.GetAll(c, &timeline)
	return timeline
}
