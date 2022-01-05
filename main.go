package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[Commentator] ", log.LstdFlags|log.Lshortfile|log.Lmsgprefix)

	var cfg GitHubConfig
	if err := cfg.Parse(); err != nil {
		logger.Fatalln("Failed to parse github config:", err)
	}

	c := NewCommentator(logger, cfg)
	s := NewServer(logger, c)

	if err := s.ListenAndServe(); err != nil {
		logger.Fatalln("Failed to listen on port 8080:", err)
	}
}
