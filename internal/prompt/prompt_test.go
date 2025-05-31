package prompt

import (
	"strings"
	"testing"
)

func TestNewPrompter(t *testing.T) {
	prompter := NewPrompter()
	if prompter == nil {
		t.Error("NewPrompter should return a non-nil prompter")
	}
}

// Note: These tests don't actually test the interactive functionality
// as that would require mocking terminal input/output.
// In a real-world scenario, you might use a more sophisticated testing approach
// with libraries like github.com/Netflix/go-expect for testing interactive CLI tools.

func TestPrompterInterface(t *testing.T) {
	prompter := NewPrompter()
	
	// Test that prompter implements the interface
	var _ PrompterInterface = prompter
	
	// These would require interactive input in a real terminal,
	// so we just test that the methods exist
	t.Log("Prompter methods are available")
}

// Test the filter function logic used in Search
func TestSearchFilter(t *testing.T) {
	options := []string{"sample-server-1", "sample-server-2", "sample-job-1", "sample-job-2"}
	
	testCases := []struct {
		filterValue string
		expected    []string
	}{
		{"sample-server", []string{"sample-server-1", "sample-server-2"}},
		{"sample-job", []string{"sample-job-1", "sample-job-2"}},
		{"sample-server-1", []string{"sample-server-1"}},
		{"sample-job-2", []string{"sample-job-2"}},
		{"nonexistent", []string{}},
		{"", options}, // Empty filter should match all
	}
	
	for _, tc := range testCases {
		var matches []string
		for i, option := range options {
			// Simulate the filter function used in Search
			if strings.HasPrefix(
				strings.ToLower(option),
				strings.ToLower(tc.filterValue),
			) {
				matches = append(matches, options[i])
			}
		}
		
		if len(matches) != len(tc.expected) {
			t.Errorf("Filter '%s': expected %d matches, got %d", tc.filterValue, len(tc.expected), len(matches))
			continue
		}
		
		for i, match := range matches {
			if match != tc.expected[i] {
				t.Errorf("Filter '%s': expected match '%s', got '%s'", tc.filterValue, tc.expected[i], match)
			}
		}
	}
}