package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Map buat store URL
var shortUrls map[string]string

func init() {
	shortUrls = make(map[string]string)
}
func validateURL(link string) (bool, error) {
	// First, use url.Parse to check if the URL is valid and well-structured
	parsedURL, err := url.Parse(link)
	if err != nil {
		return false, err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false, fmt.Errorf("invalid URL scheme: %s", parsedURL.Scheme)
	}

	return true, nil
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	var longURL string
	var parser map[string]string

	// Read JSON body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error reading request body: %v", err)
		return
	}

	err = json.Unmarshal(body, &parser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON format: %v", err)
		return
	}

	longURL = parser["long_url"]
	validate, _ := validateURL(longURL)
	// Validate long URL
	if !validate {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "URL is not the right format")
		return
	}

	// Generate unique slug
	hash := sha256.Sum256([]byte(longURL))
	slug := base64.StdEncoding.EncodeToString(hash[:])[:6] // ambil 6 byte pertama

	// Check slug exist
	fmt.Println(slug)
	fmt.Println("long URL : ", shortUrls[slug])
	if _, ok := shortUrls[slug]; ok {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Slug already exists")
		return
	}

	// Mapping dan return success
	fmt.Println(longURL)
	shortUrls[slug] = longURL
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "{\"short_url\": \"localhost:8080/%s\"}", slug)
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[1:]
	fmt.Println(r.URL.Path[1:])
	fmt.Println(shortUrls)
	fmt.Println(slug)

	// ambil URL yang dipendekin berdasarkan slug map
	longURL, ok := shortUrls[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 not Found")
		return
	}

	// Redirect to original URL
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func main() {
	http.HandleFunc("/shorten", shortenURL)
	http.HandleFunc("/", redirectURL)
	fmt.Println("Starting server on port 8080...")
	http.ListenAndServe(":8080", nil)
}
