package app

import (
	"github.com/Uzer007-admin/granio/v2"
)

var meta = granio.AppMetadata{
	ID:         "",
	Name:       "",
	Version:    "0.0.1",
	Build:      1,
	Release:    false,
	Custom:     map[string]string{},
	Migrations: map[string]bool{},
}

// SetMetadata overrides the packaged application metadata.
// This data can be used in many places like notifications and about screens.
func SetMetadata(m granio.AppMetadata) {
	meta = m

	if meta.Custom == nil {
		meta.Custom = map[string]string{}
	}
	if meta.Migrations == nil {
		meta.Migrations = map[string]bool{}
	}
}

func (*fyneApp) Metadata() granio.AppMetadata {
	if meta.ID == "" && meta.Name == "" {
		checkLocalMetadata()
	}

	return meta
}
