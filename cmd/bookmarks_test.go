// SPDX-License-Identifier: AGPL-3.0-or-later

package cmd

import (
	"strings"
	"testing"
)

func TestResolveBookmarkStoresNamedDB(t *testing.T) {
	got, err := resolveBookmarkStores("", "/tmp/profile/places.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].path != "/tmp/profile/places.sqlite" {
		t.Fatalf("firefox store = %#v", got)
	}
	if _, ok := got[0].source.(firefoxBookmarkSource); !ok {
		t.Fatalf("source = %T, want firefoxBookmarkSource", got[0].source)
	}

	got, err = resolveBookmarkStores("chrome", "/tmp/Default/Bookmarks")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got[0].source.(chromiumBookmarkSource); !ok {
		t.Fatalf("source = %T, want chromiumBookmarkSource", got[0].source)
	}

	got, err = resolveBookmarkStores("", "/tmp/Ladybird/Bookmarks.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got[0].source.(ladybirdBookmarkSource); !ok {
		t.Fatalf("source = %T, want ladybirdBookmarkSource", got[0].source)
	}
}

func TestResolveBookmarkStoresRejectsUnknown(t *testing.T) {
	if _, err := resolveBookmarkStores("safari", ""); err == nil {
		t.Fatal("expected safari to be rejected")
	}
	if _, err := resolveBookmarkStores("", "/tmp/History"); err == nil {
		t.Fatal("expected a history path to be rejected")
	}
}

func TestBookmarkSourceHasName(t *testing.T) {
	tests := []struct {
		src  bookmarkSource
		name string
		want bool
	}{
		{firefoxBookmarkSource{}, "firefox", true},
		{firefoxBookmarkSource{}, "fire", true},
		{firefoxBookmarkSource{}, "firefoxx", false},
		{firefoxBookmarkSource{}, "chrome", false},
		{chromiumBookmarkSource{}, "chrome", true},
		{chromiumBookmarkSource{}, "chrom", true},
	}
	for _, tt := range tests {
		got := bookmarkSourceHasName(tt.src, tt.name)
		if got != tt.want {
			t.Errorf("bookmarkSourceHasName(%T, %q) = %v, want %v", tt.src, tt.name, got, tt.want)
		}
	}
}

func TestResolveBookmarkStoresRejectsTypoBrowser(t *testing.T) {
	_, err := resolveBookmarkStores("firefoxx", "")
	if err == nil {
		t.Fatal("expected firefoxx to be rejected as unknown")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("error = %v, want unknown browser", err)
	}
}

func TestExcludeURLGroups(t *testing.T) {
	groups := []urlImportGroup{
		{name: "firefox", path: "/ff"},
		{name: "chrome", path: "/ch"},
		{name: "ladybird", path: "/lb"},
	}

	got := excludeURLGroups(groups, nil)
	if len(got) != 3 {
		t.Fatalf("empty tokens = %#v, want all groups", got)
	}

	got = excludeURLGroups(groups, []string{})
	if len(got) != 3 {
		t.Fatalf("empty slice = %#v, want all groups", got)
	}

	got = excludeURLGroups(groups, []string{"1"})
	if len(got) != 2 || got[0].name != "firefox" || got[1].name != "ladybird" {
		t.Fatalf("exclude by index = %#v", got)
	}

	got = excludeURLGroups(groups, []string{"chrome"})
	if len(got) != 2 || got[0].name != "firefox" || got[1].name != "ladybird" {
		t.Fatalf("exclude by name = %#v", got)
	}
}

func TestExcludeURLGroupsCRLF(t *testing.T) {
	groups := []urlImportGroup{
		{name: "firefox", path: "/ff"},
		{name: "chrome", path: "/ch"},
	}
	// promptExcludeURLGroups uses strings.Fields, which strips the trailing \r from CRLF input.
	tokens := strings.Fields("1\r\n")
	got := excludeURLGroups(groups, tokens)
	if len(got) != 1 || got[0].name != "firefox" {
		t.Fatalf("CRLF exclude by index = %#v", got)
	}
}
