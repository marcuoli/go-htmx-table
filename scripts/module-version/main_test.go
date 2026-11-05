package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestModuleTagContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		tag  string
	}{
		{"patch", []string{"-previous", "v0.2.0", "-bump", "patch", "-check", "v0.2.1+20261105.14"}, "v0.2.1+20261105.14"},
		{"minor", []string{"-previous", "v0.2.9", "-bump", "minor"}, "v0.3.0+20261105.14"},
		{"major", []string{"-previous", "v0.2.9", "-bump", "major"}, "v1.0.0+20261105.14"},
		{"full previous", []string{"-previous", "v0.2.0+20261105.13", "-bump", "patch"}, "v0.2.1+20261105.14"},
		{"inherit", []string{"-previous", "v0.2.0+20261105.13", "-bump", "inherit"}, "v0.2.0+20261105.14"},
		{"snapshot", []string{"-previous", "v0.2.0-snapshot+20261105.13", "-bump", "patch"}, ""},
		{"missing prefix", []string{"-previous", "0.2.0", "-bump", "patch"}, ""},
		{"missing bump", []string{"-previous", "v0.2.0"}, ""},
		{"mismatch", []string{"-previous", "v0.2.0", "-bump", "patch", "-check", "v0.3.0"}, ""},
		{"extra argument", []string{"-previous", "v0.2.0", "-bump", "patch", "extra"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			args := append(append([]string{}, tc.args...), "-commit-time", "2026-11-05T12:00:00Z", "-pr", "14")
			err := run(args, &output)
			if tc.tag == "" {
				if err == nil || output.Len() != 0 {
					t.Fatalf("invalid plan: err=%v output=%q", err, output.String())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var result struct {
				Tag     string `json:"tag"`
				Version string `json:"version"`
			}
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Tag != tc.tag {
				t.Fatalf("tag = %q, want %q", result.Tag, tc.tag)
			}
			if result.Version != tc.tag[1:] {
				t.Fatalf("version = %q", result.Version)
			}
		})
	}
}

func TestCommitIdentity(t *testing.T) {
	base := []string{"-previous", "v0.2.0", "-bump", "patch"}
	for _, args := range [][]string{
		{},
		{"-commit-time", "2026-11-05T12:00:00Z"},
		{"-commit-time", "invalid", "-pr", "14"},
	} {
		var output bytes.Buffer
		if err := run(append(append([]string{}, base...), args...), &output); err == nil || output.Len() != 0 {
			t.Fatalf("accepted incomplete commit identity: %v, %q", args, output.String())
		}
	}
	var output bytes.Buffer
	if err := run(append(base, "-commit-time", "2026-11-05T01:00:00Z", "-pr", "14"), &output); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Version != "0.2.1+20261104.14" {
		t.Fatalf("commit-local date: %q", result.Version)
	}
}

func TestBuildInfoIsReadOnlyAndSeparateFromPlanning(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"-build-info"}, &output); err != nil {
		t.Fatal(err)
	}
	var info struct {
		Version            string `json:"version"`
		OverallBuildNumber uint64 `json:"overall_build_number"`
	}
	if err := json.Unmarshal(output.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Version == "" || info.OverallBuildNumber != 0 {
		t.Fatalf("unallocated test binary has unexpected metadata: %+v", info)
	}
	output.Reset()
	if err := run([]string{"-build-info", "-pr", "15"}, &output); err == nil || output.Len() != 0 {
		t.Fatal("build-info accepted planning flags")
	}
}
