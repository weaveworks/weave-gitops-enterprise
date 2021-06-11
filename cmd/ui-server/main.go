package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type paramSet struct {
	port             int
	gitDeployKeyFile string
	gitURL           string
	gitPollInterval  time.Duration
	htmlRootPath     string
}

type spaFileSystem struct {
	root http.FileSystem
}

func (fs *spaFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("index.html")
	}
	return f, err
}

func main() {
	var params paramSet
	flag.IntVar(&params.port, "port", 80, "")
	flag.StringVar(&params.gitDeployKeyFile, "git-deploy-key-file", "private-key", "")
	flag.StringVar(&params.gitURL, "git-url", "git@github.com:weaveworks/test-wkp.git", "")
	flag.DurationVar(&params.gitPollInterval, "git-poll-interval", 15*time.Second, "")
	flag.StringVar(&params.htmlRootPath, "html-root-path", "html/wkp-ui", "")
	flag.Parse()

	// Serve the UI
	fs := http.FileServer(&spaFileSystem{http.Dir(params.htmlRootPath)})
	http.Handle("/", http.StripPrefix("/", fs))

	log.Println("Listening on", params.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", params.port), nil))
}
