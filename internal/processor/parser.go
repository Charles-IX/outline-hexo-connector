package processor

import (
	"fmt"
	"regexp"
	"strings"
)

type MetadataAndText struct {
	BannerImg string
	IndexImg  string
	Tags      []string
	Text      string
	Archive   bool
}

func ExtractMetadataAndText(text string) *MetadataAndText {
	metadataAndText := &MetadataAndText{
		Text: text,
	}

	// We want picture which alt text is banner_img or index_img or banner_index_img
	reBannerAndIndex := regexp.MustCompile(`!\[(?:banner_index_img|index_banner_img)\]\((.*?)\)`)
	if match := reBannerAndIndex.FindStringSubmatch(text); len(match) > 1 {
		metadataAndText.BannerImg = match[1]
		metadataAndText.IndexImg = match[1]
		metadataAndText.Text = reBannerAndIndex.ReplaceAllString(metadataAndText.Text, "[REMOVED]")
	}

	reBanner := regexp.MustCompile(`!\[banner_img\]\((.*?)\)`)
	if match := reBanner.FindStringSubmatch(text); len(match) > 1 {
		metadataAndText.BannerImg = match[1]
		metadataAndText.Text = reBanner.ReplaceAllString(metadataAndText.Text, "[REMOVED]")
	}

	reIndex := regexp.MustCompile(`!\[index_img\]\((.*?)\)`)
	if match := reIndex.FindStringSubmatch(text); len(match) > 1 {
		metadataAndText.IndexImg = match[1]
		metadataAndText.Text = reIndex.ReplaceAllString(metadataAndText.Text, "[REMOVED]")
	}

	reMore := regexp.MustCompile(`(?m)^\\?\+>\s*More:.*$`)
	metadataAndText.Text = reMore.ReplaceAllString(metadataAndText.Text, "<!-- more -->")

	// Extract tags from the line starting with "+> Tags: "
	reTags := regexp.MustCompile(`(?m)^\\?\+>\s*Tags:\s*(.*)$`)
	if match := reTags.FindStringSubmatch(text); len(match) > 1 {
		tagStr := match[1]

		rawTags := strings.FieldsFunc(tagStr, func(r rune) bool {
			return r == ',' || r == '，'
		})
		for _, t := range rawTags {
			metadataAndText.Tags = append(metadataAndText.Tags, strings.TrimSpace(t))
		}

		metadataAndText.Text = reTags.ReplaceAllString(metadataAndText.Text, "[REMOVED]")
	}

	reArchive := regexp.MustCompile(`(?m)^\\?\+>\s*Archived\s*$`)
	if reArchive.MatchString(metadataAndText.Text) {
		metadataAndText.Archive = true
		metadataAndText.Text = reArchive.ReplaceAllString(metadataAndText.Text, "[REMOVED]")
	}

	reAdmonition := regexp.MustCompile(`(?ms)^:::([a-zA-Z0-9_-]+)[ \t]*\n(.*?)\n:::[ \t]*$`)
	metadataAndText.Text = reAdmonition.ReplaceAllStringFunc(metadataAndText.Text, func(match string) string {
		submatches := reAdmonition.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}

		noteType := toHexoNoteType(submatches[1])
		body := submatches[2]

		return fmt.Sprintf("{%% note %s %%}\n%s\n{%% endnote %%}", noteType, body)
	})

	// I hate regex. Why ReplaceAllString(Text, "") would always leaves an empty line?
	// Or maybe I just suck at regex.

	return metadataAndText
}

func toHexoNoteType(rawType string) string {
	switch strings.ToLower(strings.TrimSpace(rawType)) {
	case "success":
		return "success"
	case "warning":
		return "warning"
	case "tip":
		return "primary"
	case "info":
		return "info"
	default:
		return "primary"
	}
}
