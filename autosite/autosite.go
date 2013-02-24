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
    "net/http"
    "github.com/gorilla/mux"
    "github.com/gorilla/schema"
    "github.com/gorilla/sessions"
    "log"
    "os"
    "strconv"
    "strings"
    "time"
)


// Global vars
var c appengine.Context
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
				site.TemplateKey = ToKey(r.FormValue("TemplateKey"))
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
		if extendMethod(r) == "GET" {
			render(w, []string{"manage","pages","new"}, map[string]interface{}{})
		}
	})
	
	// POST '/manage/pages/sort'
	router.HandleFunc("/manage/pages/sort", func(w http.ResponseWriter, r *http.Request) {
		if extendMethod(r) == "POST" {
			r.ParseForm()
			SortPages(strings.Split(r.Form.Get("order"),","))
		}
	})
	
	// DELETE '/manage/pages/{slug}'
	router.HandleFunc("/manage/pages/{slug}", func(w http.ResponseWriter, r *http.Request) {
		if extendMethod(r) == "DELETE" {
			var page Page
			key := GetByName(&page, vars["slug"])
			Delete(key)
		}
	})
	
	// GET '/manage/pages/{slug}/edit'
	router.HandleFunc("/manage/pages/{slug}/edit", func(w http.ResponseWriter, r *http.Request) {
		if extendMethod(r) == "GET" {
			var page Page
			render(w, []string{"manage","pages","edit"}, map[string]interface{}{"key": GetByName(&page, vars["slug"]), "content": &page})
		}
	})
	
	// GET/POST '/manage/pages'
	router.HandleFunc("/manage/pages", func(w http.ResponseWriter, r *http.Request) {
		var page Page
		method := extendMethod(r)
		if method == "POST" || method == "PUT" {
			var key string
			Build(&page, r)
			page.Body = []byte(r.FormValue("Body"))
			switch method {
				case "POST":
					pos, _ := Count("Page")
					page.Position = pos + 1
					key = Save(&page)
				case "PUT":
					key = Update(&page, r.FormValue("Key"))
			}
			render(w, []string{"manage","pages","edit"}, map[string]interface{}{"key": key, "content": &page})
			return
		}
		pages := make([]Page, 0)
		q := datastore.NewQuery("Page").Order("Position")
		q.GetAll(c, &pages)
		render(w, []string{"manage","pages"}, map[string]interface{}{"content": &pages})
	})
	
	// GET '/sign_out'
	router.HandleFunc("/sign_out", func(w http.ResponseWriter, r *http.Request) {
		if extendMethod(r) == "GET" {
			cookie, err := r.Cookie("ACSID")
			if cookie != nil && err == nil {
				w.Header().Set("Set-Cookie", "ACSID=deleted; Expires=Thu, 01-Jan-1970 00:00:00 GMT; Domain=" + cookie.Domain + "; Path=" + cookie.Path)
			}
		}
		http.Redirect(w, r, "/", 302)
	})
	
	// GET '/auth/twitter'
	router.HandleFunc("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		url := "/"
		if extendMethod(r) == "GET" {
			var account Account
			key := GetByName(&account, vars["provider"])
			if key != "" {
				if account.Version() == 1 {
					account.prepareOAuthConnection(r)
					creds := account.ServeLogin(w, r)
					Update(&account, key)
					url = oauthClient.AuthorizationURL(creds, nil)
				} else {
					switch account.Name {
						case "github":
							account.prepareOAuth2Connection(r, "user")
					}
					url = oauth2Config.AuthCodeURL("")
				}
			}
		}
		http.Redirect(w, r, url, 302)
	})
	
	// GET '/auth/twitter/callback
	router.HandleFunc("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		if extendMethod(r) == "GET" {
			var account Account
			key := GetByName(&account, vars["provider"])
			if key != "" {
				if account.Version() == 1 {
					account.prepareOAuthConnection(r)
					account.ServeOAuthCallback(r)
				} else {
					switch account.Name {
						case "github":
							account.prepareOAuth2Connection(r, "user")
					}
					account.ServeOAuth2Callback(r)
				}
				Update(&account, key)
				render(w, []string{"manage","networks"}, map[string]interface{}{"key": key, "content": &account})
			}
		}
	})
	
	// GET '/manage/refresh'
	router.HandleFunc("/manage/refresh", Refresh)
	
	// GET '/'
	router.HandleFunc("/", RootHandler)
	router.HandleFunc("/timeline/{page}", RootHandler)
	router.HandleFunc("/{slug}", RootHandler)
	
	// Hook-up router to go http package
	http.Handle("/", router)
}

