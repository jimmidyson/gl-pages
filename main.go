package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
)

type PushEvent struct {
	Before     string `json:"before"`
	After      string `json:"after"`
	Ref        string `json:"ref" binding:"required"`
	UserId     int    `json:"user_id"`
	UserName   string `json:"user_name"`
	ProjectId  int    `json:"project_id"`
	Repository struct {
		Name        string `json:"name"`
		Url         string `json:"url" binding:"required"`
		Description string `json:"description"`
		Homepage    string `json:"homepage" binding:"required"`
	} `json:"repository"`
	Commits []struct {
		Id        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		Url       string    `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commits"`
	TotalCommitsCount int `json:"total_commits_count"`
}

func main() {
	var (
		gitPath = flag.String("git", "git", "Path to git executable")
		rootDir = flag.String("root", ".", "Root directory to serve docs from")

		validGitUrl = regexp.MustCompile(`^[^\:]+:(.+)\.git$`)
	)

	flag.Parse()

	m := martini.Classic()
	m.Post("/", binding.Json(PushEvent{}), func(event PushEvent) {
		if event.Ref != "refs/heads/gl-pages" {
			return
		}
		match := validGitUrl.FindStringSubmatch(event.Repository.Url)
		if match != nil {
			repoPath := fmt.Sprintf("%s/%s", *rootDir, match[1])
			var cmd *exec.Cmd
			if _, err := os.Stat(repoPath); os.IsNotExist(err) {
				if err := os.MkdirAll(repoPath, 0755); err != nil {
					return
				}
				cmd = exec.Command(*gitPath, "clone", "-b", "gl-pages", "--single-branch", event.Repository.Url, repoPath)
			} else {
				cmd = exec.Command(*gitPath, "pull")
				cmd.Dir = repoPath
			}
			out, _ := cmd.CombinedOutput()
			os.Stdout.Write(out)
		}
	})
	m.Run()
}
