package hexo

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Post struct {
	ID        string
	Title     string
	Date      string
	Updated   string
	Category  string
	Tags      []string
	BannerImg string
	IndexImg  string
	Content   string
	Math      bool
	Mermaid   bool
	Archive   bool
}

const postTemplate = `---
title: {{.Title}}
date: {{.Date}}
updated: {{.Updated}}
categories:
  - {{.Category}}
tags:
{{- range .Tags}}
  - {{.}}
{{- end}}
banner_img: {{.BannerImg}}
index_img: {{.IndexImg}}
math: true
mermaid: true
archive: {{.Archive}}
---

{{.Content}}
`

var (
	codeBlockRe = regexp.MustCompile(`(?s)(~{3,}.*?~{3,}|\x60{3,}.*?\x60{3,}|\x60.*?\x60)`)
	newlineRe   = regexp.MustCompile(`([^\x00]|^)(\\n|\\\n)`)
)

func unescapeText(text string) string {
	// Outline sends some escaped characters in the text, we need to unescape them before processing.
	// However, characters in Markdown code blocks should be treated as is.
	lastEnd := 0
	var result strings.Builder
	matches := codeBlockRe.FindAllStringIndex(text, -1)

	for _, match := range matches {
		start, end := match[0], match[1]
		outsideText := text[lastEnd:start]
		result.WriteString(renderNewline(outsideText))
		result.WriteString(text[start:end])
		lastEnd = end
	}

	result.WriteString(renderNewline(text[lastEnd:]))
	return result.String()
}

func renderNewline(text string) string {
	text = strings.ReplaceAll(text, "\\\\", "\x00")

	re := regexp.MustCompile(`([^\x00]|^)(\\n|\\\n)`)
	text = re.ReplaceAllString(text, "$1\n")

	text = strings.ReplaceAll(text, "\x00", "\\\\")
	return text
}

func renderPost(post *Post) (string, error) {
	post.Content = unescapeText(post.Content)

	originalLines := strings.Split(post.Content, "\n")
	var Lines []string
	for i := 0; i < len(originalLines); i++ {
		if strings.TrimSpace(originalLines[i]) == "[REMOVED]" {
			if i+1 < len(originalLines) && strings.TrimSpace(originalLines[i+1]) == "" {
				i++
			}
			continue
		}
		Lines = append(Lines, originalLines[i])
	}
	post.Content = strings.Join(Lines, "\n")

	tmpl, err := template.New("post").Parse(postTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, post); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func CreateHexoPost(dir string, post *Post) error {
	content, err := renderPost(post)
	if err != nil {
		return err
	}
	filePath := filepath.Join(dir, post.ID+".md")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}
	log.Printf("Hexo post created at %s", filePath)
	return nil
}

func RemoveHexoPost(dir string, ID string) error {
	filePath := filepath.Join(dir, ID+".md")
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	log.Printf("Hexo post removed at %s", filePath)
	return nil
}
