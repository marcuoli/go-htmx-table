# Explicit project artifact build. Tests/imports/planning never call this script.
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$StorageRoot,
    [Parameter(Mandatory)][string]$Store
)

$ErrorActionPreference = 'Stop'
$sourceRoot = Split-Path -Parent $PSScriptRoot
if (![IO.Path]::IsPathFullyQualified($StorageRoot) -or ![IO.Path]::IsPathFullyQualified($Store)) {
    throw 'StorageRoot and Store must be absolute external paths.'
}
$sourcePrefix = [IO.Path]::GetFullPath($sourceRoot).TrimEnd('\', '/') + [IO.Path]::DirectorySeparatorChar
foreach ($externalPath in @($StorageRoot, $Store)) {
    $fullPath = [IO.Path]::GetFullPath($externalPath)
    if ($fullPath.TrimEnd('\', '/') -eq $sourcePrefix.TrimEnd('\', '/') -or
        $fullPath.StartsWith($sourcePrefix, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'Build storage and authority must remain outside source.'
    }
}

$requestID = [Guid]::NewGuid().ToString('N')
$attemptPath = Join-Path $StorageRoot "artifacts/build/module-version/$requestID"
$scratch = Join-Path $StorageRoot "tmp/module-version/$requestID"
$envNames = @('TEMP', 'TMP', 'TMPDIR', 'GOTMPDIR', 'GOCACHE', 'GOMODCACHE', 'GOWORK', 'CGO_ENABLED')
$savedEnvironment = @{}
foreach ($name in $envNames) { $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }

Push-Location -LiteralPath $sourceRoot
try {
    # One persisted ID per invocation; a failed compilation retains its reservation.
    # A new invocation always uses a new ID, even for identical source.
    $env:TEMP = $scratch
    $env:TMP = $scratch
    $env:TMPDIR = $scratch
    $env:GOTMPDIR = Join-Path $scratch 'go-build'
    $env:GOCACHE = Join-Path $StorageRoot 'caches/go-build'
    $env:GOMODCACHE = Join-Path $StorageRoot 'caches/go-modules'
    $env:GOWORK = 'off'
    $env:CGO_ENABLED = '0'
    foreach ($directory in @($attemptPath, $scratch, $env:GOTMPDIR, $env:GOCACHE, $env:GOMODCACHE)) {
        New-Item -ItemType Directory -Path $directory -Force | Out-Null
    }
    [ordered]@{ request_id = $requestID; app = 'go-htmx-table'; repo = $sourceRoot; store = $Store } |
        ConvertTo-Json | Set-Content -LiteralPath (Join-Path $attemptPath 'request.json')

    $sharedArgs = @('--repo', $sourceRoot, '--store', $Store, '--app', 'go-htmx-table', '--request-id', $requestID)
    $allocation = & go tool go-versioning build @sharedArgs --json 2> (Join-Path $attemptPath 'allocation-errors.log')
    if ($LASTEXITCODE -ne 0) { throw "Allocation failed; inspect $attemptPath before retrying the recorded request." }
    $allocation | Set-Content -LiteralPath (Join-Path $attemptPath 'allocation.json')

    # Reuse the request only to recover the same allocation as shared linker flags.
    # The compiler is invoked once, after both responses have succeeded.
    $linkerFlags = & go tool go-versioning build @sharedArgs --format ldflags 2> (Join-Path $attemptPath 'flags-errors.log')
    if ($LASTEXITCODE -ne 0) { throw "Allocation response retry failed; inspect $attemptPath." }
    $linkerFlags | Set-Content -LiteralPath (Join-Path $attemptPath 'ldflags.txt')
    $binaryName = if ((& go env GOOS) -eq 'windows') { 'module-version.exe' } else { 'module-version' }
    if ($LASTEXITCODE -ne 0) { throw 'Cannot determine target OS.' }
    $binary = Join-Path $attemptPath $binaryName
    & go build -trimpath -buildvcs=false -ldflags ($linkerFlags -join "`n") -o $binary ./scripts/module-version *> (Join-Path $attemptPath 'compile.log')
    if ($LASTEXITCODE -ne 0) { throw "Compilation failed; reservation remains consumed. Evidence: $attemptPath" }
    & go version -m $binary | Set-Content -LiteralPath (Join-Path $attemptPath 'compiler-metadata.txt')
    if ($LASTEXITCODE -ne 0) { throw 'Compiled artifact metadata inspection failed.' }
    Write-Output $binary
} finally {
    Pop-Location
    foreach ($name in $envNames) { [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process') }
}
