package processor

import (
	"fmt"
	"log"
	"regexp"
)

type AttachmentUrlProvider interface {
	GetAttachmentUrl(attachmentID string) (string, error)
}

func ConvertAttachmentUrl(provider AttachmentUrlProvider, text string) (string, error) {
	// Some regex magic to find outline attachment urls
	re := regexp.MustCompile(`(?P<prefix>!?)\[(?P<text>.*?)\]\(/api/attachments\.redirect\?id=(?P<id>[a-f0-9-]{36})(?P<extra>.*?)\)`)

	newText := re.ReplaceAllStringFunc(text, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		prefix := submatches[1]
		text := submatches[2]
		id := submatches[3]

		rawUrl, err := provider.GetAttachmentUrl(id)
		if err != nil {
			log.Printf("Error getting attachment OSS URL - %v", err)
			return match
		}

		textCleanup := regexp.MustCompile(`\s+(\d+x\d+|\d+)$`)
		cleanText := textCleanup.ReplaceAllString(text, "")

		return fmt.Sprintf("%s[%s](%s)", prefix, cleanText, rawUrl)
	})

	reExtra := regexp.MustCompile(`\s+\\"=?\d*x?\d*\\"`)
	text = reExtra.ReplaceAllString(newText, "")

	return text, nil

}
