/*
    Package autosite provides a simple infrastructure for running a
    personal website (off of the Google App Engine)
    
    Created by Ulf Möhring <hello@ulfmoehring.net>
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


// Authorize
func (a *Account) ServeLogin(w http.ResponseWriter, r *http.Request) {
	tempCred, err := oauthClient.RequestTemporaryCredentials(urlfetch.Client(c), "http://" + r.Host + "/auth/" + a.Name + "/callback", nil)
	if err != nil {
		http.Error(w, "Error getting temp cred, "+err.Error(), 500)
		return
	}
	a.Token = tempCred.Token
	a.Secret = tempCred.Secret
	//a.Save()
	http.Redirect(w, r, oauthClient.AuthorizationURL(tempCred, nil), 302)
}

// Authorize Callback
func (a *Account) ServeOAuthCallback(w http.ResponseWriter, r *http.Request) {
	tempCred := oauth.Credentials{Token: r.FormValue("oauth_token"), Secret: a.Secret}
	tokenCred, _, err := oauthClient.RequestToken(urlfetch.Client(c), &tempCred, r.FormValue("oauth_verifier"))
	if err != nil {
		http.Error(w, "Error getting request token, "+err.Error(), 500)
		return
	}
	a.Token = tokenCred.Token
	a.Secret = tokenCred.Secret
	//a.Save()
	http.Redirect(w, r, "/manage/networks", 302)
}

// apiGet issues a GET request to the Twitter API and decodes the response JSON to data.
func (a *Account) apiGet(urlStr string, form url.Values, data interface{}) error {
	cred := a.Creds()
	resp, err := oauthClient.Get(urlfetch.Client(c), &cred, urlStr, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, data)
}

// decodeResponse decodes the JSON response from the Twitter API.
func decodeResponse(resp *http.Response, data interface{}) error {
	if resp.StatusCode != 200 {
		p, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Get %s returned status %d, %s", resp.Request.URL, resp.StatusCode, p)
	}
	return json.NewDecoder(resp.Body).Decode(data)
}


/*
 * Twitter Client
 */

// Get Updates from Twitter
func (a *Account) GetTwitterUpdates(w http.ResponseWriter, r *http.Request) {
	var timeline []map[string]interface{}
	if err := a.apiGet(
		"https://api.twitter.com/1.1/statuses/user_timeline.json",
		nil,
		&timeline); err != nil {
		http.Error(w, "Error getting timeline, "+err.Error(), 500)
		return
	}
	for i := 0; i < len(timeline); i++ {
		created_at, _ := time.Parse("Mon Jan 2 15:04:05 -0700 2006", timeline[i]["created_at"].(string))
		urlextractor, _ := regexp.Compile(" http://[a-zA-Z0-9\\./-]*")
		link := strings.TrimLeft(urlextractor.FindString(timeline[i]["text"].(string)), " ")
		post := urlextractor.ReplaceAllString(timeline[i]["text"].(string),"")
		id, _ := strconv.ParseInt(timeline[i]["id_str"].(string), 10, 64)
		update := Status{Kind: "Twitter", OriginalId: id, Heading: post, Link: link, Created: created_at}
		update.Save()
    }
}
