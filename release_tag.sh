#!/bin/bash
#
# release_tag.sh — 创建带发布说明的 Git Tag 并推送
#
# 用法:
#   ./release_tag.sh                     # 交互模式，自动推断下一个版本号
#   ./release_tag.sh v1.4.1              # 指定版本号
#   ./release_tag.sh v1.4.1 "修复并发bug" # 指定版本号 + 说明
#   ./release_tag.sh -l                  # 查看最近 10 个标签
#   ./release_tag.sh -d                  # 干跑模式，只预览不执行
#
# 说明:
#   - 版本号格式: v主.次.修订 (如 v1.4.1)
#   - 不指定版本号时，基于最新 tag 自动递增修订号
#   - 不指定说明时，自动从 git log 生成变更摘要
#

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$REPO_DIR"

# ─── 颜色 ───
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ─── 查看标签列表 ───
list_tags() {
    echo -e "${CYAN}最近的标签:${NC}"
    git tag --sort=-creatordate -l --format='  %(refname:short)  %(creatordate:short)  %(subject)' | head -10
}

# ─── 获取最新 tag ───
get_latest_tag() {
    git tag --sort=-version:refname | head -1
}

# ─── 自动推断下一个版本号 ───
next_version() {
    local latest
    latest=$(get_latest_tag)
    if [ -z "$latest" ]; then
        echo "v1.0.0"
        return
    fi
    # 去掉 v 前缀
    local ver="${latest#v}"
    local major minor patch
    IFS='.' read -r major minor patch <<< "$ver"
    patch=$((patch + 1))
    echo "v${major}.${minor}.${patch}"
}

# ─── 从 git log 自动生成变更摘要 ───
generate_changelog() {
    local from_tag="$1"
    if [ -z "$from_tag" ]; then
        # 没有 tag 时取最近 20 条
        git log --oneline -20 --no-decorate
        return
    fi
    local range="${from_tag}..HEAD"
    local count
    count=$(git log "$range" --oneline | wc -l)
    if [ "$count" -eq 0 ]; then
        echo "(无新提交)"
        return
    fi
    git log "$range" --oneline --no-decorate
}

# ─── 生成 tag message（带变更摘要）───
generate_tag_message() {
    local version="$1"
    local from_tag="$2"
    local custom_msg="$3"

    local date_str
    date_str=$(date +%Y-%m-%d)

    if [ -n "$custom_msg" ]; then
        echo "${version} (${date_str})"$'\n\n'"${custom_msg}"
        return
    fi

    local changelog
    changelog=$(generate_changelog "$from_tag")

    local msg="${version} (${date_str})"
    msg+=$'\n\n'"变更:"
    msg+=$'\n'"${changelog}"
    echo "$msg"
}

# ─── 主流程 ───
DRY_RUN=false

case "${1:-}" in
    -l|--list)
        list_tags
        exit 0
        ;;
    -d|--dry-run)
        DRY_RUN=true
        shift || true
        ;;
    -h|--help)
        echo "用法: $0 [-l|-d|-h] [版本号] [说明]"
        echo "  -l  查看最近标签"
        echo "  -d  干跑模式（只预览不执行）"
        echo "  -h  显示帮助"
        exit 0
        ;;
esac

# 解析参数
VERSION="${1:-}"
CUSTOM_MSG="${2:-}"

if [ -z "$VERSION" ]; then
    VERSION=$(next_version)
    info "自动推断版本号: ${BOLD}${VERSION}${NC}"
fi

# 校验版本号格式
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "版本号格式错误: ${VERSION}，应为 v主.次.修订 (如 v1.4.1)"
    exit 1
fi

# 检查 tag 是否已存在
if git tag -l "$VERSION" | grep -q .; then
    error "标签 ${VERSION} 已存在！"
    list_tags
    exit 1
fi

# 检查是否有未提交的修改
if ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
    warn "存在未提交的修改，建议先提交！"
    git status --short
    echo ""
    read -rp "仍要继续打标签吗？[y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy] ]]; then
        info "已取消"
        exit 0
    fi
fi

# 生成 tag message
LATEST_TAG=$(get_latest_tag)
TAG_MSG=$(generate_tag_message "$VERSION" "$LATEST_TAG" "$CUSTOM_MSG")

# ─── 预览 ───
echo ""
echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
echo -e "${BOLD}  发布标签预览${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"
echo -e "  版本号:   ${GREEN}${BOLD}${VERSION}${NC}"
echo -e "  基于:     ${LATEST_TAG:-无}(首次)}"
echo -e "  Commit:   $(git rev-parse --short HEAD)"
echo -e "  日期:     $(date +%Y-%m-%d)"
echo ""
echo -e "  Tag Message:"
echo -e "  ─────────────────────────────────────────────────"
while IFS= read -r line; do
    echo -e "  ${line}"
done <<< "$TAG_MSG"
echo -e "  ─────────────────────────────────────────────────"
echo ""

if $DRY_RUN; then
    info "干跑模式，未执行任何操作"
    exit 0
fi

# ─── 确认 ───
read -rp "确认创建并推送标签 ${VERSION}？[y/N] " confirm
if [[ ! "$confirm" =~ ^[Yy] ]]; then
    info "已取消"
    exit 0
fi

# ─── 执行 ───
info "创建标签 ${VERSION} ..."
git tag -a "$VERSION" -m "$TAG_MSG"

info "推送标签到 origin ..."
git push origin "$VERSION"

echo ""
echo -e "${GREEN}${BOLD}✓ 标签 ${VERSION} 发布成功！${NC}"
list_tags
