package main

import (
	"fmt"
	"log"
	"net/http"
	"html/template"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/sessions"
)

var consumerKey, consumerSecret string
var store *sessions.CookieStore

const (
	sessionName   = "session_name"
	sessionSecret = "session_secret"
)

func ouathClient() *oauth.Client {
	return &oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  consumerKey,
			Secret: consumerSecret,
		},
	}
}


func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("index").Parse(`
<html>
<body>
<a href="/request_token">Twitter認証</a>
</body>
</html>`))
	t.Execute(w, nil)
}

func RequestTokenHandler(w http.ResponseWriter, r *http.Request) {
	oc := ouathClient()
	callback := "http://localhost:8080/access_token"
	tempCred, err := oc.RequestTemporaryCredentials(nil, callback, nil)
	if err != nil {
		http.Error(w, "Error getting temp cred, "+err.Error(), 500)
		return
	}
	s, _ := store.Get(r, sessionName)
	s.Values["request_token"] = tempCred.Token
	s.Values["request_token_secret"] = tempCred.Secret
	if err := store.Save(r, w, s); err != nil {
		http.Error(w, "Error saving session , "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, oc.AuthorizationURL(tempCred, nil), 302)
}

func AccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := store.Get(r, sessionName)
	requestToken, _ := s.Values["request_token"].(string)
	requestTokenSecret, _ := s.Values["request_token_secret"].(string)

	if requestToken != r.FormValue("oauth_token") {
		http.Error(w, "Unknown oauth_token.", 500)
		return
	}

	oc := ouathClient()
	tempCred := oauth.Credentials{
		Token: requestToken,
		Secret: requestTokenSecret,
	}
	tokenCred, _, err := oc.RequestToken(nil, &tempCred, r.FormValue("oauth_verifier"))

	if err != nil {
		http.Error(w, "Error getting request token, "+err.Error(), 500)
		return
	}

	t := template.Must(template.New("index").Parse(`
<html>
<body>
<p>ACCESS TOKEN: {{ .Token }}</p>
<p>ACCESS TOKEN SECRET: {{ .Secret }}</p>
</body>
</html>`))

	t.Execute(w, tokenCred)
}

func main() {
	fmt.Println("Enter ConsumerKey: ")
	fmt.Scan(&consumerKey)

	//ck, err := terminal.ReadPassword(0)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//consumerKey = string(ck)

	fmt.Println("Enter ConsumerSecret: ")
	cs, err := terminal.ReadPassword(0)
	if err != nil {
		log.Fatal(err)
	}
	consumerSecret = string(cs)
	store = sessions.NewCookieStore([]byte(sessionSecret))

	fmt.Println("Start Server")

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/request_token", RequestTokenHandler)
	http.HandleFunc("/access_token", AccessTokenHandler)
	http.ListenAndServe(":8080", nil)
}
