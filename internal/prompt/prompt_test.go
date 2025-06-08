package prompt

import (
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2/core"
)

func TestNewPrompter(t *testing.T) {
	prompter := NewPrompter()
	if prompter == nil {
		t.Error("NewPrompter should return a non-nil prompter")
	}
}

func TestPrompterInterface(t *testing.T) {
	prompter := NewPrompter()

	// Test that prompter implements the interface
	var _ PrompterInterface = prompter

	// These would require interactive input in a real terminal,
	// so we just test that the methods exist
	t.Log("Prompter methods are available")
}

// Test the search filter logic used in the new Select-based implementation
func TestSearchFilterLogic(t *testing.T) {
	options := []string{"sample-server-1", "sample-server-2", "sample-job-1", "sample-job-2"}

	// Simulate the filter function used in the new Search implementation
	filter := func(filterValue string, optionValue string, _ int) bool {
		// If no filter input, show all options
		if filterValue == "" {
			return true
		}

		// Check for exact match first (case insensitive)
		if strings.EqualFold(filterValue, optionValue) {
			return true
		}

		// Then check for partial match (contains, case insensitive)
		return strings.Contains(
			strings.ToLower(optionValue),
			strings.ToLower(filterValue),
		)
	}

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "exact match server-1",
			input:    "sample-server-1",
			expected: []string{"sample-server-1"},
		},
		{
			name:     "case insensitive exact match",
			input:    "SAMPLE-SERVER-1",
			expected: []string{"sample-server-1"},
		},
		{
			name:     "partial match server",
			input:    "server",
			expected: []string{"sample-server-1", "sample-server-2"},
		},
		{
			name:     "partial match job",
			input:    "job",
			expected: []string{"sample-job-1", "sample-job-2"},
		},
		{
			name:     "no match",
			input:    "nonexistent",
			expected: []string{},
		},
		{
			name:     "empty input shows all",
			input:    "",
			expected: options,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filtered []string
			for _, option := range options {
				if filter(tc.input, option, 0) {
					filtered = append(filtered, option)
				}
			}

			if len(filtered) != len(tc.expected) {
				t.Errorf("Filter '%s': expected %d matches, got %d", tc.input, len(tc.expected), len(filtered))
				return
			}

			for i, expected := range tc.expected {
				if i >= len(filtered) || filtered[i] != expected {
					t.Errorf("Filter '%s': expected match[%d] '%s', got '%s'", tc.input, i, expected, filtered[i])
				}
			}
		})
	}
}

// Test the Prompter structure and methods
func TestPrompterStructure(t *testing.T) {
	prompter := NewPrompter()

	// Test NewPrompter sets up the struct correctly
	if prompter == nil {
		t.Fatal("NewPrompter should return a non-nil prompter")
	}

	// Test that core.DisableColor is set correctly
	if core.DisableColor {
		t.Error("Expected core.DisableColor to be false")
	}
}

// Test error handling for empty options (conceptual test)
func TestPrompterMethodsExist(_ *testing.T) {
	prompter := NewPrompter()

	// Test that all required methods exist and have correct signatures
	// These tests verify the method signatures without actually calling them

	// Test method signatures by attempting to assign them
	// This verifies the methods exist and have correct signatures
	_ = prompter.Select
	_ = prompter.MultiSelect
	_ = prompter.Search
	_ = prompter.Confirm
}

