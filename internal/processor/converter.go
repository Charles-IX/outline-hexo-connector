package processor

import (
	"fmt"
	"log"
	"outline-hexo-connector/internal/outline"
	"regexp"
)

type AttachmentUrlProvider interface {
	GetAttachmentURL(attachmentID string) (string, error)
}

func convertAttachmentUrl(provider AttachmentUrlProvider, blog *outline.Document) (*outline.Document, error) {
	// Some regex magic to find outline attachment urls
	re := regexp.MustCompile(`(?P<prefix>!?)\[(?P<text>.*?)\]\(/api/attachments\.redirect\?id=(?P<id>[a-f0-9-]{36})(?P<extra>.*?)\)`)

	newText := re.ReplaceAllStringFunc(blog.Text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		prefix := submatches[1]
		text := submatches[2]
		id := submatches[3]

		rawUrl, err := provider.GetAttachmentURL(id)
		if err != nil {
			log.Printf("Error getting attachment OSS URL - %v", err)
			return match
		}

		textCleanup := regexp.MustCompile(`\s+(\d+x\d+|\d+)$`)
		cleanText := textCleanup.ReplaceAllString(text, "")

		return fmt.Sprintf("%s[%s](%s)", prefix, cleanText, rawUrl)
	})

	reExtra := regexp.MustCompile(`\s+\\"=?\d*x?\d*\\"`)
	blog.Text = reExtra.ReplaceAllString(newText, "")

	return blog, nil

}
