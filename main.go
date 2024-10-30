package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	defaultPort     = 6419
	defaultHost     = "localhost"
	githubAPIURL    = "https://api.github.com/markdown"
	defaultFileName = "README.md"
)

type Config struct {
	Host     string
	Port     int
	FilePath string
	Token    string
	Browser  bool
}

func renderMarkdown(content []byte, token string) (string, error) {
	payload := map[string]interface{}{
		"text": string(content),
		"mode": "gfm",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", githubAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	rendered, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	return string(rendered), nil
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("Error opening browser: %v", err)
	}
}

func createServer(config Config, tmpl *template.Template) http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(filepath.Dir(config.FilePath)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handleRoot(config, tmpl)(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
	
	return mux
}

func handleRoot(config Config, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := ioutil.ReadFile(config.FilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
			return
		}

		rendered, err := renderMarkdown(content, config.Token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to render markdown: %v", err), http.StatusInternalServerError)
			return
		}

		data := struct {
			Content template.HTML
		}{
			Content: template.HTML(rendered),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	}
}

func main() {
	config := Config{}
	
	flag.StringVar(&config.Host, "host", defaultHost, "Host to listen on")
	flag.IntVar(&config.Port, "port", defaultPort, "Port to listen on")
	flag.StringVar(&config.FilePath, "f", defaultFileName, "File to render")
	flag.StringVar(&config.Token, "token", "", "GitHub personal access token")
	flag.BoolVar(&config.Browser, "b", false, "Open browser automatically")
	flag.Parse()

	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		config.FilePath = filepath.Join(".", config.FilePath)
	}

	var err error
	config.FilePath, err = filepath.Abs(config.FilePath)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	tmpl, err := template.New("grip").Parse(pageTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	server := createServer(config, tmpl)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	url := fmt.Sprintf("http://%s/", addr)
	log.Printf("Running on %s", url)

	if config.Browser {
		go openBrowser(url)
	}

	log.Fatal(http.ListenAndServe(addr, server))
}

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GitHub Readme Preview</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.2.0/github-markdown.min.css">
    <style>
        .markdown-body {
            box-sizing: border-box;
            min-width: 200px;
            max-width: 980px;
            margin: 0 auto;
            padding: 45px;
        }
    </style>
</head>
<body>
    <div class="markdown-body">
        {{.Content}}
    </div>
</body>
</html>`