// Handler: Deal with all requests related to visitor frontend
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if extendMethod(r) == "GET" {
		var site Site
		Get(&site)
		if vars["slug"] != "" {
			var page Page
			GetByName(&page, vars["slug"])
			if len(page.Name) > 0 {
				if page.IsTemplate() {
					css, _ := template.New("css").Parse(page.BodyString())
					w.Header().Set("Content-Type", "text/css; charset=utf-8")
					css.Execute(w, nil)
					return
				}
				render(w, []string{"page"}, map[string]interface{}{"site": &site, "page": &page})
				return
			}
		}
		current := 1
		if vars["page"] != "" {
			temp, _ := strconv.ParseInt(vars["page"], 10, 0)
			current = int(temp)
		}
		timeline := Timeline(current)
		render(w, []string{"index"}, map[string]interface{}{"site": &site, "timeline": timeline, "current":current})
	}
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
    layout := "templates/" + url[0] + "/base.html"
    if _, err := os.Stat("templates/" + url[0]); os.IsNotExist(err) {
		layout = "templates/base.html"
    }
    funcMap := template.FuncMap {
		"formatTime": formatTime,
		"htmlSafe": htmlSafe,
		"navigation": navigation,
		"pagination": pagination,
	}
    pageTemplate, _ := template.New("website").Funcs(funcMap).ParseFiles(layout, "templates/" + strings.Join(url,"/") + ".html")
    if err := pageTemplate.Execute(w, pageData); err != nil {
		log.Printf("Somethings wrong: %s", err.Error())
	}	
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

// Helper: View wrapper for html safe
func htmlSafe(text string) template.HTML {
	return template.HTML(text)
}

// Helper: Return time in proper format
func formatTime(t time.Time) string {
	elapsed := time.Since(t).Seconds()
	switch {
		case elapsed < 60:
			return "A few seconds ago"
		case elapsed >= 60 && elapsed < 3600:
			return "About " + strconv.FormatFloat(elapsed/60, 'f', 0, 64) + " minutes ago"
		case elapsed >= 3600 && elapsed < 86400:
			return "About " + strconv.FormatFloat(elapsed/60/60, 'f', 0, 64) + " hours ago"
		case elapsed >= 86400 && elapsed < 604800:
			return "About " + strconv.FormatFloat(elapsed/60/60/24, 'f', 0, 64) + " days ago"
		case elapsed >= 604800 && elapsed < 2419200:
			return "About " + strconv.FormatFloat(elapsed/60/60/24/7, 'f', 0, 64) + " weeks ago"
		case elapsed >= 2419200 && elapsed < 29030400:
			return "About " + strconv.FormatFloat(elapsed/60/60/24/7/4, 'f', 0, 64) + " months ago"
	}
	return "More than a year ago"
}

// Helper: Render navigation
func navigation() []map[string]string {
	var pages []map[string]string
	q := datastore.NewQuery("Page").Filter("Published = ", true).Order("Position")
	for t := q.Run(c); ; {
		var page Page
		_, err := t.Next(&page)
		if err == datastore.Done {
			break
        }
        if err != nil {
			session.AddFlash("An error occured while loading: %s", err.Error())
			break
        }
        if !page.IsTemplate() {
			pages = append(pages, map[string]string{"Name": page.Name, "Title": page.Title})
		}
	}
	return pages
}

// Helper: Render pagination widget
func pagination(current int) template.HTML {
	var snippet string
	n, _ := Count("Status")
	for i := 1; i <= (n / 9); i++ {
		var temp string
		index := strconv.FormatInt(int64(i), 10)
		if i == 1 {
			temp = `<a class="previous_page" rel="next" href="/timeline/` + index + `">&#8592; Previous</a>
	`
			if i == current {
				temp = `<span class="previous_page disabled">&#8592; Previous</span>
	`
			}
		}
		if i == current {
			temp = temp + `<em class="current">` + index + `</em>
	`
		}
		if i != current {
			temp = temp + `<a href="/timeline/` + index + `">` + index + `</a>
	`
		}
		if i == (n / 9) {
			if i == current {
				temp = temp + `<span class="next_page disabled">Next &#8594;</span>`
			}
			if i != current {
				temp = temp + `<a class="next_page" rel="next" href="/timeline/` + index + `">Next &#8594;</a>`
			}
		}
		snippet = snippet + temp
	}
	return template.HTML(snippet)
}
