package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"time"
)

// templates holds the parsed HTML templates
var templates *template.Template

func main() {
	var err error
	// Parse HTML templates
	templates, err = template.ParseFiles("templates/index.html")
	if err != nil {
		panic(err)
	}

	// Set up routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/encrypt", encryptHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Configure server with timeouts for security
	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       5 * time.Minute,
	}

	fmt.Println("Server is running on http://localhost:8080")
	// Start the server
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

// homeHandler serves the home page
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	templates.ExecuteTemplate(w, "index.html", nil)
}

// encryptHandler handles the encryption request
func encryptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form data
	err := r.ParseMultipartForm(1 << 20) // 1 MB max memory
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	text := r.FormValue("text")
	namespace := r.FormValue("namespace")
	kubernetesSecret := r.FormValue("kubernetesSecret") == "true"

	// Validate input (basic example, enhance as needed)
	if text == "" || namespace == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	args := []string{"--raw", "--scope", "namespace-wide", "--namespace", namespace, "--cert", "/app/config/cert.pem"}
	if kubernetesSecret {
		args = []string{"--scope", "namespace-wide", "--namespace", namespace, "-o", "yaml", "--cert", "/app/config/cert.pem" }
	}
	
	cmd := exec.Command("kubeseal", args...)
	cmd.Stdin = bytes.NewBufferString(text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Log the error, but don't expose internal details to the client
		fmt.Printf("Error executing kubeseal: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"result": %q}`, string(output))
}