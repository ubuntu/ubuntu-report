package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra/doc"
)

/*
 * using test file for manpage and bash completion generate so that
 * we don't embed the code and dependencies in final binary
 */

var out = "build"

var generate = flag.Bool("generate", false, "generate manpages and completion files")

func TestGenerateManpage(t *testing.T) {
	if !*generate {
		t.Skip("skipping man page generation, --generate isn't set")
	}
	t.Parallel()

	if err := os.Mkdir(out, 0755); err != nil && os.IsNotExist(err) {
		t.Fatalf("couldn't create %s directory: %v", out, err)
	}
	header := &doc.GenManHeader{
		Title:   "Ubuntu Report",
		Section: "3",
	}
	if err := doc.GenManTree(generateRootCmd(), header, out); err != nil {
		t.Fatalf("couldn't generate manpage: %v", err)
	}
}

func TestGenerateCompletion(t *testing.T) {
	if !*generate {
		t.Skip("skipping bash and zsh completion generation, --generate isn't set")
	}
	t.Parallel()

	rootCmd := generateRootCmd()
	if err := os.Mkdir(out, 0755); err != nil && os.IsNotExist(err) {
		t.Fatalf("couldn't create %s directory: %v", out, err)
	}
	if err := rootCmd.GenBashCompletionFile(filepath.Join(out, "bash-completion")); err != nil {
		t.Fatalf("couldn't generate bash completion: %v", err)
	}
	if err := rootCmd.GenZshCompletionFile(filepath.Join(out, "zsh-completion")); err != nil {
		t.Fatalf("couldn't generate bazshsh completion: %v", err)
	}
}
