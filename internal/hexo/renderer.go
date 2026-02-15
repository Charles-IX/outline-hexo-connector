package hexo

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
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
---

{{.Content}}
`

func unescapeText(text string) string {
	// Outline sends some escaped characters in the text, we need to unescape them before processing.
	// I tried to write `text = strings.ReplaceAll(text, "\\\\", "\\")` but failed:
	// ReplaceAll can only be used on strings, where '\n' is treated as literal letters.
	// '\\' can not be replaced by '\', for the former can not be simply searched (escaped) in the text.
	// So here is some AI generated masterpiece:
	text = strings.ReplaceAll(text, "\\\\", "\x00")

	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\\r", "\r")
	text = strings.ReplaceAll(text, "\\t", "\t")

	text = strings.ReplaceAll(text, "\\+", "+")
	text = strings.ReplaceAll(text, "\\-", "-")
	text = strings.ReplaceAll(text, "\\*", "*")
	text = strings.ReplaceAll(text, "\\#", "#")
	text = strings.ReplaceAll(text, "\\>", ">")
	text = strings.ReplaceAll(text, "\\|", "|")

	text = strings.ReplaceAll(text, "\x00", "\\")
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
