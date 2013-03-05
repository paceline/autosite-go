/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <ulf@moehring.me>
*/

package autosite

import (
	"appengine/urlfetch"
    "encoding/json"
    "fmt"
    "io/ioutil"
	"github.com/garyburd/go-oauth/oauth"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"strconv"
	"time"
)


/*
 * OAuth Client
 */
 
// Global vars
var oauthClient oauth.Client

// Authorize
func (a *Account) ServeLogin(w http.ResponseWriter, r *http.Request) *oauth.Credentials {
	tempCred, err := oauthClient.RequestTemporaryCredentials(urlfetch.Client(c), "http://" + r.Host + "/auth/" + a.Name + "/callback", nil)
	if err != nil {
		session.AddFlash("Error during " + a.Name + " authentication: " + err.Error())
		return nil
	}
	a.Token = tempCred.Token
	a.Secret = tempCred.Secret
	return tempCred
}

// Authorize Callback
func (a *Account) ServeOAuthCallback(r *http.Request) {
	tempCred := oauth.Credentials{Token: r.FormValue("oauth_token"), Secret: a.Secret}
	tokenCred, _, err := oauthClient.RequestToken(urlfetch.Client(c), &tempCred, r.FormValue("oauth_verifier"))
	if err != nil {
		session.AddFlash("Error during " + a.Name + " authentication: " + err.Error())
		return
	}
	a.Token = tokenCred.Token
	a.Secret = tokenCred.Secret
}

// apiGet issues a GET request to the API and decodes the response JSON to data.
func (a *Account) apiGet(urlStr string, form url.Values, data interface{}) error {
	cred := a.Creds()
	resp, err := oauthClient.Get(urlfetch.Client(c), &cred, urlStr, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, data)
}

// apiPost issues a POST request to the API and decodes the response JSON to data.
func (a *Account) apiPost(urlStr string, form url.Values) (string, error) {
	cred := a.Creds()
	resp, err := oauthClient.Post(urlfetch.Client(c), &cred, urlStr, form)
	if err != nil {
		return "", err
	}
	msg, err := ioutil.ReadAll(resp.Body)
	return string(msg), nil
}

// decodeResponse decodes the JSON response from the Twitter API.
func decodeResponse(resp *http.Response, data interface{}) error {
	if resp.StatusCode != 200 {
		p, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Get %s returned status %d, %s", resp.Request.URL, resp.StatusCode, p)
	}
	return json.NewDecoder(resp.Body).Decode(data)
}

// Return oauth creds
func (a *Account) Creds() oauth.Credentials {
	return oauth.Credentials{Token: a.Token, Secret: a.Secret}
}

// OAuth settings
func (a * Account) prepareOAuthConnection(r *http.Request) {
	oauthClient = oauth.Client{
		TemporaryCredentialRequestURI: a.RequestUrl,
		ResourceOwnerAuthorizationURI: a.AuthUrl,
		TokenRequestURI:               a.AccessUrl,
		Credentials: oauth.Credentials{Token: a.ConsumerKey, Secret: a.ConsumerSecret},
	}
}


/*
 * Twitter Client
 */
 

// Get Updates from Twitter
func (a *Account) GetTwitterUpdates(r *http.Request) {
	var timeline []map[string]interface{}
	params := url.Values{}
	srcmatch, _ := regexp.Compile("Autosite</a>$")
	if latest := Latest("twitter"); latest.OriginalId > 0 {
		params.Add("since_id", strconv.FormatInt(latest.OriginalId, 10))
	}
	if err := a.apiGet("https://api.twitter.com/1.1/statuses/user_timeline.json", params, &timeline); err != nil {
		session.AddFlash("Error getting " + a.Name + " updates: " + err.Error())
		return
	}
	for i := 0; i < len(timeline); i++ {
		if srcmatch.FindString(timeline[i]["source"].(string)) == "" {
			created_at, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", timeline[i]["created_at"].(string))
			urlextractor, _ := regexp.Compile(" http://[a-zA-Z0-9\\./-]*")
			link := strings.TrimLeft(urlextractor.FindString(timeline[i]["text"].(string)), " ")
			post := urlextractor.ReplaceAllString(timeline[i]["text"].(string),"")
			id, _ := strconv.ParseInt(timeline[i]["id_str"].(string), 10, 64)
			profileName := timeline[i]["user"].(map[string]interface {})["screen_name"].(string)
			profileUrl := "https://twitter.com/" + profileName
			update := Status{Name: "twitter", OriginalId: id, Heading: post, Link: link, Created: created_at, User: profileName, UserUrl: profileUrl}
			Save(&update)
		}
    }
}

// Post updates to Twitter
func (a *Account) PostTwitterUpdate(posts []map[string]string) {
	for i := 1; i <= len(posts); i++ {
		post := posts[len(posts) - i]
		if len(post["status"]) > 119 {
			post["status"] = post["status"][0:115] + "..."  
		}
		msg, err := a.apiPost("https://api.twitter.com/1.1/statuses/update.json", url.Values{"status": {post["status"] + " " + post["link"]}})
		if err != nil {
			session.AddFlash("Error posting " + a.Name + " update: " + err.Error())
		}
		session.AddFlash("Response posting " + a.Name + " update '" +  post["status"] + "': " + msg)
	}
}
