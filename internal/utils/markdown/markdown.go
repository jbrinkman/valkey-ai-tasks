package markdown

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// MaxNotesLength is the maximum allowed length for notes content
const MaxNotesLength = 100000 // 100KB limit for notes

// Common errors
var (
	ErrNotesSizeExceeded = errors.New("notes size exceeds maximum allowed length")
	ErrInvalidMarkdown   = errors.New("invalid markdown content")
)

// Validate checks if the provided markdown content is valid and within size limits
func Validate(content string) error {
	// Check size limit
	if len(content) > MaxNotesLength {
		return ErrNotesSizeExceeded
	}

	// Basic validation for unbalanced markdown elements
	if !validateBalancedElements(content) {
		return ErrInvalidMarkdown
	}

	return nil
}

// Sanitize cleans the markdown content to prevent potential security issues
// and ensures it follows proper markdown formatting
func Sanitize(content string) string {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Remove potentially harmful HTML tags
	content = sanitizeHTML(content)

	// Ensure proper line endings
	content = normalizeLineEndings(content)

	return content
}

// Format applies consistent formatting to markdown content
func Format(content string) string {
	// Ensure content ends with a newline
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// Ensure consistent heading format (space after #)
	content = formatHeadings(content)

	// Ensure consistent list formatting
	content = formatLists(content)

	return content
}

// validateBalancedElements checks for balanced markdown elements like code blocks
func validateBalancedElements(content string) bool {
	// Check for balanced code blocks
	codeBlockCount := strings.Count(content, "```")
	if codeBlockCount%2 != 0 {
		return false
	}

	// Check for balanced inline code
	inlineCodeCount := strings.Count(content, "`")
	// Subtract the code block backticks (each block has 3)
	inlineCodeCount -= codeBlockCount * 3
	return inlineCodeCount%2 == 0
}

// sanitizeHTML removes potentially harmful HTML tags
func sanitizeHTML(content string) string {
	// Simple regex to remove script tags and other potentially harmful elements
	scriptTagRegex := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	content = scriptTagRegex.ReplaceAllString(content, "")

	// Remove iframe tags
	iframeTagRegex := regexp.MustCompile(`(?i)<iframe[\s\S]*?</iframe>`)
	content = iframeTagRegex.ReplaceAllString(content, "")

	// Remove other potentially harmful tags - handle each tag separately since Go regex doesn't support backreferences
	for _, tag := range []string{"object", "embed", "form", "input", "button", "style"} {
		tagRegex := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[\s\S]*?</%s>`, tag, tag))
		content = tagRegex.ReplaceAllString(content, "")
	}

	return content
}

// normalizeLineEndings ensures consistent line endings (LF)
func normalizeLineEndings(content string) string {
	// Replace CRLF with LF
	content = strings.ReplaceAll(content, "\r\n", "\n")
	// Replace CR with LF
	content = strings.ReplaceAll(content, "\r", "\n")
	// Replace multiple consecutive line breaks with two line breaks
	multipleLineBreaksRegex := regexp.MustCompile(`\n{3,}`)
	content = multipleLineBreaksRegex.ReplaceAllString(content, "\n\n")

	return content
}

// formatHeadings ensures headings have a space after the # characters
func formatHeadings(content string) string {
	headingRegex := regexp.MustCompile(`(?m)^(#{1,6})([^#\s].*)$`)
	return headingRegex.ReplaceAllString(content, "$1 $2")
}

// formatLists ensures list items have proper spacing
func formatLists(content string) string {
	// Format unordered lists (ensure space after bullet)
	ulRegex := regexp.MustCompile(`(?m)^(\s*)([*+-])([^\s].*)$`)
	content = ulRegex.ReplaceAllString(content, "$1$2 $3")

	// Format ordered lists (ensure space after number)
	olRegex := regexp.MustCompile(`(?m)^(\s*)(\d+\.)([^\s].*)$`)
	content = olRegex.ReplaceAllString(content, "$1$2 $3")

	return content
}
