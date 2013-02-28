/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
*/

package autosite

import (
	"appengine/urlfetch"
	"github.com/paceline/goauth2/oauth"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
 * OAuth2 Client
 */

// Global vars
var oauth2Config *oauth.Config

// Authorize Callback
func (a *Account) ServeOAuth2Callback(r *http.Request) {
	code := r.FormValue("code")
	t := oauth.Transport{oauth.Config: oauth2Config, Transport: &urlfetch.Transport{Context: c}}
	tokenCred, err := t.Exchange(code)
	if err != nil {
		session.AddFlash("Error during " + a.Name + " authentication: " + err.Error())
		return
	}
	a.Token = tokenCred.AccessToken
}

// OAuth2 settings
func (a *Account) prepareOAuth2Connection(r *http.Request, scope string) {
	oauth2Config = &oauth.Config {
		ClientId: a.ConsumerKey,
        ClientSecret: a.ConsumerSecret,
        AuthURL: a.AuthUrl,
        TokenURL: a.AccessUrl,
        RedirectURL: "http://" + r.Host + "/auth/" + a.Name + "/callback",
        Scope: scope,
	}
}


/*
 * GitHub Client
 */
 

// Get Updates from GitHub
func (a *Account) GetGithubUpdates(r *http.Request) {
	
	// Initialize connection
	var tweets []map[string]string
	t := oauth.Transport{oauth.Config: oauth2Config, oauth.Token: &oauth.Token{AccessToken: a.Token}, Transport: &urlfetch.Transport{Context: c}}
	latest := Latest("github")
	login := latest.User
	
	// Get authenticated user
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	if latest.OriginalId > 0 {
		req.Header.Set("If-Modified-Since", latest.Created.Format("Mon, 2 Jan 2006 15:04:05 MST"))
	}
	resp, err := t.Client().Do(req)
	if err != nil {
		session.AddFlash("Error getting " + a.Name + " user info: " + err.Error())
		return
	}
	if resp.StatusCode == 200 {
		var user map[string]interface{}
		defer resp.Body.Close()
		decodeResponse(resp, &user)
		login = user["login"].(string)
	}
	
	// Fire request and save timeline
	req, _ = http.NewRequest("GET", "https://api.github.com/users/" + login + "/events", nil)
	if latest.OriginalId > 0 {
		req.Header.Set("If-Modified-Since", latest.Created.Format("Mon, 2 Jan 2006 15:04:05 MST"))
	}
	resp, err = t.Client().Do(req)
	if err != nil {
		session.AddFlash("Error getting " + a.Name + " updates: " + err.Error())
		return
	}
	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		var timeline []map[string]interface{}
		decodeResponse(resp, &timeline)
		for i := 0; i < len(timeline); i++ {
			created_at, _ := time.Parse("2006-01-02T15:04:05Z", timeline[i]["created_at"].(string))
			if created_at.After(latest.Created) {  
				id, _ := strconv.ParseInt(timeline[i]["id"].(string), 10, 64)
				profileUrl := "https://github.com/" + login
				link := "https://github.com/" + strings.ToLower(timeline[i]["repo"].(map[string]interface {})["name"].(string))
				var title string
				var text string
				switch (timeline[i]["type"].(string)) {
					case "CommitCommentEvent", "PullRequestReviewCommentEvent":
						title = "commented"
						text = timeline[i]["payload"].(map[string]interface {})["comment"].(map[string]interface {})["body"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["comment"].(map[string]interface {})["html_url"].(string)
					case "CreateEvent":
						ref_type := timeline[i]["payload"].(map[string]interface {})["ref_type"].(string)
						ref := timeline[i]["repo"].(map[string]interface {})["name"].(string)
						if ref_type != "repository" {
							ref = timeline[i]["payload"].(map[string]interface {})["ref"].(string)
						}
						title = "created " + ref_type + " " + ref
					case "DeleteEvent":
						title = "deleted " + timeline[i]["payload"].(map[string]interface {})["ref_type"].(string)
					case "DownloadEvent":
						title = "uploaded " + timeline[i]["payload"].(map[string]interface {})["download"].(map[string]interface {})["name"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["download"].(map[string]interface {})["html_url"].(string)
					case "FollowEvent":
						title = "is now following " + timeline[i]["payload"].(map[string]interface {})["target"].(map[string]interface {})["login"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["target"].(map[string]interface {})["html_url"].(string)
					case "ForkEvent":
						title = "forked " + timeline[i]["repo"].(map[string]interface {})["name"].(string)
					case "ForkApplyEvent":
						title = "applied fork to " + timeline[i]["payload"].(map[string]interface {})["head"].(string)
					case "GistEvent":
						title = timeline[i]["payload"].(map[string]interface {})["action"].(string) + "d " + timeline[i]["payload"].(map[string]interface {})["gist"].(map[string]interface {})["description"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["gist"].(map[string]interface {})["html_url"].(string)
					case "GollumEvent":
						title = "updated pages"
						text = strings.Join(timeline[i]["payload"].(map[string]interface {})["pages"].([]string), ", ")
					case "IssueCommentEvent":
						title = timeline[i]["payload"].(map[string]interface {})["action"].(string) + " on issue #" +  timeline[i]["payload"].(map[string]interface {})["issue"].(map[string]interface {})["number"].(string)
						text = timeline[i]["payload"].(map[string]interface {})["comment"].(map[string]interface {})["body"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["issue"].(map[string]interface {})["html_url"].(string)
					case "IssuesEvent":
						title = timeline[i]["payload"].(map[string]interface {})["action"].(string) + " issue #" +  timeline[i]["payload"].(map[string]interface {})["issue"].(map[string]interface {})["number"].(string)
						text = timeline[i]["payload"].(map[string]interface {})["issue"].(map[string]interface {})["body"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["issue"].(map[string]interface {})["html_url"].(string)
					case "MemberEvent":
						title = timeline[i]["payload"].(map[string]interface {})["action"].(string) + " " +  timeline[i]["payload"].(map[string]interface {})["member"].(map[string]interface {})["login"].(string) + " to " + timeline[i]["repo"].(map[string]interface {})["name"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["member"].(map[string]interface {})["html_url"].(string)
					case "PublicEvent":
						title = "Open sourced " + timeline[i]["repo"].(map[string]interface {})["name"].(string) 
					case "PullRequestEvent":
						title = timeline[i]["payload"].(map[string]interface {})["action"].(string) + " pull request"
						text = timeline[i]["payload"].(map[string]interface {})["pull_request"].(map[string]interface {})["title"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["pull_request"].(map[string]interface {})["html_url"].(string)
					case "PushEvent":
						title = "pushed to " + timeline[i]["repo"].(map[string]interface {})["name"].(string)
						text = timeline[i]["payload"].(map[string]interface {})["commits"].([]interface{})[0].(map[string]interface{})["message"].(string)
						if a.Repost {
							tweets = append(tweets, map[string]string{"status": "I updated my app #" + strings.Replace(strings.Split(timeline[i]["repo"].(map[string]interface {})["name"].(string), "/")[1], "-", "", -1) + " on @github: " + text, "link": link})
						}
					case "TeamAddEvent":
						title = "added " + timeline[i]["payload"].(map[string]interface {})["user"].(map[string]interface {})["login"].(string) + " to " + timeline[i]["payload"].(map[string]interface {})["team"].(map[string]interface {})["name"].(string)
						link = timeline[i]["payload"].(map[string]interface {})["user"].(map[string]interface {})["html_url"].(string)
					case "WatchEvent":
						title = timeline[i]["payload"].(map[string]interface {})["action"].(string) + " watching " + timeline[i]["repo"].(map[string]interface {})["name"].(string)
				}
				if len(title) > 0 {
					update := Status{Name: "github", OriginalId: id, Heading: title, Link: link, Content: text, Created: created_at, User: login, UserUrl: profileUrl}
					Save(&update)
				}
			}
		}
		if a.Repost {
			var twitter Account
			GetByName(&twitter, "twitter")
			twitter.PostTwitterUpdate(tweets)
		}
    }
}
