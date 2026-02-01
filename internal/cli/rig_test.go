package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rikurb8/bordertown/internal/rig"
)

func setupFakeGit(t *testing.T) string {
	t.Helper()

	binDir := t.TempDir()
	gitPath := filepath.Join(binDir, "git")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = \"clone\" ]; then\n" +
		"  mkdir -p \"$3\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"echo \"unexpected git args: $*\" >&2\n" +
		"exit 1\n"

	if err := os.WriteFile(gitPath, []byte(script), 0755); err != nil {
		t.Fatalf("write fake git: %v", err)
	}

	return binDir
}

func TestRigAddCommandCreatesRegistry(t *testing.T) {
	homeDir := t.TempDir()
	fakeGitDir := setupFakeGit(t)

	t.Setenv("HOME", homeDir)
	t.Setenv("PATH", fakeGitDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"rig", "add", "alpha", "git@github.com:example/alpha.git"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Added rig") {
		t.Errorf("expected output to mention added rig, got %q", result)
	}

	loaded, err := rig.LoadRig("alpha")
	if err != nil {
		t.Fatalf("load rig: %v", err)
	}
	if loaded.Remote != "git@github.com:example/alpha.git" {
		t.Errorf("expected remote to be saved, got %q", loaded.Remote)
	}
	if _, err := os.Stat(loaded.LocalPath); err != nil {
		t.Fatalf("expected rig path to exist: %v", err)
	}
}

func TestRigAddCommandDuplicateName(t *testing.T) {
	homeDir := t.TempDir()
	fakeGitDir := setupFakeGit(t)

	t.Setenv("HOME", homeDir)
	t.Setenv("PATH", fakeGitDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	root := NewRootCommand()
	output := &bytes.Buffer{}
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"rig", "add", "alpha", "git@github.com:example/alpha.git"})

	if err := root.Execute(); err != nil {
		t.Fatalf("expected first add to succeed, got %v", err)
	}

	root = NewRootCommand()
	root.SetOut(output)
	root.SetErr(output)
	root.SetArgs([]string{"rig", "add", "alpha", "git@github.com:example/alpha.git"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error on duplicate rig add")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestResolveRigDirFromRegistry(t *testing.T) {
	homeDir := t.TempDir()
	rigDir := t.TempDir()

	t.Setenv("HOME", homeDir)
	if err := os.MkdirAll(rigDir, 0755); err != nil {
		t.Fatalf("create rig dir: %v", err)
	}

	if err := rig.SaveRegistry(&rig.Registry{
		Rigs: []rig.Rig{
			{
				Name:      "alpha",
				Remote:    "git@github.com:example/alpha.git",
				LocalPath: rigDir,
			},
		},
	}); err != nil {
		t.Fatalf("save registry: %v", err)
	}

	resolved, err := resolveRigDir("alpha", "/fallback")
	if err != nil {
		t.Fatalf("resolve rig dir: %v", err)
	}
	if resolved != rigDir {
		t.Fatalf("expected rig dir %q, got %q", rigDir, resolved)
	}
}
