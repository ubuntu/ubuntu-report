package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

func TestGenerateREADME(t *testing.T) {
	if !*generate {
		t.Skip("skipping README generation, --generate isn't set")
	}
	t.Parallel()

	sp := filepath.Join("..", "..", "README.md")
	dp := sp + ".new"
	src, err := os.Open(sp)
	if err != nil {
		t.Fatalf("couldn't open %s: %v", sp, err)
	}
	defer src.Close()
	dst, err := os.Create(dp)
	if err != nil {
		t.Fatalf("couldn't create %s: %v", dp, err)
	}
	defer dst.Close()

	// write start of README
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		txt := scanner.Text()
		mustWrite(t, dst, txt)
		if txt == "## Command line usage" {
			mustWrite(t, dst, "")
			break
		}
	}

	// write generated command line
	cmds := []*cobra.Command{generateRootCmd()}
	cmds = append(cmds, cmds[0].Commands()...)
	for _, cmd := range cmds {
		pr, pw := io.Pipe()

		go func() {
			if err := doc.GenMarkdown(cmd, pw); err != nil {
				pw.CloseWithError(errors.Wrapf(err, "generate markdown for %s failed", cmd.Name()))
				return
			}
			pw.Close()
		}()

		mScanner := bufio.NewScanner(pr)
		for mScanner.Scan() {
			txt := mScanner.Text()
			if strings.HasPrefix(txt, "### SEE ALSO") {
				break
			}
			// interactive is an alias to ubuntu-report only, don't file up more info
			if cmd.Name() == "interactive" && strings.HasPrefix(txt, "### Synopsis") {
				break
			}
			// add a subindentation
			if strings.HasPrefix(txt, "##") {
				txt = "#" + txt
			}
			mustWrite(t, dst, txt)
		}

		if err = mScanner.Err(); err != nil {
			t.Fatalf("error while reading generated markdown for %s: %v", cmd.Name(), err)
		}
	}

	// skip to next paragraph (ignore previous generation) and write to the end of file
	skip := true
	for scanner.Scan() {
		txt := scanner.Text()
		if skip && strings.HasPrefix(txt, "## ") {
			skip = false
		}
		if !skip {
			mustWrite(t, dst, txt)
		}
	}
	if err = scanner.Err(); err != nil {
		t.Fatalf("error while reading %s: %v", sp, err)
	}

	dst.Close()
	if err = os.Rename(dp, sp); err != nil {
		t.Fatalf("couldn't rename %s to %s: %v", dp, sp, err)
	}
}

func mustWrite(t *testing.T, f *os.File, s string) {
	if _, err := f.WriteString(s + "\n"); err != nil {
		t.Fatalf("couldn't write '%s' to %s: %v", s, f.Name(), err)
	}
}
