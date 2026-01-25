package cli

import (
	"bytes"
	"testing"
)

func TestRootCommandExecutesDummy(t *testing.T) {
	root := NewRootCommand()
	output := &bytes.Buffer{}

	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"hello"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.Len() == 0 {
		t.Fatalf("expected command output, got none")
	}
}
