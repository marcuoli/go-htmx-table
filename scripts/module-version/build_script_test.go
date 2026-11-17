package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBuildScriptPreflight(t *testing.T) {
	pwsh, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("PowerShell 7 is required for the build-wrapper regression")
	}
	script, err := filepath.Abs(filepath.Join("..", "build-module-version.Tests.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(pwsh, "-NoProfile", "-NonInteractive", "-File", script, "-TestRoot", t.TempDir())
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build-wrapper preflight: %v\n%s", err, output)
	}
}
