package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/pat"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	// "github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	// "github.com/markbates/goth/providers/vk"
)

func main() {
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "https://my-page-vhfo.onrender.com/auth/google/callback"),
		// github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), "https://my-page-vhfo.onrender.com/auth/github/callback"),
		// vk.New(os.Getenv("VK_KEY"), os.Getenv("VK_SECRET"), "https://my-page-vhfo.onrender.com/auth/vk/callback"),
	)

	m := map[string]string{
		"github": "Github",
		"google": "Google",
		"vk":     "VK",
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	p := pat.New()

	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			return
		}
		t, _ := template.New("").Parse(userTemplate)
		t.Execute(res, user)
	})

	p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		// try to get the user without re-authenticating
		if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
			t, _ := template.New("foo").Parse(userTemplate)
			t.Execute(res, gothUser)
		} else {
			gothic.BeginAuthHandler(res, req)
		}
	})

	p.PathPrefix("/src/").Handler(http.StripPrefix("/src/", http.FileServer(http.Dir("src")))) // for css

	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.ParseFiles("index.html")
		t.Execute(res, providerIndex)
	})

	var port = fmt.Sprintf(":%s", os.Getenv("PORT"))

	log.Println("Start the server")
	log.Fatal(http.ListenAndServe(port, p))
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`
