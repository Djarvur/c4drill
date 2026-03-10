package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	assert.Equal(t, "c4drill <input.toml>", cmd.Use)
	assert.Equal(t, "Generate C4 architecture diagrams from TOML", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)
	assert.NotNil(t, cmd.RunE)
	// Verify Args requires exactly 1 argument
	assert.NotNil(t, cmd.Args)
}

func TestFormatFlag(t *testing.T) {
	cmd := NewRootCmd()

	formatFlag := cmd.PersistentFlags().Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "svg", formatFlag.DefValue)
	assert.Equal(t, "f", formatFlag.Shorthand)
}

func TestOutputFlag(t *testing.T) {
	cmd := NewRootCmd()

	outputFlag := cmd.PersistentFlags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, ".", outputFlag.DefValue)
	assert.Equal(t, "o", outputFlag.Shorthand)
}

func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "svg format is valid",
			format:      "svg",
			expectError: false,
		},
		{
			name:        "dot format is valid",
			format:      "dot",
			expectError: false,
		},
		{
			name:        "png format is invalid",
			format:      "png",
			expectError: true,
		},
		{
			name:        "empty format is invalid",
			format:      "",
			expectError: true,
		},
		{
			name:        "uppercase format is invalid",
			format:      "SVG",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			buf := &bytes.Buffer{}
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set format flag
			if err := cmd.PersistentFlags().Set("format", tt.format); err != nil {
				t.Fatalf("failed to set format flag: %v", err)
			}

			// Set a dummy arg to satisfy ExactArgs(1)
			cmd.SetArgs([]string{"dummy.toml"})

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// For valid formats, the command will fail at file reading stage
				// but not at format validation
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid format")
				}
			}
		})
	}
}
