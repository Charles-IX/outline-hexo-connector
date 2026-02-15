package processor

import (
	"regexp"
	"strings"
)

type MetadataAndText struct {
	BannerImg string
	IndexImg  string
	Tags      []string
	Text      string
}

func ExtractMetadataAndText(text string) *MetadataAndText {
	metadataAndText := &MetadataAndText{
		Text: text,
	}

	// We want picture which alt text is banner_img or index_img
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

	// I hate regex. Why ReplaceAllString(Text, "") would always leaves an empty line?
	// Or maybe I just suck at regex.

	return metadataAndText
}
