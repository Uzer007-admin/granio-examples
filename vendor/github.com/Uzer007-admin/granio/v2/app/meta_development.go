package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/internal/metadata"
)

func checkLocalMetadata() {
	if build.NoMetadata || build.Mode == granio.BuildRelease {
		return
	}

	dir := getProjectPath()
	file := filepath.Join(dir, "FyneApp.toml")
	ref, err := os.Open(file)
	if err != nil { // no worries, this is just an optional fallback
		return
	}
	defer ref.Close()

	data, err := metadata.Load(ref)
	if err != nil || data == nil {
		granio.LogError("failed to parse FyneApp.toml", err)
		return
	}

	meta.ID = data.Details.ID
	meta.Name = data.Details.Name
	meta.Version = data.Details.Version
	meta.Build = data.Details.Build

	const iconBaseSize = 512
	if data.Details.Icon != "" {
		res, err := granio.LoadResourceFromPath(data.Details.Icon)
		if err == nil {
			meta.Icon = metadata.ScaleIcon(res, iconBaseSize)
		}
	} else { // Icon.png fallback
		res, err := granio.LoadResourceFromPath("Icon.png")
		if err == nil {
			meta.Icon = metadata.ScaleIcon(res, iconBaseSize)
		}
	}

	meta.Release = false
	meta.Custom = data.Development
	meta.Migrations = data.Migrations
}

func getProjectPath() string {
	exe, err := os.Executable()
	work, _ := os.Getwd()

	if err != nil {
		granio.LogError("failed to lookup build executable", err)
		return work
	}

	temp := os.TempDir()
	if strings.Contains(exe, temp) || strings.Contains(exe, "go-build") { // this happens with "go run"
		return work
	}

	// we were called with an executable from "go build"
	return filepath.Dir(exe)
}
