// Command module-version plans full release identities using go-versioning.
// It requires explicit merge metadata and never changes Git refs.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	versioning "github.com/marcuoli/go-versioning"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "module-version:", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("module-version", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	previous := flags.String("previous", "", "previous full release tag or legacy module tag")
	bump := flags.String("bump", "", "major, minor or patch")
	expected := flags.String("check", "", "optional expected full release tag")
	commitTime := flags.String("commit-time", "", "target commit timestamp in RFC3339 format")
	pr := flags.Uint64("pr", 0, "merged pull-request number")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(output, "Usage: module-version -previous TAG -bump inherit|major|minor|patch -commit-time RFC3339 -pr NUMBER [-check FULL_TAG]")
			return err
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("positional arguments are not accepted")
	}
	if !strings.HasPrefix(*previous, "v") {
		return fmt.Errorf("-previous must be a v-prefixed release tag")
	}
	core, err := versioning.ParseCore(strings.TrimPrefix(*previous, "v"))
	if err != nil {
		previousVersion, _, parseErr := versioning.ParseTag(*previous)
		if parseErr != nil {
			return fmt.Errorf("previous release tag: %w", parseErr)
		}
		core = previousVersion.Core
	}
	// Inherit preserves the core while the actual merge date/PR identifies a new build.
	switch versioning.Bump(*bump) {
	case versioning.BumpInherit, versioning.BumpMajor, versioning.BumpMinor, versioning.BumpPatch:
	default:
		return fmt.Errorf("-bump must be inherit, major, minor or patch")
	}
	timestamp, err := time.Parse(time.RFC3339, *commitTime)
	if err != nil {
		return fmt.Errorf("-commit-time must be the target commit's RFC3339 timestamp: %w", err)
	}
	// go-versioning selects the commit's calendar date in America/Sao_Paulo.
	// Never substitute the execution time or invent a PR for an unmerged build.
	full, err := versioning.Next(core, timestamp, *pr, versioning.Bump(*bump))
	if err != nil {
		return err
	}
	tag, err := full.Tag()
	if err != nil {
		return err
	}
	if *expected != "" && tag != *expected {
		return fmt.Errorf("planned tag %s differs from expected %s", tag, *expected)
	}
	return json.NewEncoder(output).Encode(struct {
		Previous string `json:"previous"`
		Bump     string `json:"bump"`
		Tag      string `json:"tag"`
		Version  string `json:"version"`
	}{*previous, *bump, tag, full.String()})
}
