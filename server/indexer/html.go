// SPDX-License-Identifier: AGPL-3.0-or-later

package indexer

import (
	"unicode/utf8"

	"github.com/asciimoo/hister/server/document"
	"github.com/asciimoo/hister/server/extractor"
)

type htmlFileType struct{}

func (htmlFileType) Match(path string) bool {
	return hasExtension(path, ".html", ".htm")
}

func (htmlFileType) Prepare(d *document.Document, content []byte) error {
	if !utf8.Valid(content) {
		return ErrBinaryFile
	}
	d.HTML = string(content)
	// Reuse web extraction, leaving file metadata, sensitive content checks,
	// and language detection to normal document processing.
	return extractor.Extract(d)
}
