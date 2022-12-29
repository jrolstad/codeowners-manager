package models

import "time"

type RepositoryOwnerData struct {
	Id           string
	Host         string
	Organization string
	Repository   string
	Pattern      string
	Owners       []string
	Parent       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}
