package main

import (
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Get("/api/check", func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		checkDisposable := strings.TrimSpace(r.URL.Query().Get("check_disposable"))

		if email == "" {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20000, "error": "email query string not given"})
			return
		}

		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "error": "invalid email format"})
			return
		}

		if len(email) < 8 || len(email) > 254 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20002, "error": "email must be from 8 to 254 characters"})
			return
		}

		re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !re.MatchString(email) {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "error": "invalid email format"})
			return
		}

		mxRecords, err := net.LookupMX(parts[1])
		if err != nil || len(mxRecords) == 0 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 21000, "message": "email domain does not exist or has no mail servers"})
			return
		}

		if checkDisposable == "true" || checkDisposable == "1" {
			disposableDomains := []string{
				"tempmail.com",
				"mailinator.com",
				"sharklasers.com",
				"guerrillamail.info",
				"grr.la",
				"guerrillamail.biz",
				"guerrillamail.com",
				"guerrillamail.de",
				"guerrillamail.net",
				"guerrillamail.org",
				"guerrilamailblock.com",
				"pokemail.net",
				"spam4.me",
				"dealexp.org",
			}

			for _, d := range disposableDomains {
				if strings.HasSuffix(parts[1], d) {
					render.DefaultResponder(w, r, render.M{"valid": false, "code": 20010, "error": "disposable email address"})
					return
				}
			}
		}

		render.DefaultResponder(w, r, render.M{"valid": true, "email": email})
	})

	port := "127.0.0.1:3002"

	if _, err := os.Stat("/.dockerenv"); err == nil {
		port = ":3000"
	}
}
