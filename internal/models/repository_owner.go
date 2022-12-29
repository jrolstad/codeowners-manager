package models

type RepositoryOwner struct {
	Host         string
	Organization string
	Repository   string
	Pattern      string
	Owners       []string
	Parent       string
}
