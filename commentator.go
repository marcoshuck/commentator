package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/google/go-github/v41/github"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
)

type GitHubConfig struct {
	PrivateKeyPath string `env:"GITHUB_PRIVATE_KEY_FILEPATH,required"`
	ApplicationID  int64  `env:"GITHUB_APP_IDENTIFIER,required"`
	WebhookSecret  string `env:"GITHUB_WEBHOOK_SECRET,required"`
}

func (c *GitHubConfig) Parse() error {
	return env.Parse(c)
}

type Commentator struct {
	config GitHubConfig
	logger *log.Logger
	db     *gorm.DB
}

func (c *Commentator) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", c.Home)
	r.Post("/", c.GitHubWebhook)
	return r
}

func (c *Commentator) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := fmt.Fprint(w, "<h1>Commentator App</h1>")
	if err != nil {
		http.Error(w, "Failed to show home page", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Commentator) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	b := new(bytes.Buffer)
	_, err := io.Copy(b, r.Body)
	if err != nil {
		c.logger.Println("Failed to read request's body:", err)
		http.Error(w, "Failed to process request: Reading body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal payload
	var payload github.PullRequestEvent
	if err = json.Unmarshal(b.Bytes(), &payload); err != nil {
		c.logger.Println("Failed to unmarshal body:", err)
		http.Error(w, "Failed to process request: Webhook payload read", http.StatusInternalServerError)
		return
	}

	transport, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, c.config.ApplicationID, payload.Installation.GetID(), c.config.PrivateKeyPath)
	if err != nil {
		c.logger.Println("Failed to initialize key from file:", err)
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		return
	}
	client := github.NewClient(&http.Client{
		Transport: transport,
	})
	status := "queued"
	_, res, err := client.Checks.CreateCheckRun(r.Context(), payload.GetRepo().GetOwner().GetLogin(), payload.GetRepo().GetName(), github.CreateCheckRunOptions{
		Name:        "Commentator CI",
		HeadSHA:     payload.GetPullRequest().GetHead().GetSHA(),
		DetailsURL:  nil,
		ExternalID:  nil,
		Status:      &status,
		Conclusion:  nil,
		StartedAt:   nil,
		CompletedAt: nil,
		Output:      nil,
		Actions:     nil,
	})
	if err != nil {
		c.logger.Println("Failed to post comment on Pull Request:", err)
		http.Error(w, "Failed to post comment", http.StatusInternalServerError)
		return
	}

	if res.StatusCode > 299 {
		c.logger.Println("Failed to post comment on Pull Request:", err)
		http.Error(w, "Failed to post comment", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func NewCommentator(logger *log.Logger, cfg GitHubConfig) *Commentator {
	return &Commentator{
		logger: logger,
		config: cfg,
	}
}
