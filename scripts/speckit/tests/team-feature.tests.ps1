#!/usr/bin/env pwsh
# Plain PowerShell smoke tests for scripts/speckit/team-feature.ps1.
#
# The tests build a throwaway bare repository and two clones so the helper is
# exercised the way two developers would use it. No network access is required.

$ErrorActionPreference = 'Stop'

$repoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '../../..')).Path
$tool = Join-Path $repoRoot 'scripts/speckit/team-feature.ps1'
$pwsh = (Get-Process -Id $PID).Path
$tempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ('littlecafe-team-feature-' + [guid]::NewGuid().ToString('N'))

$script:failed = 0
$script:passed = 0

function Assert-True {
    param([bool]$Condition, [string]$Message)

    if ($Condition) {
        $script:passed++
        Write-Host "PASS: $Message" -ForegroundColor Green
    } else {
        $script:failed++
        Write-Host "FAIL: $Message" -ForegroundColor Red
    }
}

function Invoke-Git {
    param([string]$Directory, [string[]]$Arguments)

    $output = & git -C $Directory @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed in $Directory`n$($output -join "`n")"
    }
    return ($output -join "`n")
}

function Invoke-Tool {
    param([string]$Directory, [string[]]$Arguments)

    # Expected failures (wrong branch, duplicate slug) write to stderr. Keep
    # native stderr from becoming a terminating error in Windows PowerShell 5.1.
    $oldPreference = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    Push-Location -LiteralPath $Directory
    try {
        $output = & $pwsh -NoProfile -File $tool @Arguments 2>&1
        $code = $LASTEXITCODE
    } finally {
        Pop-Location
        $ErrorActionPreference = $oldPreference
    }

    return [PSCustomObject]@{
        Code   = $code
        Output = ($output -join "`n")
    }
}

function New-FixtureRepo {
    param([string]$Directory)

    New-Item -ItemType Directory -Force -Path (Join-Path $Directory '.specify/templates') | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $Directory '.specify/scripts/powershell') | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $Directory 'scripts/speckit') | Out-Null
    New-Item -ItemType Directory -Force -Path (Join-Path $Directory 'specs') | Out-Null

    Copy-Item -LiteralPath $tool -Destination (Join-Path $Directory 'scripts/speckit/team-feature.ps1')

    $common = @'
function Resolve-TemplateContent {
    param([string]$TemplateName, [string]$RepoRoot)
    $path = Join-Path $RepoRoot (".specify/templates/{0}.md" -f $TemplateName)
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) { return $null }
    return [System.IO.File]::ReadAllText($path)
}
'@
    $utf8 = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText((Join-Path $Directory '.specify/scripts/powershell/common.ps1'), $common, $utf8)
    [System.IO.File]::WriteAllText((Join-Path $Directory '.specify/templates/spec-template.md'), "# Spec: [FEATURE NAME]`n", $utf8)
    [System.IO.File]::WriteAllText((Join-Path $Directory '.gitignore'), ".specify/owner.json`n.specify/feature.json`n", $utf8)
    [System.IO.File]::WriteAllText((Join-Path $Directory 'README.md'), "fixture`n", $utf8)
}

function Add-Owner {
    param([string]$Directory, [string]$Handle)

    $result = Invoke-Tool -Directory $Directory -Arguments @('owner', '-Handle', $Handle)
    if ($result.Code -ne 0) { throw "owner -Handle $Handle failed: $($result.Output)" }
}

function Add-FeatureCommit {
    param([string]$Directory, [string]$Message)

    Invoke-Git -Directory $Directory -Arguments @('add', '.') | Out-Null
    Invoke-Git -Directory $Directory -Arguments @('commit', '-m', $Message) | Out-Null
}

