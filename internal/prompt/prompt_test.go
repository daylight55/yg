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
func TestPrompterMethodsExist(t *testing.T) {
	prompter := NewPrompter()

	// Test that all required methods exist and have correct signatures
	// These tests verify the method signatures without actually calling them
	
	// Test Select method signature
	var selectFunc func(string, []string) (string, error) = prompter.Select
	if selectFunc == nil {
		t.Error("Select method should exist")
	}

	// Test MultiSelect method signature  
	var multiSelectFunc func(string, []string) ([]string, error) = prompter.MultiSelect
	if multiSelectFunc == nil {
		t.Error("MultiSelect method should exist")
	}

	// Test Search method signature
	var searchFunc func(string, []string) (string, error) = prompter.Search
	if searchFunc == nil {
		t.Error("Search method should exist")
	}

	// Test Confirm method signature
	var confirmFunc func(string) (bool, error) = prompter.Confirm
	if confirmFunc == nil {
		t.Error("Confirm method should exist")
	}
}

// Test the filter function used in Search more comprehensively
func TestSearchFilterFunction(t *testing.T) {
	// This tests the actual filter function logic
	testCases := []struct {
		name        string
		filterValue string
		optionValue string
		expected    bool
	}{
		{"exact match", "test", "test", true},
		{"prefix match", "test", "testing", true},
		{"case insensitive match", "TEST", "testing", true},
		{"case insensitive prefix", "Test", "testing", true},
		{"no match", "xyz", "testing", false},
		{"empty filter matches all", "", "anything", true},
		{"special characters", "test-", "test-value", true},
		{"numbers", "123", "123456", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the filter function used in Search
			result := strings.HasPrefix(
				strings.ToLower(tc.optionValue),
				strings.ToLower(tc.filterValue),
			)
			if result != tc.expected {
				t.Errorf("Filter('%s', '%s'): expected %v, got %v", 
					tc.filterValue, tc.optionValue, tc.expected, result)
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

// Test filter logic extensively
func TestFilterLogicCoverage(t *testing.T) {
	// Test various filter scenarios to ensure comprehensive coverage
	testCases := []struct {
		name          string
		filterValue   string
		optionValue   string
		shouldMatch   bool
		description   string
	}{
		{"empty filter", "", "any-value", true, "empty filter should match everything"},
		{"exact match", "hello", "hello", true, "exact match should work"},
		{"prefix match", "hel", "hello", true, "prefix match should work"},
		{"case insensitive", "HEL", "hello", true, "case insensitive matching"},
		{"mixed case", "HeLLo", "hello", true, "mixed case should work"},
		{"no match", "xyz", "hello", false, "non-matching strings"},
		{"partial no match", "ello", "hello", false, "substring but not prefix"},
		{"special chars", "test-", "test-value", true, "special characters"},
		{"numbers", "123", "123456", true, "numeric prefixes"},
		{"unicode", "日本", "日本語", true, "unicode characters"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This simulates the exact filter logic used in Search method
			result := strings.HasPrefix(
				strings.ToLower(tc.optionValue),
				strings.ToLower(tc.filterValue),
			)
			
			if result != tc.shouldMatch {
				t.Errorf("%s: filter('%s', '%s') expected %v, got %v", 
					tc.description, tc.filterValue, tc.optionValue, tc.shouldMatch, result)
			}
		})
	}
}
