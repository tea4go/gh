# ============================================================
# release_tag.ps1 — 创建带发布说明的 Git Tag 并推送 (Windows PowerShell)
#
# 用法:
#   .\release_tag.ps1                              # 交互模式，自动推断下一个版本号
#   .\release_tag.ps1 v1.4.1                       # 指定版本号
#   .\release_tag.ps1 v1.4.1 "修复并发bug"          # 指定版本号 + 说明
#   .\release_tag.ps1 -List                        # 查看最近 10 个标签
#   .\release_tag.ps1 -DryRun                      # 干跑模式，只预览不执行
#   .\release_tag.ps1 -DryRun v1.5.0 "新增XX功能"   # 干跑 + 指定版本
#
# 说明:
#   - 版本号格式: v主.次.修订 (如 v1.4.1)
#   - 不指定版本号时，基于最新 tag 自动递增修订号
#   - 不指定说明时，自动从 git log 生成变更摘要
# ============================================================

param(
    [string]$Version  = "",
    [string]$Message  = "",
    [switch]$List     = $false,
    [switch]$DryRun   = $false
)

# ─── 切换到脚本所在目录 ────────────────────────────────────────
Set-Location $PSScriptRoot

# ─── 颜色输出 ──────────────────────────────────────────────────
function Write-Info  { Write-Host "[INFO]  $args" -ForegroundColor Green }
function Write-Warn  { Write-Host "[WARN]  $args" -ForegroundColor Yellow }
function Write-Err   { Write-Host "[FAIL]  $args" -ForegroundColor Red }
function Write-Ok    { Write-Host "[PASS]  $args" -ForegroundColor Green }

# ─── 查看标签列表 ──────────────────────────────────────────────
function Show-Tags {
    Write-Host "最近的标签:" -ForegroundColor Cyan
    git tag --sort=-creatordate -l --format "  %(refname:short)  %(creatordate:short)  %(subject)" 2>$null | Select-Object -First 10
}

# ─── 获取最新 tag ──────────────────────────────────────────────
function Get-LatestTag {
    git tag --sort=-version:refname 2>$null | Select-Object -First 1
}

# ─── 自动推断下一个版本号 ──────────────────────────────────────
function Get-NextVersion {
    $latest = Get-LatestTag
    if (-not $latest) { return "v1.0.0" }

    $ver = $latest.TrimStart('v')
    $parts = $ver -split '\.'
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2] + 1
    return "v${major}.${minor}.${patch}"
}

# ─── 生成变更摘要 ──────────────────────────────────────────────
function Get-Changelog {
    param([string]$FromTag)

    if (-not $FromTag) {
        $log = git log --oneline -20 --no-decorate 2>$null
        if ($log) { return $log } else { return "(无提交记录)" }
    }

    $log = git log "${FromTag}..HEAD" --oneline --no-decorate 2>$null
    if ($log) { return $log } else { return "(无新提交)" }
}

# ─── 生成 tag message ──────────────────────────────────────────
function New-TagMessage {
    param([string]$Ver, [string]$FromTag, [string]$CustomMsg)

    $dateStr = Get-Date -Format "yyyy-MM-dd"
    $lines = @("${Ver} (${dateStr})", "")

    if ($CustomMsg) {
        $lines += $CustomMsg
    } else {
        $lines += "变更:"
        $lines += (Get-Changelog $FromTag)
    }

    return $lines -join "`n"
}

# ─── 查看标签 ──────────────────────────────────────────────────
if ($List) {
    Show-Tags
    exit 0
}

# ─── 自动推断版本号 ────────────────────────────────────────────
if (-not $Version) {
    $Version = Get-NextVersion
    Write-Info "自动推断版本号: $Version"
}

# ─── 校验版本号格式 ────────────────────────────────────────────
if ($Version -notmatch '^v\d+\.\d+\.\d+$') {
    Write-Err "版本号格式错误: $Version，应为 v主.次.修订 (如 v1.4.1)"
    exit 1
}

# ─── 检查 tag 是否已存在 ───────────────────────────────────────
$existing = git tag -l $Version 2>$null
if ($existing) {
    Write-Err "标签 $Version 已存在！"
    Show-Tags
    exit 1
}

# ─── 检查是否有未提交的修改 ────────────────────────────────────
$hasChanges = $false
if (git diff --quiet 2>$null)  { } else { $hasChanges = $true }
if (git diff --cached --quiet 2>$null) { } else { $hasChanges = $true }

if ($hasChanges) {
    Write-Warn "存在未提交的修改，建议先提交！"
    git status --short
    Write-Host ""
    $confirm = Read-Host "仍要继续打标签吗？[y/N]"
    if ($confirm -notmatch '^[Yy]') {
        Write-Info "已取消"
        exit 0
    }
}

# ─── 获取信息 ──────────────────────────────────────────────────
$latestTag  = Get-LatestTag
$headCommit = git rev-parse --short HEAD 2>$null
$dateStr    = Get-Date -Format "yyyy-MM-dd"
$tagMsg     = New-TagMessage $Version $latestTag $Message

# ─── 预览 ──────────────────────────────────────────────────────
Write-Host ""
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  发布标签预览" -ForegroundColor White
Write-Host "══════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  版本号:   $Version" -ForegroundColor Green
if ($latestTag) {
    Write-Host "  基于:     $latestTag"
} else {
    Write-Host "  基于:     无(首次)"
}
Write-Host "  Commit:   $headCommit"
Write-Host "  日期:     $dateStr"
Write-Host ""
Write-Host "  Tag Message:"
Write-Host "  ─────────────────────────────────────────────────"
$tagMsg -split "`n" | ForEach-Object { Write-Host "  $_" }
Write-Host "  ─────────────────────────────────────────────────"
Write-Host ""

if ($DryRun) {
    Write-Info "干跑模式，未执行任何操作"
    exit 0
}

# ─── 确认 ──────────────────────────────────────────────────────
$confirm = Read-Host "确认创建并推送标签 $Version？[y/N]"
if ($confirm -notmatch '^[Yy]') {
    Write-Info "已取消"
    exit 0
}

# ─── 执行 ──────────────────────────────────────────────────────
Write-Info "创建标签 $Version ..."
git tag -a $Version -m $tagMsg

Write-Info "推送标签到 origin ..."
git push origin $Version

Write-Host ""
Write-Ok "标签 $Version 发布成功！"
Show-Tags
