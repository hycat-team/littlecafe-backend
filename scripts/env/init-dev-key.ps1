# Create or inspect the age identity used by SOPS on this machine.
#
# Usage:
#   pwsh scripts/env/init-dev-key.ps1        # create if missing, then print the public key
#   pwsh scripts/env/init-dev-key.ps1 -New   # add another identity to the same file
#
# Only the public key (age1...) is safe to share. Never send the private key.
[CmdletBinding()]
param(
    [switch]$New
)

$ErrorActionPreference = 'Stop'

$keyDir = Join-Path $HOME '.config\sops\age'
$keyFile = Join-Path $keyDir 'keys.txt'

$userKeyFile = [Environment]::GetEnvironmentVariable('SOPS_AGE_KEY_FILE', 'User')
if (-not $userKeyFile) {
    [Environment]::SetEnvironmentVariable('SOPS_AGE_KEY_FILE', $keyFile, 'User')
    $env:SOPS_AGE_KEY_FILE = $keyFile
    Write-Host "Set user environment variable SOPS_AGE_KEY_FILE to $keyFile"
    Write-Host "Restart the terminal so new sessions pick it up."
} elseif ($userKeyFile -ne $keyFile) {
    Write-Warning "SOPS_AGE_KEY_FILE already points to $userKeyFile"
}

if (-not (Get-Command age-keygen -ErrorAction SilentlyContinue)) {
    throw 'age-keygen not found. Install age first: winget install --id FiloSottile.age -e'
}

New-Item -ItemType Directory -Force $keyDir | Out-Null

$existing = $null
if (Test-Path $keyFile) {
    $match = Select-String -Path $keyFile -Pattern '^# public key: (age1\S+)$' -AllMatches
    if ($match) {
        $existing = $match.Matches[-1].Groups[1].Value
    }
}

if ($existing -and -not $New) {
    Write-Host "Existing age identity found."
    Write-Host "Public key (safe to share): $existing"
    Write-Host "Private key file (never share or commit): $keyFile"
    Write-Host "Run with -New to add another identity to this file."
    exit 0
}

$output = & age-keygen 2>$null
if ($LASTEXITCODE -ne 0 -or -not $output) {
    throw 'age-keygen failed'
}

$output | Out-File -Append -Encoding ascii $keyFile
$public = ($output | Select-String '^# public key: (age1\S+)').Matches[0].Groups[1].Value

Write-Host "Created a new age identity."
Write-Host "Public key (safe to share): $public"
Write-Host "Private key file (never share or commit): $keyFile"
Write-Host "Back up this file in your password manager; losing it only requires re-enrolment."
