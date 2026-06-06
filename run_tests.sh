#!/usr/bin/env bash
#
# run_tests.sh — gh 项目全量回归测试脚本
#
# 用法:
#   ./run_tests.sh              # 默认：测试 + 覆盖率 + 汇总报告
#   ./run_tests.sh --race       # 额外启用 race 检测
#   ./run_tests.sh --quick      # 跳过覆盖率，仅跑测试
#   ./run_tests.sh --html       # 生成 HTML 覆盖率报告到 cover.html
#
# 环境变量:
#   CGO_ENABLED   默认 1（image 包需要 cgo）
#   GO_TEST_FLAGS  额外的 go test 参数
#

set -euo pipefail

# ─── 颜色 ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[0;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; RST='\033[0m'

# ─── 配置 ──────────────────────────────────────────────────────────────────
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

CGO_ENABLED="${CGO_ENABLED:-1}"
export CGO_ENABLED

COVERAGE_DIR="$ROOT_DIR/.coverage"
COVERPROFILE="$COVERAGE_DIR/cover.out"
COVER_THRESHOLD=90          # 目标覆盖率 %
RACE_MODE=false
QUICK_MODE=false
HTML_MODE=false

# ─── 参数解析 ──────────────────────────────────────────────────────────────
for arg in "$@"; do
    case "$arg" in
        --race)  RACE_MODE=true  ;;
        --quick) QUICK_MODE=true ;;
        --html)  HTML_MODE=true  ;;
        -h|--help)
            head -15 "$0" | grep '^#' | sed 's/^# \?//'
            exit 0
            ;;
        *)
            echo -e "${RED}未知参数: $arg${RST}" >&2; exit 1 ;;
    esac
done

# ─── 工具函数 ───────────────────────────────────────────────────────────────
info()  { echo -e "${BLUE}[INFO]${RST}  $*"; }
ok()    { echo -e "${GREEN}[PASS]${RST}  $*"; }
warn()  { echo -e "${YELLOW}[WARN]${RST}  $*"; }
fail()  { echo -e "${RED}[FAIL]${RST}  $*"; }

# ─── 1. 编译检查 ───────────────────────────────────────────────────────────
info "编译检查 go build ./..."
if ! go build ./...; then
    fail "编译失败，终止测试"
    exit 1
fi
ok "编译通过"

# ─── 2. go vet ─────────────────────────────────────────────────────────────
info "静态分析 go vet ./..."
vet_ok=true
go vet ./... 2>&1 || vet_ok=false
if $vet_ok; then
    ok "go vet 通过"
else
    warn "go vet 发现问题（不阻塞测试）"
fi

# ─── 3. 获取包列表 ─────────────────────────────────────────────────────────
# 按功能分组，方便阅读报告
declare -A PKG_GROUP  # pkg -> group
ALL_PKGS=()

# 纯逻辑包（≥90%）
for p in utils syslog syslog/format \
         syslog/internal/syslogparser \
         syslog/internal/syslogparser/rfc3164 \
         syslog/internal/syslogparser/rfc5424 \
         radius log4go image \
         tcping/ping tcping/ping/http tcping/ping/tcp \
         execplus fsnotify timewheel ldapserver; do
    ALL_PKGS+=("github.com/tea4go/gh/$p")
    PKG_GROUP["github.com/tea4go/gh/$p"]="纯逻辑"
done

# 接近目标包（80%~90%）
for p in network nustdbclient; do
    ALL_PKGS+=("github.com/tea4go/gh/$p")
    PKG_GROUP["github.com/tea4go/gh/$p"]="接近目标"
done

# 外部服务依赖包
for p in ldapclient dingtalk etcdclient redisclient weixin; do
    ALL_PKGS+=("github.com/tea4go/gh/$p")
    PKG_GROUP["github.com/tea4go/gh/$p"]="外部依赖"
done

# 辅助/示例包（不统计覆盖率）
for p in log4go/testlog/conn/client log4go/testlog/conn/server \
         log4go/testlog/console log4go/testlog/logger \
         dingtalk/test_dingtalk_app execplus/example timewheel/gtype; do
    ALL_PKGS+=("github.com/tea4go/gh/$p")
    PKG_GROUP["github.com/tea4go/gh/$p"]="辅助/示例"
done

# ─── 4. 准备覆盖率目录 ────────────────────────────────────────────────────
if ! $QUICK_MODE; then
    rm -rf "$COVERAGE_DIR"
    mkdir -p "$COVERAGE_DIR"
fi

# ─── 5. 逐包测试 ───────────────────────────────────────────────────────────
info "开始全量测试（共 ${#ALL_PKGS[@]} 个包）"
echo ""

declare -A PKG_RESULT   # pkg -> PASS/FAIL/SKIP
declare -A PKG_COVERAGE # pkg -> 覆盖率字符串
FAIL_PKGS=()
TOTAL=0; PASS_N=0; FAIL_N=0; SKIP_N=0

