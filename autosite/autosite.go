/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/
package autosite

import (
    "html/template"
    "log"
    "net/http"
    "strings"
)

// Map out actions
func init() {
	
	// GET/POST '/manage'
	http.HandleFunc("/manage", func(w http.ResponseWriter, r *http.Request) {
		var site Site
		switch extendMethod(r) {
			case "GET":
				site.Load(r)
				render(w, []string{"manage","index"}, site)
			case "PUT":
				site.Save(r)
				render(w, []string{"manage","index"}, site)
			default:
				render(w, []string{"manage","error"}, nil)
		}
	})
	
	// GET/POST/PUT/DELETE '/manage/pages'
	http.HandleFunc("/manage/pages", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			default:
				render(w, []string{"manage","error"}, nil)
				return
			/*case "GET":
				// TODO
			case "POST":
				// TODO
			case "PUT":
				// TODO
			case "DELETE":
				// TODO*/
		}
	})
	
	// GET '/manage/pages/new'
	http.HandleFunc("/manage/pages/new", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
			default:
				render(w, []string{"manage","error"}, nil)
				return
			/*case "GET":
				// TODO
			case "POST":
				// TODO*/
		}
	})
	
	// GET '/manage/pages/edit'
	http.HandleFunc("/manage/pages/edit", func(w http.ResponseWriter, r *http.Request) {
		switch extendMethod(r) {
			default:
				render(w, []string{"manage","error"}, nil)
				return
			/*case "GET":
				// TODO*/
		}
	})
}

// Helper: Parses and returns template files for given url pattern
func render(w http.ResponseWriter, url []string, pageData interface{})  {
	pageTemplate, err := template.ParseFiles("templates/" + strings.Join(url[0:len(url)-1],"/") + "/base.html", "templates/" + strings.Join(url,"/") + ".html")
    if err != nil { 
      log.Fatalf("Template parsing error: %s", err)
    }
    pageTemplate.Execute(w, pageData)
}

// Helper: Use _method form field to support PUT and DELETE requests just like the regular GET and POST (-> RESTful routes)
func extendMethod(r *http.Request) string {
	if r.Method == "POST" && (strings.ToUpper(r.FormValue("_method")) == "PUT" || strings.ToUpper(r.FormValue("_method")) == "DELETE")  {
		return strings.ToUpper(r.FormValue("_method"))
	}
	return r.Method
}
