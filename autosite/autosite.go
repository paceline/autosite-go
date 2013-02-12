/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite


import (
    "appengine"
    "appengine/datastore"
    "html/template"
    "log"
    "net/http"
    "github.com/garyburd/go-oauth/oauth"
    "github.com/gorilla/mux"
    "github.com/gorilla/schema"
    "github.com/gorilla/sessions"
    "strings"
)


// Global vars
var c appengine.Context
var oauthClient oauth.Client
var decoder *schema.Decoder
var store *sessions.CookieStore
var session *sessions.Session
var vars map[string]string

func init() {
	router := mux.NewRouter()
	decoder = schema.NewDecoder()
	store = sessions.NewCookieStore([]byte("what-a-great-secret"))
	
	// GET/PUT '/manage'
	router.HandleFunc("/manage", func(w http.ResponseWriter, r *http.Request) {
		var site Site
		var key string
		switch extendMethod(r) {
			case "GET":
				key = Get(&site)
			case "POST":
				Build(&site, r)
				site.TrackerCode = []byte(r.FormValue("TrackerCode"))
				if r.FormValue("Key") != "" {
					key = Update(&site, r.FormValue("Key"))
					break
				}
				key = Save(&site)
		}
		render(w, []string{"manage","index"}, map[string]interface{}{"key": key, "content": &site})
	})
  
	// GET/POST/PUT '/manage/networks'
	router.HandleFunc("/manage/networks", NetworksHandler)
	router.HandleFunc("/manage/networks/{slug}", NetworksHandler)
	
	// GET '/manage/pages/new'
	router.HandleFunc("/manage/pages/new", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "GET":
				render(w, []string{"manage","pages","new"}, map[string]interface{}{})
		}
	})
	
	// POST '/manage/pages/sort'
	router.HandleFunc("/manage/pages/sort", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "POST":
				r.ParseForm()
				SortPages(strings.Split(r.Form.Get("order"),","))
		}
	})
	
	// DELETE '/manage/pages/{slug}'
	router.HandleFunc("/manage/pages/{slug}", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "DELETE":
				var page Page
				key := GetByName(&page, vars["slug"])
				Delete(key)
		}
	})
	
	// GET '/manage/pages/{slug}/edit'
	router.HandleFunc("/manage/pages/{slug}/edit", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "GET":
				var page Page
				key := GetByName(&page, vars["slug"])
				render(w, []string{"manage","pages","edit"}, map[string]interface{}{"key": key, "content": &page})
		}
	})
	
	// GET/POST '/manage/pages'
	router.HandleFunc("/manage/pages", func(w http.ResponseWriter, r *http.Request) {
		var page Page
		switch extendMethod(r) {
			case "GET":
				pages := make([]Page, 0)
				q := datastore.NewQuery("Page").Order("Position")
				q.GetAll(c, &pages)
				render(w, []string{"manage","pages"}, map[string]interface{}{"content": &pages})
			case "POST":
				Build(&page, r)
				pos, _ := Count(&page)
				page.Position = pos + 1
				page.Body = []byte(r.FormValue("Body"))
				key := Save(&page)
				render(w, []string{"manage","pages","edit"}, map[string]interface{}{"key": key, "content": &page})
			case "PUT":
				page.Published = false
				Build(&page, r)
				page.Body = []byte(r.FormValue("Body"))
				key := Update(&page, r.FormValue("Key"))
				render(w, []string{"manage","pages","edit"}, map[string]interface{}{"key": key, "content": &page})
		}
	})
	
	// GET '/auth/twitter'
	router.HandleFunc("/auth/twitter", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "GET":
				twitterAccount := prepareTwitterConnection(r)
				twitterAccount.ServeLogin(w, r)
		}
	})
	
	// GET '/auth/twitter/callback
	router.HandleFunc("/auth/twitter/callback", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "GET":
				twitterAccount := prepareTwitterConnection(r)
				twitterAccount.ServeOAuthCallback(w, r)
		
		}
	})
	
	// GET '/refresh'
	router.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			case "GET":
				twitterAccount := prepareTwitterConnection(r)
				twitterAccount.GetTwitterUpdates(w, r)
		}
	})
	
	http.Handle("/", router)
}

// Handler: Deal with all requests related to networks entity
func NetworksHandler(w http.ResponseWriter, r *http.Request) {
		var key string
		vars = mux.Vars(r)
		account := Account{Name: "twitter"}
		if vars["slug"] != "" {
			account.Name = vars["slug"]
		}
		switch extendMethod(r) {
			case "GET":
				key = GetByName(&account, account.Name)
			case "PUT":
				Build(&account, r)
				key = Update(&account, r.FormValue("Key"))
			case "POST":
				Build(&account, r)
				key = Save(&account)
		}
		render(w, []string{"manage","networks"}, map[string]interface{}{"key": key, "content": &account})
}

// Helper: Parses and returns template files for given url pattern
func render(w http.ResponseWriter, url []string, pageData map[string]interface{})  {
	if flashes := session.Flashes(); len(flashes) > 0 {
		pageData["notice"] = flashes
    }
	pageTemplate, err := template.ParseFiles("templates/" + url[0] + "/base.html", "templates/" + strings.Join(url,"/") + ".html")
    if err != nil { 
      log.Fatalf("Template parsing error: %s", err)
    }
    pageTemplate.Execute(w, pageData)
}

// Helper: Use _method form field to support PUT and DELETE requests just like the regular GET and POST (-> RESTful routes)
func extendMethod(r *http.Request) string {
	c = appengine.NewContext(r)
	session, _ = store.Get(r, "autosite-go-session")
	vars = mux.Vars(r)
	if r.Method == "POST" && (strings.ToUpper(r.FormValue("_method")) == "PUT" || strings.ToUpper(r.FormValue("_method")) == "DELETE")  {
		return strings.ToUpper(r.FormValue("_method"))
	}
	return r.Method
}

// Helper: return OAuth settings for Twitter
func prepareTwitterConnection(r *http.Request) Account {
	var account Account
	GetByName(&account, "twitter")
	oauthClient = oauth.Client{
		TemporaryCredentialRequestURI: "http://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "http://api.twitter.com/oauth/authenticate",
		TokenRequestURI:               "http://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{Token: account.ConsumerKey, Secret: account.ConsumerSecret},
	}
	return account
}
