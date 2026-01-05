package onebot

import (
	"testing"
)

func TestConvertLegacyPlaceholders(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Hello [Face6.gif] world",
			expected: "Hello [CQ:face,id=6] world",
		},
		{
			input:    "[@:123456] hello",
			expected: "[CQ:at,qq=123456] hello",
		},
		{
			input:    "Multiple [@:123] [@:456]",
			expected: "Multiple [CQ:at,qq=123] [CQ:at,qq=456]",
		},
		{
			input:    "Check this [Image:abc.jpg] out",
			expected: "Check this [CQ:image,file=abc.jpg] out",
		},
		{
			input:    "Mixed [Face1.gif] and [@:789] and [Image:test.png]",
			expected: "Mixed [CQ:face,id=1] and [CQ:at,qq=789] and [CQ:image,file=test.png]",
		},
		{
			input:    "No change here",
			expected: "No change here",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertLegacyPlaceholders(tt.input)
			if got != tt.expected {
				t.Errorf("ConvertLegacyPlaceholders() = %v, want %v", got, tt.expected)
			}
		})
	}
}