for pkg in "${ALL_PKGS[@]}"; do
    TOTAL=$((TOTAL + 1))
    short="${pkg#github.com/tea4go/gh/}"
    group="${PKG_GROUP[$pkg]}"

    # 构建 go test 命令
    cmd="go test -v -count=1 -timeout 120s"
    if $RACE_MODE; then
        cmd="$cmd -race"
    fi
    if ! $QUICK_MODE; then
        pkg_profile="$COVERAGE_DIR/$(echo "$short" | tr '/' '_').out"
        cmd="$cmd -coverprofile=$pkg_profile"
    fi
    cmd="$cmd $pkg"

    # 执行
    output=$(eval "$cmd" 2>&1) || true

    # 判断结果
    if echo "$output" | grep -q "build failure"; then
        FAIL_PKGS+=("$short")
        PKG_RESULT[$pkg]="FAIL"
        FAIL_N=$((FAIL_N + 1))
        fail "$short  [构建失败]"
    elif echo "$output" | grep -q "^FAIL"; then
        FAIL_PKGS+=("$short")
        PKG_RESULT[$pkg]="FAIL"
        FAIL_N=$((FAIL_N + 1))
        fail "$short"
    elif echo "$output" | grep -q "^ok"; then
        PKG_RESULT[$pkg]="PASS"
        PASS_N=$((PASS_N + 1))

        # 提取覆盖率
        cov=$(echo "$output" | grep -oP 'coverage:\s+\K[\d.]+(?=%)' || true)
        if [ -n "$cov" ]; then
            PKG_COVERAGE[$pkg]="$cov"
            cov_int=$(echo "$cov" | awk '{printf "%d", $1}')
            if [ "$cov_int" -ge "$COVER_THRESHOLD" ]; then
                ok "$short  ${cov}%  [$group]"
            elif [ "$cov_int" -ge 80 ]; then
                warn "$short  ${cov}%  [$group]（未达 ${COVER_THRESHOLD}%）"
            else
                warn "$short  ${cov}%  [$group]（外部依赖/结构性限制）"
            fi
        else
            ok "$short  [无测试文件]  [$group]"
        fi
    else
        PKG_RESULT[$pkg]="SKIP"
        SKIP_N=$((SKIP_N + 1))
        warn "$short  [跳过]"
    fi
done

# ─── 6. 合并覆盖率 ────────────────────────────────────────────────────────
if ! $QUICK_MODE; then
    info "合并覆盖率数据..."
    # 使用 gocovmerge 或手动拼接 profile
    echo "mode: set" > "$COVERPROFILE"
    for f in "$COVERAGE_DIR"/*.out; do
        [ -f "$f" ] || continue
        # 跳过第一行 mode 行
        tail -n +2 "$f" >> "$COVERPROFILE" 2>/dev/null || true
    done

    # HTML 报告
    if $HTML_MODE; then
        info "生成 HTML 覆盖率报告..."
        go tool cover -html="$COVERPROFILE" -o "$COVERAGE_DIR/cover.html"
        ok "HTML 报告: $COVERAGE_DIR/cover.html"
    fi
fi

# ─── 7. 汇总报告 ───────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}═══════════════════════════════════════════════════════════════${RST}"
echo -e "${BOLD}                       测试汇总报告${RST}"
echo -e "${BOLD}═══════════════════════════════════════════════════════════════${RST}"
echo ""

# 分组输出
for group_name in "纯逻辑" "接近目标" "外部依赖" "辅助/示例"; do
    echo -e "${BOLD}[$group_name]${RST}"
    for pkg in "${ALL_PKGS[@]}"; do
        [ "${PKG_GROUP[$pkg]}" = "$group_name" ] || continue
        short="${pkg#github.com/tea4go/gh/}"
        result="${PKG_RESULT[$pkg]}"
        cov="${PKG_COVERAGE[$pkg]:---}"

        case "$result" in
            PASS) tag="${GREEN}✓${RST}" ;;
            FAIL) tag="${RED}✗${RST}" ;;
            SKIP) tag="${YELLOW}○${RST}" ;;
        esac

        # 格式化覆盖率
        if [ "$cov" != "---" ]; then
            cov_int=$(echo "$cov" | awk '{printf "%d", $1}')
            if [ "$cov_int" -ge "$COVER_THRESHOLD" ]; then
                cov_display="${GREEN}${cov}%${RST}"
            elif [ "$cov_int" -ge 80 ]; then
                cov_display="${YELLOW}${cov}%${RST}"
            else
                cov_display="${RED}${cov}%${RST}"
            fi
        else
            cov_display="${YELLOW}N/A${RST}"
        fi

        printf "  %s %-30s %s\n" "$tag" "$short" "$cov_display"
    done
    echo ""
done

# 统计
echo -e "${BOLD}统计:${RST}  总计 $TOTAL  通过 $PASS_N  失败 $FAIL_N  跳过 $SKIP_N"
echo ""

# 纯逻辑包覆盖率统计
if ! $QUICK_MODE; then
    logic_total=0; logic_pass=0
    for pkg in "${ALL_PKGS[@]}"; do
        [ "${PKG_GROUP[$pkg]}" = "纯逻辑" ] || continue
        cov="${PKG_COVERAGE[$pkg]:---}"
        [ "$cov" = "---" ] && continue
        logic_total=$((logic_total + 1))
        cov_int=$(echo "$cov" | awk '{printf "%d", $1}')
        [ "$cov_int" -ge "$COVER_THRESHOLD" ] && logic_pass=$((logic_pass + 1))
    done
    if [ "$logic_total" -gt 0 ]; then
        echo -e "纯逻辑包达 ${COVER_THRESHOLD}%:  ${BOLD}${logic_pass}/${logic_total}${RST}"
    fi

    # 总体覆盖率
    total_pct=$(go tool cover -func="$COVERPROFILE" 2>/dev/null | tail -1 | grep -oP '[\d.]+(?=%)' || echo "N/A")
    echo -e "总体覆盖率:  ${BOLD}${total_pct}%${RST}"
    echo ""
fi

# 失败包列表
if [ ${#FAIL_PKGS[@]} -gt 0 ]; then
    echo -e "${RED}${BOLD}失败包:${RST}"
    for p in "${FAIL_PKGS[@]}"; do
        echo -e "  ${RED}✗ $p${RST}"
    done
    echo ""
fi

# ─── 8. 退出码 ─────────────────────────────────────────────────────────────
if [ "$FAIL_N" -gt 0 ]; then
    exit 1
else
    echo -e "${GREEN}${BOLD}全部通过 ✓${RST}"
    exit 0
fi
