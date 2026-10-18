//go:build !android && !ios && !mobile && !wasm && !test_web_driver && !tamago && !noos && !tinygo

package app

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/repository"
)

func watchFileAddTarget(watcher *fsnotify.Watcher, path string) {
	dir := filepath.Dir(path)
	ensureDirExists(dir)

	err := watcher.Add(dir)
	if err != nil {
		granio.LogError("Settings watch error:", err)
	}
}

func ensureDirExists(dir string) {
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return
	}

	err := os.MkdirAll(dir, repository.PermUserReadWriteExec)
	if err != nil {
		granio.LogError("Unable to create settings storage:", err)
	}
}

func watchFile(path string, callback func()) *fsnotify.Watcher {
	path = filepath.Clean(path)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		granio.LogError("Failed to watch settings file:", err)
		return nil
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// The directory watch survives deletion and atomic replacement of
				// the file. Events for neighboring files do not change this target.
				if filepath.Clean(event.Name) == path &&
					event.Op.Has(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) {
					granio.Do(callback)
				}
			case watchErr, ok := <-watcher.Errors:
				if !ok {
					return
				}
				granio.LogError("Settings watch error:", watchErr)
			}
		}
	}()

	watchFileAddTarget(watcher, path)
	return watcher
}

func (s *settings) watchSettings() {
	watcher := watchFile(s.schema.StoragePath(), s.fileChanged)
	s.appearanceMu.Lock()
	stopped := s.appearanceStopped
	if !stopped {
		s.watcher = watcher
	}
	s.appearanceMu.Unlock()
	if stopped {
		if watcher != nil {
			watcher.Close()
		}
		return
	}
	s.startSystemAppearance()

	if s.explicitThemeVariantName() == "" {
		a := granio.CurrentApp()
		if a != nil && s != nil && a.Settings() == s { // ignore if testing
			watchTheme(s)
		}
	}
}

func (s *settings) stopWatching() {
	s.stopSystemAppearance()
	s.appearanceMu.Lock()
	watcher := s.watcher
	s.watcher = nil
	s.appearanceMu.Unlock()
	if watcher == nil {
		return
	}

	watcher.(*fsnotify.Watcher).Close() // fsnotify returns false positives, see https://github.com/fsnotify/fsnotify/issues/268
}
