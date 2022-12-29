package models

type Host struct {
	Id                     string
	Name                   string
	BaseUrl                string
	Type                   string
	SubType                string
	AuthenticationType     string
	ClientSecretName       string
	ParentOwnerLinePattern string
}
