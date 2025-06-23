package markdown

import (
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Valid markdown",
			content: "# Title\n\nThis is a paragraph with **bold** and *italic* text.",
			wantErr: false,
		},
		{
			name:    "Empty content",
			content: "",
			wantErr: false,
		},
		{
			name:    "Unbalanced code blocks",
			content: "```\nCode block without closing",
			wantErr: true,
		},
		{
			name:    "Unbalanced inline code",
			content: "This is `inline code without closing",
			wantErr: true,
		},
		{
			name:    "Content exceeding size limit",
			content: strings.Repeat("a", MaxNotesLength+1),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Clean content",
			content:  "# Title\n\nThis is a paragraph.",
			expected: "# Title\n\nThis is a paragraph.",
		},
		{
			name:     "Content with script tags",
			content:  "# Title\n\n<script>alert('XSS')</script>This is a paragraph.",
			expected: "# Title\n\nThis is a paragraph.",
		},
		{
			name:     "Content with iframe",
			content:  "# Title\n\n<iframe src=\"evil.com\"></iframe>This is a paragraph.",
			expected: "# Title\n\nThis is a paragraph.",
		},
		{
			name:     "Mixed line endings",
			content:  "Line 1\r\nLine 2\rLine 3\nLine 4",
			expected: "Line 1\nLine 2\nLine 3\nLine 4",
		},
		{
			name:     "Multiple consecutive line breaks",
			content:  "Line 1\n\n\n\nLine 2",
			expected: "Line 1\n\nLine 2",
		},
		{
			name:     "Trim whitespace",
			content:  "  \t  # Title  \n\n  ",
			expected: "# Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sanitize(tt.content)
			if result != tt.expected {
				t.Errorf("Sanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Headings without space",
			content:  "#Title\n##Subtitle",
			expected: "# Title\n## Subtitle\n",
		},
		{
			name:     "Unordered list without space",
			content:  "*Item 1\n-Item 2\n+Item 3",
			expected: "* Item 1\n- Item 2\n+ Item 3\n",
		},
		{
			name:     "Ordered list without space",
			content:  "1.First item\n2.Second item",
			expected: "1. First item\n2. Second item\n",
		},
		{
			name:     "Content without final newline",
			content:  "# Title",
			expected: "# Title\n",
		},
		{
			name:     "Empty content",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(tt.content)
			if result != tt.expected {
				t.Errorf("Format() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateBalancedElements(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Balanced code blocks",
			content:  "```\nCode block\n```",
			expected: true,
		},
		{
			name:     "Unbalanced code blocks",
			content:  "```\nCode block",
			expected: false,
		},
		{
			name:     "Balanced inline code",
			content:  "This is `inline code`",
			expected: true,
		},
		{
			name:     "Unbalanced inline code",
			content:  "This is `inline code",
			expected: false,
		},
		{
			name:     "Multiple code blocks",
			content:  "```\nCode block 1\n```\n```\nCode block 2\n```",
			expected: true,
		},
		{
			name:     "Mixed inline and block code",
			content:  "This is `inline code` and ```\nCode block\n```",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateBalancedElements(tt.content)
			if result != tt.expected {
				t.Errorf("validateBalancedElements() = %v, want %v", result, tt.expected)
			}
		})
	}
}
