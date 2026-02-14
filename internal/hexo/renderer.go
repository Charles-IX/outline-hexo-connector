package hexo

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
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

func renderPost(post *Post) (string, error) {
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
