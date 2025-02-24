package util

import (
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const ianaTLDURL = "https://data.iana.org/TLD/tlds-alpha-by-domain.txt"

var (
	domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)
	tldCache    = make(map[string]bool)
	cacheMutex  sync.Mutex
	cacheExpiry time.Time
)

func fetchTLDs() (map[string]bool, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if time.Now().Before(cacheExpiry) {

		return tldCache, nil
	}

	resp, err := http.Get(ianaTLDURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(body), "\n")
	tlds := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && !strings.HasPrefix(line, "#") {
			tlds["."+strings.ToLower(line)] = true
		}
	}

	tldCache = tlds
	cacheExpiry = time.Now().Add(time.Hour)

	return tldCache, nil
}

func IsValidDomain(domain string) bool {
	validTLDs, _ := fetchTLDs()

	if !domainRegex.MatchString(domain) {
		return false
	}
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return false
	}
	tld := "." + strings.ToLower(domainParts[len(domainParts)-1])
	return validTLDs[tld]
}
