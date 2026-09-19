// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// bookmarkSource reads one bookmark format from browser profiles or an explicit path.
type bookmarkSource interface {
	names() []string
	accepts(path string) bool
	detect(profiles []browserDB) []bookmarkStore
	listURLs(path string) ([]string, error)
}

type bookmarkStore struct {
	browser string
	path    string
	source  bookmarkSource
}

func bookmarkSources() []bookmarkSource {
	return []bookmarkSource{
		firefoxBookmarkSource{},
		chromiumBookmarkSource{},
		ladybirdBookmarkSource{},
	}
}

func resolveBookmarkStores(browser, dbPath string) ([]bookmarkStore, error) {
	browser = strings.ToLower(strings.TrimSpace(browser))
	if dbPath != "" {
		src := bookmarkSourceAccepting(dbPath)
		if src == nil {
			return nil, fmt.Errorf("unsupported bookmark --db %s", dbPath)
		}
		if browser != "" && !bookmarkSourceHasName(src, browser) {
			return nil, fmt.Errorf("--browser %s does not match --db %s", browser, dbPath)
		}
		name := browser
		if name == "" {
			name = src.names()[0]
		}
		return []bookmarkStore{{browser: name, path: dbPath, source: src}}, nil
	}
	if browser != "" && bookmarkSourceByName(browser) == nil {
		return nil, fmt.Errorf("unknown --browser %s", browser)
	}

	var profiles []browserDB
	for _, db := range getDBPaths() {
		if browser == "" || strings.HasPrefix(strings.ToLower(db.name), browser) {
			profiles = append(profiles, db)
		}
	}
	var stores []bookmarkStore
	for _, src := range bookmarkSources() {
		if browser != "" && !bookmarkSourceHasName(src, browser) {
			continue
		}
		stores = append(stores, src.detect(profiles)...)
	}
	if len(stores) == 0 {
		if browser != "" {
			return nil, fmt.Errorf("no bookmark store found for browser %s", browser)
		}
		return nil, fmt.Errorf("no bookmark stores found")
	}
	return stores, nil
}

func bookmarkSourceAccepting(path string) bookmarkSource {
	for _, src := range bookmarkSources() {
		if src.accepts(path) {
			return src
		}
	}
	return nil
}

func bookmarkSourceByName(name string) bookmarkSource {
	for _, src := range bookmarkSources() {
		if bookmarkSourceHasName(src, name) {
			return src
		}
	}
	return nil
}

func bookmarkSourceHasName(src bookmarkSource, name string) bool {
	name = strings.ToLower(name)
	for _, n := range src.names() {
		if n == name || strings.HasPrefix(n, name) {
			return true
		}
	}
	return false
}

func uniqueBookmarkURLs(urls []string) []string {
	seen := make(map[string]struct{}, len(urls))
	var out []string
	for _, u := range urls {
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func bookmarkFileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func bookmarkSiblingFile(path, name string) string {
	return filepath.Join(filepath.Dir(path), name)
}
