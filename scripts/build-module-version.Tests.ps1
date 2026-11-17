# Exercise the actual wrapper up to its shared gate without a real allocator.
[CmdletBinding()]
param([Parameter(Mandatory)][string]$TestRoot)

$ErrorActionPreference = 'Stop'
$wrapper = Join-Path $PSScriptRoot 'build-module-version.ps1'
$sourceRoot = Split-Path -Parent $PSScriptRoot
$global:gateCalls = 0
function global:go {
    if (($args -join ' ') -notlike 'tool go-versioning counter-check *') {
        throw "Unexpected Go invocation: $args"
    }
    $global:gateCalls++
    $global:LASTEXITCODE = 42
}

$envNames = @('TEMP', 'TMP', 'TMPDIR', 'GOTMPDIR', 'GOCACHE', 'GOMODCACHE', 'GOWORK', 'CGO_ENABLED')
$saved = @{}
foreach ($name in $envNames) { $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
$savedLocation = (Get-Location).Path

foreach ($case in @('external', 'trailing-separators', 'source-storage', 'source-store', 'relative')) {
    $storage = Join-Path $TestRoot $case
    $store = Join-Path $TestRoot 'unused-authority'
    $expected = 'Counter location validation failed; no build allocation attempted.*'
    $before = $global:gateCalls
    switch ($case) {
        'trailing-separators' { $storage += [IO.Path]::DirectorySeparatorChar; $store += [IO.Path]::DirectorySeparatorChar }
        'source-storage' { $storage = Join-Path $sourceRoot 'must-not-create'; $expected = 'Build storage and authority must remain outside source.' }
        'source-store' { $store = $sourceRoot + [IO.Path]::DirectorySeparatorChar; $expected = 'Build storage and authority must remain outside source.' }
        'relative' { $storage = 'relative-storage'; $expected = 'StorageRoot, Store and MainRepository must be absolute paths.' }
    }
    $caught = $false
    try { & $wrapper -StorageRoot $storage -Store $store -MainRepository $sourceRoot } catch {
        $caught = $true
        if ($_.Exception.Message -notlike $expected) { throw }
    }
    if (!$caught) { throw "Expected preflight failure for $case" }
    $expectedCalls = if ($case -in @('external', 'trailing-separators')) { 1 } else { 0 }
    if ($global:gateCalls - $before -ne $expectedCalls) { throw "Wrong gate count for $case" }
    if ((Get-Location).Path -ne $savedLocation) { throw "Location changed for $case" }
    foreach ($name in $envNames) {
        if ([Environment]::GetEnvironmentVariable($name, 'Process') -cne $saved[$name]) { throw "Environment $name changed for $case" }
    }
}
if (Test-Path -LiteralPath (Join-Path $sourceRoot 'must-not-create')) { throw 'Source-local storage was created' }
