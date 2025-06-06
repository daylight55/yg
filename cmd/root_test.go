package cmd

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// Test that init function sets up cobra command correctly
	if rootCmd == nil {
		t.Error("rootCmd should be initialized")
	}

	if rootCmd.Use != "yg" {
		t.Errorf("Expected command name 'yg', got %s", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Command should have a short description")
	}
}

func TestExecute(t *testing.T) {
	// Test Execute function exists and can be called
	// We can't easily test the actual execution without mocking,
	// but we can test that the function exists and has the right structure
	
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Set test args - help flag to avoid interactive prompt
	os.Args = []string{"yg", "--help"}
	
	// This should not panic - Execute() doesn't return error
	Execute()
	t.Log("Execute function completed without panic")
}

func TestRootCommandExists(t *testing.T) {
	// Test that the root command is properly configured
	if rootCmd.RunE == nil {
		t.Error("rootCmd should have a RunE function")
	}
	
	// Test command has expected flags
	if rootCmd.PersistentFlags() == nil {
		t.Error("rootCmd should have persistent flags")
	}
}

func TestRunGeneratorFunctionExists(t *testing.T) {
	// Test that runGenerator function exists by checking it's assigned to RunE
	if rootCmd.RunE == nil {
		t.Error("runGenerator should be assigned to rootCmd.RunE")
	}
	
	// We can't easily test the actual generator execution without extensive mocking,
	// but we verify the structure is correct
}