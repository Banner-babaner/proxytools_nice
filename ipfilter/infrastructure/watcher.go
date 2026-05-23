package infrastructure

import (

	"github.com/Banner-babaner/proxytools_nice/config"
	"github.com/Banner-babaner/proxytools_nice/ipfilter/entity"
	"github.com/Banner-babaner/proxytools_nice/logger"
	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	path string
}

func NewFileWatcher(path string) *FileWatcher {
	return &FileWatcher{path: path}
}

func (fw *FileWatcher) Watch(callback func(entity.ListsConfig)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := watcher.Add(fw.path); err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.Info("Config file changed, reloading IP lists")
					cfg, err := config.Load(fw.path)
					if err != nil {
						logger.Error("Failed to reload config", err)
						continue
					}
					callback(entity.ListsConfig{
						Whitelist: cfg.IPFilter.Lists.Whitelist,
						Blacklist: cfg.IPFilter.Lists.Blacklist,
						Graylist:  cfg.IPFilter.Lists.Graylist,
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error("Watcher error", err)
			}
		}
	}()

	return nil
}