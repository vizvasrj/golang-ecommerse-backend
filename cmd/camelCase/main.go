package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"unicode"
)

// Convert snake_case to camelCase
func toCamelCase(input string) string {
	isToUpper := false
	output := ""
	for i, v := range input {
		if v == '_' {
			isToUpper = true
		} else {
			if isToUpper {
				output += string(unicode.ToUpper(v))
				isToUpper = false
			} else {
				if i == 0 {
					output += string(unicode.ToLower(v))
				} else {
					output += string(v)
				}
			}
		}
	}
	return output
}

func main() {
	// Read the file content
	filePath := "pkg/module/review/model.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	// Regular expression to find JSON tags
	re := regexp.MustCompile(`json:"([^"]+)"`)

	// Replace JSON tags with camel case versions
	updatedContent := re.ReplaceAllFunc(content, func(match []byte) []byte {
		// Extract the JSON tag value
		tag := string(match)
		tagValue := tag[6 : len(tag)-1]

		// Convert to camel case
		camelCaseValue := toCamelCase(tagValue)

		// Replace the original tag with the camel case version
		return []byte(fmt.Sprintf(`json:"%s"`, camelCaseValue))
	})

	// Write the updated content back to the file
	err = os.WriteFile(filePath, updatedContent, 0644)
	if err != nil {
		log.Fatalf("failed to write file: %v", err)
	}

	fmt.Println("JSON tags have been converted to camel case.")
}
