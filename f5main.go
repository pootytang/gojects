package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/pootytang/gojects/OAuth2/F5Oauth2/f5oauth20"
)

var (
	f5OauthConfig = f5oauth20.F5Config{
		JWT: false,
	}
	fs  = http.FileServer(http.Dir("public/"))
	tpl = template.Must(template.ParseFiles("templates/post.gohtml", "templates/getJWT.gohtml", "templates/displayCode.gohtml"))
)

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/post", handlePost)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	//http.ListenAndServe(":8080", nil)
	http.ListenAndServeTLS(":8080", "https-server.crt", "https-server.key", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fs.ServeHTTP(w, r)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// TODO: clean this up. May be able to do this within f5_oauth2.go instead of here
	// TODO: Test that a POST request was made, if not maybe respond with Header values or Hello World

	r.ParseForm()

	// Grab the posted parameters
	f5OauthConfig.SetEndpoint(r.PostForm["hostname"][0])
	f5OauthConfig.ClientID = f5OauthConfig.CleanString(r.PostForm["clientid"][0])
	f5OauthConfig.ClientSecret = f5OauthConfig.CleanString(r.PostForm["clientsecret"][0])
	f5OauthConfig.RedirectURL = f5OauthConfig.CleanString(r.PostForm["redirecturl"][0])
	f5OauthConfig.CAList = r.PostForm["ca_list"]
	f5OauthConfig.Scopes = r.PostForm["scope_list"]
	if r.PostForm["state"] != nil {
		f5OauthConfig.State = f5OauthConfig.CheckState(r.PostForm["state"][0])
	} else {
		f5OauthConfig.State = ""
	}

	// f5OauthConfig.JWT default is set to false
	if r.PostForm["jwt"] != nil {
		if r.PostForm["jwt"][0] == "JWT" {
			f5OauthConfig.JWT = true
		}
	}
	defer r.Body.Close()

	// configure the request URL or the URL we'll redirect the client to
	f5OauthConfig.AuthCodeURL()
	//fmt.Println(url)

	//defer req.Body.Close()
	//fmt.Println(f5OauthConfig.Endpoint)
	//fmt.Println(f5OauthConfig.Endpoint.AuthURL)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "post.gohtml", f5OauthConfig)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Display a form to POST the data to the APM
	/* content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Fprintf(w, "Content: %s\n", content) */

	// extract the code and store in the f5OauthConfig object
	f5OauthConfig.Code = f5OauthConfig.CleanString(r.FormValue("code"))
	if f5OauthConfig.Code == "" {
		log.Fatalln("Problem getting the token code")
	}
	defer r.Body.Close()

	if f5OauthConfig.JWT == true {
		tpl.ExecuteTemplate(w, "getJWT.gohtml", f5OauthConfig)
	} else {
		tpl.ExecuteTemplate(w, "displayCode.gohtml", f5OauthConfig)
	}
}
