package app

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/repository"
)

var once sync.Once

func (a *fyneApp) cachedIconPath() string {
	if a.Icon() == nil {
		return ""
	}

	dirPath := rootCacheDir(a)
	filePath := filepath.Join(dirPath, "icon.png")
	once.Do(func() {
		err := a.saveIconToCache(dirPath, filePath)
		if err != nil {
			filePath = ""
		}
	})

	return filePath
}

func (a *fyneApp) saveIconToCache(dirPath, filePath string) error {
	err := os.MkdirAll(dirPath, repository.PermUserReadWriteExec)
	if err != nil {
		granio.LogError("Unable to create application cache directory", err)
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		granio.LogError("Unable to create icon file", err)
		return err
	}

	defer file.Close()

	if icon := a.Icon(); icon != nil {
		_, err = file.Write(icon.Content())
		if err != nil {
			granio.LogError("Unable to write icon contents", err)
			return err
		}
	}

	return nil
}