try {
    New-Item -ItemType Directory -Force -Path $tempRoot | Out-Null

    $remote = Join-Path $tempRoot 'remote.git'
    $null = & git init --bare $remote | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'git init --bare failed' }
    Invoke-Git -Directory $remote -Arguments @('symbolic-ref', 'HEAD', 'refs/heads/main') | Out-Null

    $dev1 = Join-Path $tempRoot 'dev1'
    $null = & git clone $remote $dev1 | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'git clone dev1 failed' }

    New-FixtureRepo -Directory $dev1
    Invoke-Git -Directory $dev1 -Arguments @('config', 'user.email', 'duck@example.test') | Out-Null
    Invoke-Git -Directory $dev1 -Arguments @('config', 'user.name', 'Duck') | Out-Null
    Add-FeatureCommit -Directory $dev1 -Message 'chore: fixture'
    Invoke-Git -Directory $dev1 -Arguments @('branch', '-M', 'main') | Out-Null
    Invoke-Git -Directory $dev1 -Arguments @('push', '-u', 'origin', 'main') | Out-Null

    $dev2 = Join-Path $tempRoot 'dev2'
    $null = & git clone $remote $dev2 | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'git clone dev2 failed' }
    Invoke-Git -Directory $dev2 -Arguments @('config', 'user.email', 'minh@example.test') | Out-Null
    Invoke-Git -Directory $dev2 -Arguments @('config', 'user.name', 'Minh') | Out-Null

    Add-Owner -Directory $dev1 -Handle 'duck'
    Add-Owner -Directory $dev2 -Handle 'minh'

    $onMain = Invoke-Tool -Directory $dev1 -Arguments @('new', '-Slug', 'should-fail', '-Description', 'x')
    Assert-True ($onMain.Code -ne 0) 'new refuses to run on main'

    Invoke-Git -Directory $dev1 -Arguments @('switch', '-c', 'feat/user-auth') | Out-Null
    Invoke-Git -Directory $dev1 -Arguments @('push', '-u', 'origin', 'feat/user-auth') | Out-Null

    $created = Invoke-Tool -Directory $dev1 -Arguments @('new', '-Slug', 'user-auth', '-Description', 'Authentication', '-Json')
    Assert-True ($created.Code -eq 0) "new creates duck's first feature"
    $createdJson = $created.Output | ConvertFrom-Json
    Assert-True ((Split-Path -Leaf $createdJson.FEATURE_DIR) -eq 'duck-001-user-auth') 'new names the directory owner-NNN-slug'
    Assert-True ($createdJson.FEATURE_NUM -eq '001') 'first feature number is 001'
    Assert-True (Test-Path -LiteralPath (Join-Path $dev1 'specs/duck-001-user-auth/meta.json')) 'new writes meta.json'
    Add-FeatureCommit -Directory $dev1 -Message 'feat: spec user-auth'
    Invoke-Git -Directory $dev1 -Arguments @('push') | Out-Null

    $metaPath = Join-Path $dev1 'specs/duck-001-user-auth/meta.json'
    $metaHash = (Get-FileHash -LiteralPath $metaPath -Algorithm SHA256).Hash
    $bound = Invoke-Tool -Directory $dev1 -Arguments @('bind', '-Json')
    Assert-True ($bound.Code -eq 0) 'bind succeeds on the feature branch'
    Assert-True ((Get-FileHash -LiteralPath $metaPath -Algorithm SHA256).Hash -eq $metaHash) 'bind does not rewrite meta.json when the branch is unchanged'

    $list = Invoke-Tool -Directory $dev1 -Arguments @('list', '-Json')
    $listJson = @($list.Output | ConvertFrom-Json)
    Assert-True ($list.Code -eq 0 -and $listJson.Count -eq 1 -and $listJson[0].Owner -eq 'duck') 'list -Json returns a machine-readable array'

    Invoke-Git -Directory $dev1 -Arguments @('switch', 'main') | Out-Null
    $resolveWrong = Invoke-Tool -Directory $dev1 -Arguments @('resolve', '-Json')
    Assert-True ($resolveWrong.Code -ne 0) 'resolve refuses to run on main'

    Invoke-Git -Directory $dev2 -Arguments @('fetch', 'origin') | Out-Null
    Invoke-Git -Directory $dev2 -Arguments @('switch', '-c', 'feat/user-auth', 'origin/feat/user-auth') | Out-Null
    $slugTaken = Invoke-Tool -Directory $dev2 -Arguments @('new', '-Slug', 'user-auth', '-Description', 'Taken')
    Assert-True ($slugTaken.Code -ne 0) 'new rejects a slug already owned by another owner on origin'

    Invoke-Git -Directory $dev2 -Arguments @('switch', 'main') | Out-Null
    Invoke-Git -Directory $dev2 -Arguments @('switch', '-c', 'feat/order-cancel') | Out-Null
    Invoke-Git -Directory $dev2 -Arguments @('push', '-u', 'origin', 'feat/order-cancel') | Out-Null
    $minhCreated = Invoke-Tool -Directory $dev2 -Arguments @('new', '-Slug', 'order-cancel', '-Description', 'Cancel order', '-Json')
    $minhJson = $minhCreated.Output | ConvertFrom-Json
    Assert-True ($minhCreated.Code -eq 0 -and (Split-Path -Leaf $minhJson.FEATURE_DIR) -eq 'minh-001-order-cancel') 'a second owner can also use 001'
    Add-FeatureCommit -Directory $dev2 -Message 'feat: spec order-cancel'
    Invoke-Git -Directory $dev2 -Arguments @('push') | Out-Null

    Invoke-Git -Directory $dev1 -Arguments @('switch', 'main') | Out-Null
    Invoke-Git -Directory $dev1 -Arguments @('pull', '--ff-only') | Out-Null
    Invoke-Git -Directory $dev1 -Arguments @('switch', '-c', 'feat/payment') | Out-Null
    Invoke-Git -Directory $dev1 -Arguments @('push', '-u', 'origin', 'feat/payment') | Out-Null
    $second = Invoke-Tool -Directory $dev1 -Arguments @('new', '-Slug', 'payment', '-Description', 'Payments', '-Json')
    $secondJson = $second.Output | ConvertFrom-Json
    Assert-True ($second.Code -eq 0 -and (Split-Path -Leaf $secondJson.FEATURE_DIR) -eq 'duck-002-payment') 'the same owner gets the next number from remote branches'
    Add-FeatureCommit -Directory $dev1 -Message 'feat: spec payment'
    Invoke-Git -Directory $dev1 -Arguments @('push') | Out-Null

    $cleanCheck = Invoke-Tool -Directory $dev1 -Arguments @('check', '-Strict')
    Assert-True ($cleanCheck.Code -eq 0) 'check -Strict passes a clean repository'

    $duplicateDir = Join-Path $dev1 'specs/duck-002-conflict'
    New-Item -ItemType Directory -Force -Path $duplicateDir | Out-Null
    $duplicateMeta = '{"owner":"duck","number":2,"slug":"conflict","branch":"feat/conflict","kind":"feat","status":"open","services":[],"description":"x","opened_at":"2026-10-22"}'
    $utf8 = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText((Join-Path $duplicateDir 'meta.json'), $duplicateMeta, $utf8)
    $duplicateCheck = Invoke-Tool -Directory $dev1 -Arguments @('check')
    Assert-True ($duplicateCheck.Code -ne 0 -and $duplicateCheck.Output -match 'reuses number') 'check rejects a duplicate owner number'
    Remove-Item -LiteralPath $duplicateDir -Recurse -Force

    $legacyDir = Join-Path $dev1 'specs/legacy-feature'
    New-Item -ItemType Directory -Force -Path $legacyDir | Out-Null
    [System.IO.File]::WriteAllText((Join-Path $legacyDir 'spec.md'), '# legacy`n', $utf8)
    $legacyLoose = Invoke-Tool -Directory $dev1 -Arguments @('check')
    $legacyStrict = Invoke-Tool -Directory $dev1 -Arguments @('check', '-Strict')
    Assert-True ($legacyLoose.Code -eq 0 -and $legacyLoose.Output -match 'does not follow') 'check reports a legacy directory as a warning'
    Assert-True ($legacyStrict.Code -ne 0) 'check -Strict turns a warning into a failure'
    Remove-Item -LiteralPath $legacyDir -Recurse -Force

    $ownerBackup = Join-Path $dev1 '.specify/owner.json'
    Remove-Item -LiteralPath $ownerBackup -Force
    $ciCheck = Invoke-Tool -Directory $dev1 -Arguments @('check', '-Strict')
    Assert-True ($ciCheck.Code -eq 0) 'check -Strict works without a local owner.json (CI)'
    Add-Owner -Directory $dev1 -Handle 'duck'
} catch {
    $script:failed++
    Write-Host "FAIL: unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ScriptStackTrace) { Write-Host $_.ScriptStackTrace }
} finally {
    if (Test-Path -LiteralPath $tempRoot) {
        Remove-Item -LiteralPath $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Write-Host ""
Write-Host "Passed: $script:passed  Failed: $script:failed"
if ($script:failed -gt 0) { exit 1 }
exit 0
