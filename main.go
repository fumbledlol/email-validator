package main

import (
	"1433/email-validator/util"
	"bufio"
	"net"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strings"

	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

var (
	disposableDomains = make(map[string]bool)
	domainLabelRegex  = regexp.MustCompile(`^[A-Za-z0-9-]{1,63}$`)
)

func main() {
	file, err := os.Open("disposable_emails.txt")
	if err != nil {
		log.Printf("Warning: could not open disposable_emails.txt: %v. Continuing without disposable checks.", err)
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain != "" {
				disposableDomains[domain] = true
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Warning: error reading disposable_emails.txt: %v", err)
		}
	}

	r := chi.NewRouter()

	go util.ConnectCache()

	r.Use(middleware.Logger)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	/*r.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
	})*/

	r.Get("/api/check", func(w http.ResponseWriter, r *http.Request) {
		email := strings.TrimSpace(r.URL.Query().Get("email"))
		checkDisposable := strings.TrimSpace(r.URL.Query().Get("check_disposable"))

		if email == "" {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20000, "message": "Email query string not given"})
			return
		}

		_, err := mail.ParseAddress(email)
		if err != nil {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "message": "Invalid email format"})
			return
		}

		_, domain, _ := strings.Cut(email, "@")

		if len(email) < 7 || len(email) > 254 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20002, "message": "Email must be from 7 to 254 characters"})
			return
		}

		domain = strings.TrimSuffix(strings.ToLower(domain), ".")
		labels := strings.Split(domain, ".")
		if len(labels) < 2 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "message": "Invalid email format"})
			return
		}

		for _, l := range labels {
			if l == "" || strings.HasPrefix(l, "-") || strings.HasSuffix(l, "-") || !domainLabelRegex.MatchString(l) {
				render.DefaultResponder(w, r, render.M{"valid": false, "code": 20001, "message": "Invalid email format"})
				return
			}
		}

		if valid := util.IsValidDomain(domain); !valid {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20003, "message": "Invalid domain tld"})
			return
		}

		mxRecords, err := net.LookupMX(domain)
		if err != nil || len(mxRecords) == 0 {
			render.DefaultResponder(w, r, render.M{"valid": false, "code": 20011, "message": "Email domain does not exist or has no mail servers"})
			return
		}

		if checkDisposable == "true" || checkDisposable == "1" {
			for d := range disposableDomains {
				if strings.HasSuffix(domain, d) {
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
