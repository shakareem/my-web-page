package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	// "github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	// "github.com/markbates/goth/providers/vk"
)

var store = sessions.NewCookieStore([]byte("session-secret"))

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

	p.PathPrefix("/src/").Handler(http.StripPrefix("/src/", http.FileServer(http.Dir("src")))) // for css

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.BeginAuthHandler(res, req)
	})

	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {
		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			fmt.Fprintln(res, err)
			log.Printf("Error completing user auth: %v", err)
			return
		}

		// Save user in session
		session, _ := store.Get(req, "session-name")
		session.Values["user"] = user
		session.Save(req, res)

		http.Redirect(res, req, "/", http.StatusSeeOther)
	})

	p.Get("/logout", func(res http.ResponseWriter, req *http.Request) {
		session, _ := store.Get(req, "session-name")
		delete(session.Values, "user")
		session.Save(req, res)
		http.Redirect(res, req, "/", http.StatusSeeOther)
	})

	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		session, _ := store.Get(req, "session-name")
		user, loggedIn := session.Values["user"].(goth.User)

		t, _ := template.ParseFiles("index.html")
		t.Execute(res, map[string]interface{}{
			"Providers": providerIndex,
			"User":      user,
			"LoggedIn":  loggedIn,
		})
	})

	var port = fmt.Sprintf(":%s", os.Getenv("PORT"))

	log.Println("Start the server")
	log.Fatal(http.ListenAndServe(port, p))
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}
