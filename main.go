package main

import (
	"1433/email-validator/util"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func main() {
	r := chi.NewRouter()

	go util.ConnectCache()

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.DefaultResponder(w, r, render.M{
			"message":      "This service takes a very brief look to validate emails and makes sure they are valid.",
			"route_to_hit": "/api/check?email=<email>",
			"error_codes": map[int]string{
				20000: "Email query string not given",
				20001: "Invalid email format",
				20002: "Email must be from 8 to 254 characters",
				20003: "Invalid domain tld",
				20010: "disposable email address (query string: check_disposable=1 to check for such)",
				20011: "Email domain does not exist or has no mail servers",
			},
		})
	})

	r.Get("/api/check", func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		checkDisposable := strings.TrimSpace(r.URL.Query().Get("check_disposable"))

		if email == "" {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20000, "message": "Email query string not given"})
			return
		}

		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "message": "Invalid email format"})
			return
		}

		if len(email) < 8 || len(email) > 254 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20002, "message": "Email must be from 8 to 254 characters"})
			return
		}

		re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !re.MatchString(email) {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "message": "Invalid email format"})
			return
		}

		if valid := util.IsValidDomain(parts[1]); !valid {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20003, "message": "Invalid domain tld"})
			return
		}

		mxRecords, err := net.LookupMX(parts[1])
		if err != nil || len(mxRecords) == 0 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20011, "message": "Email domain does not exist or has no mail servers"})
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
					render.DefaultResponder(w, r, render.M{"valid": false, "code": 20010, "message": "disposable email address"})
					return
				}
			}
		}

		render.DefaultResponder(w, r, render.M{"valid": true, "email": email})
	})

	host := "127.0.0.1"
	port := "3002"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
		host = "0.0.0.0"
	}

	address := host + ":" + port

	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatal(err)
	}
}
