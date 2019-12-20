package messaging

import "bytes"

// GenerateMessageContent helps generating a string of a
// given size based on the provided pattern. It can be used
// to generate large message content with predictable result.
func GenerateMessageContent(pattern string, size int) string {
	var buf bytes.Buffer
	patLen := len(pattern)
	times := size / patLen
	rem := size % patLen
	for i := 0; i < times; i++ {
		buf.WriteString(pattern)
	}
	if rem > 0 {
		buf.WriteString(pattern[:rem])
	}
	return buf.String()
}
