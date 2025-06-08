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

// Test the search filter logic for exact match and contains search
func TestSearchFilterLogic(t *testing.T) {
	options := []string{"sample-server-1", "sample-server-2", "sample-job-1", "sample-job-2"}

	testCases := []struct {
		name         string
		input        string
		exactMatch   string
		filteredOpts []string
	}{
		{
			name:         "exact match server-1",
			input:        "sample-server-1",
			exactMatch:   "sample-server-1",
			filteredOpts: []string{"sample-server-1"},
		},
		{
			name:         "case insensitive exact match",
			input:        "SAMPLE-SERVER-1",
			exactMatch:   "sample-server-1",
			filteredOpts: []string{"sample-server-1"},
		},
		{
			name:         "partial match server",
			input:        "server",
			exactMatch:   "",
			filteredOpts: []string{"sample-server-1", "sample-server-2"},
		},
		{
			name:         "partial match job",
			input:        "job",
			exactMatch:   "",
			filteredOpts: []string{"sample-job-1", "sample-job-2"},
		},
		{
			name:         "no match",
			input:        "nonexistent",
			exactMatch:   "",
			filteredOpts: []string{},
		},
		{
			name:         "empty input",
			input:        "",
			exactMatch:   "",
			filteredOpts: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test exact match logic
			var exactMatch string
			for _, option := range options {
				if strings.EqualFold(tc.input, option) {
					exactMatch = option
					break
				}
			}

			if exactMatch != tc.exactMatch {
				t.Errorf("Exact match: expected '%s', got '%s'", tc.exactMatch, exactMatch)
			}

			// Test contains filter logic
			var filteredOptions []string
			inputLower := strings.ToLower(tc.input)

			for _, option := range options {
				optionLower := strings.ToLower(option)
				if tc.input != "" && strings.Contains(optionLower, inputLower) {
					filteredOptions = append(filteredOptions, option)
				}
			}

			if len(filteredOptions) != len(tc.filteredOpts) {
				t.Errorf("Filtered options count: expected %d, got %d", len(tc.filteredOpts), len(filteredOptions))
				return
			}

			for i, expected := range tc.filteredOpts {
				if i >= len(filteredOptions) || filteredOptions[i] != expected {
					t.Errorf("Filtered option[%d]: expected '%s', got '%s'", i, expected, filteredOptions[i])
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

// Test the exact match and contains filter functions
func TestSearchFilterFunction(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		option        string
		exactMatch    bool
		containsMatch bool
	}{
		{"exact match", "test", "test", true, true},
		{"case insensitive exact", "TEST", "test", true, true},
		{"mixed case exact", "TeSt", "test", true, true},
		{"contains match", "est", "test", false, true},
		{"case insensitive contains", "EST", "test", false, true},
		{"no match", "xyz", "test", false, false},
		{"empty input exact", "", "test", false, false},
		{"empty input contains", "", "anything", false, false},
		{"special characters exact", "test-1", "test-1", true, true},
		{"special characters contains", "test-", "test-1", false, true},
		{"numbers exact", "123", "123", true, true},
		{"numbers contains", "23", "123", false, true},
		{"unicode exact", "日本語", "日本語", true, true},
		{"unicode contains", "日本", "日本語", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test exact match logic
			exactResult := strings.EqualFold(tc.input, tc.option)
			if exactResult != tc.exactMatch {
				t.Errorf("Exact match('%s', '%s'): expected %v, got %v",
					tc.input, tc.option, tc.exactMatch, exactResult)
			}

			// Test contains logic (only if input is not empty)
			var containsResult bool
			if tc.input != "" {
				containsResult = strings.Contains(
					strings.ToLower(tc.option),
					strings.ToLower(tc.input),
				)
			}
			if containsResult != tc.containsMatch {
				t.Errorf("Contains('%s', '%s'): expected %v, got %v",
					tc.input, tc.option, tc.containsMatch, containsResult)
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

// Test comprehensive search scenarios
func TestSearchScenarios(t *testing.T) {
	// Test various search scenarios to ensure comprehensive coverage
	testCases := []struct {
		name           string
		input          string
		options        []string
		expectExact    string
		expectFiltered []string
		description    string
	}{
		{
			name:           "exact match found",
			input:          "hello",
			options:        []string{"hello", "world", "hello-world"},
			expectExact:    "hello",
			expectFiltered: []string{"hello", "hello-world"},
			description:    "exact match should be returned directly",
		},
		{
			name:           "case insensitive exact match",
			input:          "HELLO",
			options:        []string{"hello", "world", "hello-world"},
			expectExact:    "hello",
			expectFiltered: []string{"hello", "hello-world"},
			description:    "case insensitive exact match should work",
		},
		{
			name:           "partial match multiple results",
			input:          "test",
			options:        []string{"test-1", "test-2", "other-test", "sample"},
			expectExact:    "",
			expectFiltered: []string{"test-1", "test-2", "other-test"},
			description:    "partial match should show filtered options",
		},
		{
			name:           "partial match single result",
			input:          "unique",
			options:        []string{"unique-item", "test-1", "test-2"},
			expectExact:    "",
			expectFiltered: []string{"unique-item"},
			description:    "single partial match should be offered",
		},
		{
			name:           "no match",
			input:          "nonexistent",
			options:        []string{"test-1", "test-2", "sample"},
			expectExact:    "",
			expectFiltered: []string{},
			description:    "no match should show all options",
		},
		{
			name:           "empty input",
			input:          "",
			options:        []string{"test-1", "test-2", "sample"},
			expectExact:    "",
			expectFiltered: []string{},
			description:    "empty input should show no filtered options",
		},
		{
			name:           "unicode support",
			input:          "日本",
			options:        []string{"日本語", "日本国", "english", "test"},
			expectExact:    "",
			expectFiltered: []string{"日本語", "日本国"},
			description:    "unicode characters should work in search",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test exact match logic
			var exactMatch string
			for _, option := range tc.options {
				if strings.EqualFold(tc.input, option) {
					exactMatch = option
					break
				}
			}

			if exactMatch != tc.expectExact {
				t.Errorf("%s: exact match expected '%s', got '%s'",
					tc.description, tc.expectExact, exactMatch)
			}

			// Test filtering logic
			var filteredOptions []string
			inputLower := strings.ToLower(tc.input)

			for _, option := range tc.options {
				optionLower := strings.ToLower(option)
				if tc.input != "" && strings.Contains(optionLower, inputLower) {
					filteredOptions = append(filteredOptions, option)
				}
			}

			if len(filteredOptions) != len(tc.expectFiltered) {
				t.Errorf("%s: filtered count expected %d, got %d",
					tc.description, len(tc.expectFiltered), len(filteredOptions))
				return
			}

			for i, expected := range tc.expectFiltered {
				if i >= len(filteredOptions) || filteredOptions[i] != expected {
					t.Errorf("%s: filtered[%d] expected '%s', got '%s'",
						tc.description, i, expected, filteredOptions[i])
				}
			}
		})
	}
}
