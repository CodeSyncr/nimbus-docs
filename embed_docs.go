package main

import (
	"embed"
	"sort"
	"strings"

	"nimbus-starter/app/docs"
)

//go:embed resources/views/docs/*.nimbus
var docsFS embed.FS

func init() {
	entries, _ := docsFS.ReadDir("resources/views/docs")
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".nimbus") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		data, err := docsFS.ReadFile("resources/views/docs/" + name)
		if err != nil {
			continue
		}
		slug := strings.TrimSuffix(name, ".nimbus")
		text := docs.ExtractTextFromHTML(string(data))
		if text == "" {
			continue
		}
		b.WriteString("\n\n--- DOC: ")
		b.WriteString(slug)
		b.WriteString(" ---\n\n")
		b.WriteString(text)
	}
	docs.SetDocsContext(b.String())
}
