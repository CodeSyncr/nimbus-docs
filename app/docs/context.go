package docs

import (
	"regexp"
	"strings"
)

// DocsContext holds the full Nimbus documentation as plain text for AI context.
var DocsContext string

// SetDocsContext sets the documentation context (called from init).
func SetDocsContext(s string) {
	DocsContext = s
}

// GetDocsContext returns the full docs context for the AI.
func GetDocsContext() string {
	return DocsContext
}

// ExtractTextFromHTML strips HTML from .nimbus content and preserves structure.
func ExtractTextFromHTML(raw string) string {
	// Remove @layout directive
	layoutRe := regexp.MustCompile(`@layout\([^)]+\)\s*\n?`)
	raw = layoutRe.ReplaceAllString(raw, "")

	// Extract code blocks first (preserve them)
	codeBlockRe := regexp.MustCompile(`(?s)<pre[^>]*><code[^>]*>(.*?)</code></pre>`)
	codeBlocks := codeBlockRe.FindAllStringSubmatch(raw, -1)
	raw = codeBlockRe.ReplaceAllString(raw, "\n\n[CODE_BLOCK]\n\n")

	// Replace code blocks back with markdown-style
	for _, m := range codeBlocks {
		if len(m) > 1 {
			code := strings.TrimSpace(m[1])
			code = strings.ReplaceAll(code, "&lt;", "<")
			code = strings.ReplaceAll(code, "&gt;", ">")
			code = strings.ReplaceAll(code, "&amp;", "&")
			raw = strings.Replace(raw, "\n\n[CODE_BLOCK]\n\n", "\n\n```go\n"+code+"\n```\n\n", 1)
		}
	}
	raw = strings.ReplaceAll(raw, "[CODE_BLOCK]", "```")

	// Strip remaining HTML tags
	tagRe := regexp.MustCompile(`<[^>]+>`)
	raw = tagRe.ReplaceAllString(raw, " ")

	// Decode common entities
	raw = strings.ReplaceAll(raw, "&amp;", "&")
	raw = strings.ReplaceAll(raw, "&lt;", "<")
	raw = strings.ReplaceAll(raw, "&gt;", ">")
	raw = strings.ReplaceAll(raw, "&quot;", "\"")

	// Collapse multiple newlines and spaces
	raw = regexp.MustCompile(`\n\s*\n\s*\n+`).ReplaceAllString(raw, "\n\n")
	raw = regexp.MustCompile(`[ \t]+`).ReplaceAllString(raw, " ")
	raw = strings.TrimSpace(raw)

	return raw
}
