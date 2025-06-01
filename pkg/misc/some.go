package misc

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
)

func GenerateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove non-alphanumeric characters and replace spaces with hyphens
	re := regexp.MustCompile("[^a-z0-9]+")
	s = re.ReplaceAllString(s, "-")

	// Trim hyphens from the start and end
	s = strings.Trim(s, "-")

	// Generate a random string for uniqueness
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	uniqueID := fmt.Sprintf("-%x", b)

	return s + uniqueID
}