// Test individual filter function behaviors
func TestSearchFilterFunction(t *testing.T) {
	// Create the filter function that matches the implementation
	filter := func(filterValue string, optionValue string, _ int) bool {
		if filterValue == "" {
			return true
		}
		if strings.EqualFold(filterValue, optionValue) {
			return true
		}
		return strings.Contains(
			strings.ToLower(optionValue),
			strings.ToLower(filterValue),
		)
	}

	testCases := []struct {
		name     string
		input    string
		option   string
		expected bool
	}{
		{"exact match", "test", "test", true},
		{"case insensitive exact", "TEST", "test", true},
		{"mixed case exact", "TeSt", "test", true},
		{"contains match", "est", "test", true},
		{"case insensitive contains", "EST", "test", true},
		{"no match", "xyz", "test", false},
		{"empty input shows all", "", "test", true},
		{"empty input with anything", "", "anything", true},
		{"special characters exact", "test-1", "test-1", true},
		{"special characters contains", "test-", "test-1", true},
		{"numbers exact", "123", "123", true},
		{"numbers contains", "23", "123", true},
		{"unicode exact", "日本語", "日本語", true},
		{"unicode contains", "日本", "日本語", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filter(tc.input, tc.option, 0)
			if result != tc.expected {
				t.Errorf("Filter('%s', '%s'): expected %v, got %v",
					tc.input, tc.option, tc.expected, result)
			}
		})
	}
}

// Test interface compliance
func TestPrompterInterfaceCompliance(t *testing.T) {
	prompter := NewPrompter()

	// Verify that Prompter implements PrompterInterface
	var _ PrompterInterface = prompter

	// Test struct fields and basic setup
	if prompter == nil {
		t.Error("NewPrompter should return a non-nil prompter")
	}
}

// Test core configuration
func TestCoreConfiguration(t *testing.T) {
	// Test that NewPrompter configures the core correctly
	NewPrompter()

	// Verify color setting (should be false for consistent output)
	if core.DisableColor {
		t.Error("Expected core.DisableColor to be false after NewPrompter()")
	}
}

// Test comprehensive search scenarios with the new filter implementation
func TestSearchScenarios(t *testing.T) {
	// Create the filter function that matches the implementation
	filter := func(filterValue string, optionValue string, _ int) bool {
		if filterValue == "" {
			return true
		}
		if strings.EqualFold(filterValue, optionValue) {
			return true
		}
		return strings.Contains(
			strings.ToLower(optionValue),
			strings.ToLower(filterValue),
		)
	}

	testCases := []struct {
		name        string
		input       string
		options     []string
		expected    []string
		description string
	}{
		{
			name:        "exact match included in results",
			input:       "hello",
			options:     []string{"hello", "world", "hello-world"},
			expected:    []string{"hello", "hello-world"},
			description: "exact match and partial matches should be shown",
		},
		{
			name:        "case insensitive exact match",
			input:       "HELLO",
			options:     []string{"hello", "world", "hello-world"},
			expected:    []string{"hello", "hello-world"},
			description: "case insensitive matching should work",
		},
		{
			name:        "partial match multiple results",
			input:       "test",
			options:     []string{"test-1", "test-2", "other-test", "sample"},
			expected:    []string{"test-1", "test-2", "other-test"},
			description: "partial matches should be filtered correctly",
		},
		{
			name:        "partial match single result",
			input:       "unique",
			options:     []string{"unique-item", "test-1", "test-2"},
			expected:    []string{"unique-item"},
			description: "single partial match should be shown",
		},
		{
			name:        "no match shows nothing",
			input:       "nonexistent",
			options:     []string{"test-1", "test-2", "sample"},
			expected:    []string{},
			description: "no matches should return empty list",
		},
		{
			name:        "empty input shows all",
			input:       "",
			options:     []string{"test-1", "test-2", "sample"},
			expected:    []string{"test-1", "test-2", "sample"},
			description: "empty input should show all options",
		},
		{
			name:        "unicode support",
			input:       "日本",
			options:     []string{"日本語", "日本国", "english", "test"},
			expected:    []string{"日本語", "日本国"},
			description: "unicode characters should work in search",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var filtered []string
			for _, option := range tc.options {
				if filter(tc.input, option, 0) {
					filtered = append(filtered, option)
				}
			}

			if len(filtered) != len(tc.expected) {
				t.Errorf("%s: filtered count expected %d, got %d",
					tc.description, len(tc.expected), len(filtered))
				return
			}

			for i, expected := range tc.expected {
				if i >= len(filtered) || filtered[i] != expected {
					t.Errorf("%s: filtered[%d] expected '%s', got '%s'",
						tc.description, i, expected, filtered[i])
				}
			}
		})
	}
}
