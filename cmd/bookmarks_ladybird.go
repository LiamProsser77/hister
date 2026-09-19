// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ladybirdBookmarkSource struct{}

type ladybirdBookmarkItem struct {
	Type     string                 `json:"type"`
	URL      string                 `json:"url"`
	Title    string                 `json:"title"`
	Children []ladybirdBookmarkItem `json:"children"`
}

type ladybirdBookmarksFile struct {
	Items []ladybirdBookmarkItem `json:"items"`
}

func (ladybirdBookmarkSource) names() []string {
	return []string{"ladybird"}
}

func (ladybirdBookmarkSource) accepts(path string) bool {
	return strings.HasSuffix(path, "Bookmarks.json")
}

func (s ladybirdBookmarkSource) detect(profiles []browserDB) []bookmarkStore {
	var stores []bookmarkStore
	seen := map[string]struct{}{}
	add := func(path string) {
		if !bookmarkFileExists(path) {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		stores = append(stores, bookmarkStore{browser: "ladybird", path: path, source: s})
	}

	for _, db := range profiles {
		if db.table_name != "History" {
			continue
		}
		for _, path := range db.paths {
			add(bookmarkSiblingFile(path, "Bookmarks.json"))
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return stores
	}
	for _, cand := range []string{
		filepath.Join(home, ".config", "Ladybird", "Bookmarks.json"),
		filepath.Join(home, ".local", "share", "Ladybird", "Bookmarks.json"),
		filepath.Join(home, "Library", "Application Support", "Ladybird", "Bookmarks.json"),
		filepath.Join(home, ".var", "app", "org.ladybird.Ladybird", "config", "Ladybird", "Bookmarks.json"),
	} {
		add(cand)
	}
	for _, pattern := range []string{
		filepath.Join(home, ".config", "Ladybird", "Profiles", "*", "Bookmarks.json"),
		filepath.Join(home, ".local", "share", "Ladybird", "Profiles", "*", "Bookmarks.json"),
		filepath.Join(home, "Library", "Application Support", "Ladybird", "Profiles", "*", "Bookmarks.json"),
		filepath.Join(home, ".var", "app", "org.ladybird.Ladybird", "config", "Ladybird", "Profiles", "*", "Bookmarks.json"),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, path := range matches {
			add(path)
		}
	}
	return stores
}

func (ladybirdBookmarkSource) listURLs(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ladybird bookmarks: %w", err)
	}
	var file ladybirdBookmarksFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("parse ladybird bookmarks: %w", err)
	}
	return uniqueBookmarkURLs(collectLadybirdBookmarkURLs(file.Items)), nil
}

func collectLadybirdBookmarkURLs(items []ladybirdBookmarkItem) []string {
	var urls []string
	for _, item := range items {
		if strings.EqualFold(item.Type, "bookmark") {
			urls = append(urls, item.URL)
			continue
		}
		urls = append(urls, collectLadybirdBookmarkURLs(item.Children)...)
	}
	return urls
}
