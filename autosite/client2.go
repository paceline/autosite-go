/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <ulf@moehring.me>
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
	a.Expires = tokenCred.Expiry
}

// OAuth2 settings
func (a *Account) prepareOAuth2Connection(r *http.Request) {
	oauth2Config = &oauth.Config {
		ClientId: a.ConsumerKey,
        ClientSecret: a.ConsumerSecret,
        AuthURL: a.AuthUrl,
        TokenURL: a.AccessUrl,
        RedirectURL: "http://" + r.Host + "/auth/" + a.Name + "/callback",
        Scope: a.RequestUrl,
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


/*
 * LinkedIn Client
 */
 

// Get Updates from LinkedIn
func (a *Account) GetLinkedInUpdates(r *http.Request) {
	
	// Initialize connection
	var tweets []map[string]string
	latest := Latest("linkedin")
	url := "https://api.linkedin.com/v1/people/~/network/updates?format=json&scope=self&type=SHAR&oauth2_access_token=" + a.Token
	if latest.OriginalId > 0 {
		url = url + "&after=" + strconv.FormatInt(latest.OriginalId + 1, 10)
	}
	
	// Fire request
	resp, err := urlfetch.Client(c).Get(url)
	if err != nil {
		session.AddFlash("Error getting " + a.Name + " updates: " + err.Error())
		return
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	decodeResponse(resp, &data)
	
	// Parse and save timeline
	if data["_total"] != nil && data["_total"].(float64) > 0 {
		for i := 0; i < len(data["values"].([]interface{})); i++ {
			content := data["values"].([]interface{})[i].(map[string]interface{})["updateContent"].(map[string]interface{})["person"].(map[string]interface{})
			update := Status {
				Name: "linkedin",
				OriginalId: int64(content["currentShare"].(map[string]interface{})["timestamp"].(float64)),
				Heading: "shared a link",
				Created: time.Unix(int64(content["currentShare"].(map[string]interface{})["timestamp"].(float64)*0.001), 0),
				User: content["currentShare"].(map[string]interface{})["author"].(map[string]interface{})["firstName"].(string) + " " + content["currentShare"].(map[string]interface{})["author"].(map[string]interface{})["lastName"].(string),
				UserUrl: strings.Split(content["siteStandardProfileRequest"].(map[string]interface{})["url"].(string), "&")[0],
			}
			if content["currentShare"].(map[string]interface{})["comment"] != nil {
				update.Heading = content["currentShare"].(map[string]interface{})["comment"].(string)
			}
			if content["currentShare"].(map[string]interface{})["content"] != nil {
				update.Link = content["currentShare"].(map[string]interface{})["content"].(map[string]interface{})["submittedUrl"].(string)
				update.Content = content["currentShare"].(map[string]interface{})["content"].(map[string]interface{})["submittedUrl"].(string)
			}
			if a.Repost {
				status := update.Heading
				if status == "shared a link" {
					status = "I " + status
				}
				link := update.Link
				if len(link) == 0 {
					link = update.UserUrl
				}
				tweets = append(tweets, map[string]string{"status": status, "link": link})
			}
			Save(&update)
		}
		if a.Repost {
			var twitter Account
			GetByName(&twitter, "twitter")
			twitter.PostTwitterUpdate(tweets)
		}
	}
}
