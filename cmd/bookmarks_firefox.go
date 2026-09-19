// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const firefoxBookmarkURLsQuery = "SELECT DISTINCT p.url FROM moz_bookmarks b JOIN moz_places p ON p.id = b.fk WHERE b.type = 1 AND (p.url LIKE 'http://%' OR p.url LIKE 'https://%')"

type firefoxBookmarkSource struct{}

func (firefoxBookmarkSource) names() []string {
	return []string{"firefox", "zen", "waterfox"}
}

func (firefoxBookmarkSource) accepts(path string) bool {
	return strings.HasSuffix(path, "places.sqlite")
}

func (s firefoxBookmarkSource) detect(profiles []browserDB) []bookmarkStore {
	var stores []bookmarkStore
	for _, db := range profiles {
		if db.table_name != "moz_places" {
			continue
		}
		for _, path := range db.paths {
			stores = append(stores, bookmarkStore{
				browser: strings.ToLower(db.name),
				path:    path,
				source:  s,
			})
		}
	}
	return stores
}

func (firefoxBookmarkSource) listURLs(path string) ([]string, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?immutable=1&mode=ro", path))
	if err != nil {
		return nil, fmt.Errorf("open firefox places: %w", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query(firefoxBookmarkURLsQuery)
	if err != nil {
		return nil, fmt.Errorf("query firefox bookmarks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return uniqueBookmarkURLs(urls), nil
}
