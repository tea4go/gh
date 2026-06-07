@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul 2>&1

REM ============================================================
REM run_tests.bat — gh 项目 Windows 全量回归测试脚本
REM
REM 用法:
REM   run_tests.bat              默认：测试 + 覆盖率 + 汇总报告
REM   run_tests.bat --race       额外启用 race 检测
REM   run_tests.bat --quick      跳过覆盖率，仅跑测试
REM   run_tests.bat --html       生成 HTML 覆盖率报告
REM   run_tests.bat --help       显示帮助
REM
REM 环境变量:
REM   CGO_ENABLED    默认 1（image 包需要 cgo）
REM
REM .env 文件:
REM   脚本自动加载项目根目录的 .env 文件（已 gitignore）。
REM   首次使用: copy .env.example .env && 编辑填入凭证
REM ============================================================

REM ─── 配置 ──────────────────────────────────────────────────
set "ROOT_DIR=%~dp0"
cd /d "%ROOT_DIR%"
set "ROOT_DIR=%cd%"

if not defined CGO_ENABLED set "CGO_ENABLED=1"
set "COVER_THRESHOLD=90"
set "RACE_MODE=0"
set "QUICK_MODE=0"
set "HTML_MODE=0"
set "FAIL_N=0"
set "PASS_N=0"
set "SKIP_N=0"
set "TOTAL=0"

REM ─── 参数解析 ──────────────────────────────────────────────
for %%a in (%*) do (
    if "%%a"=="--race"  set "RACE_MODE=1"
    if "%%a"=="--quick" set "QUICK_MODE=1"
    if "%%a"=="--html"  set "HTML_MODE=1"
    if "%%a"=="--help"  goto :show_help
    if "%%a"=="-h"      goto :show_help
)

REM ─── 加载 .env ─────────────────────────────────────────────
if exist "%ROOT_DIR%\.env" (
    call :log_info "加载 .env 文件"
    for /f "usebackq tokens=1,* delims==" %%k in ("%ROOT_DIR%\.env") do (
        set "env_key=%%k"
        set "env_val=%%l"
        REM 跳过注释行和空行
        if not "!env_key!"=="" (
            set "first_char=!env_key:~0,1!"
            if not "!first_char!"=="#" (
                REM 仅在环境变量未设置时才从 .env 注入（命令行优先）
                if not defined !env_key! (
                    set "!env_key!=!env_val!"
                )
            )
        )
    )
) else (
    call :log_info "未找到 .env 文件（外部服务依赖的测试将 Skip）"
    call :log_info "如需启用: copy .env.example .env && 编辑 .env 填入凭证"
)

REM ─── 1. 编译检查 ────────────────────────────────────────────
call :log_info "编译检查 go build ./..."
go build ./...
if errorlevel 1 (
    call :log_fail "编译失败，终止测试"
    exit /b 1
)
call :log_ok "编译通过"

REM ─── 2. go vet ──────────────────────────────────────────────
call :log_info "静态分析 go vet ./..."
go vet ./... >nul 2>&1
if errorlevel 1 (
    call :log_warn "go vet 发现问题（不阻塞测试）"
) else (
    call :log_ok "go vet 通过"
)

REM ─── 3. 准备覆盖率目录 ──────────────────────────────────────
if "%QUICK_MODE%"=="0" (
    if exist "%ROOT_DIR%\.coverage" rmdir /s /q "%ROOT_DIR%\.coverage"
    mkdir "%ROOT_DIR%\.coverage"
)

REM ─── 4. 定义包列表 ──────────────────────────────────────────
REM 清空结果文件
if exist "%ROOT_DIR%\.coverage\results.txt" del "%ROOT_DIR%\.coverage\results.txt"
if exist "%ROOT_DIR%\.coverage\fail_pkgs.txt" del "%ROOT_DIR%\.coverage\fail_pkgs.txt"

REM ─── 5. 逐包测试 ────────────────────────────────────────────
call :log_info "开始全量测试"
echo.

REM === 纯逻辑包 ===
call :log_bold "[纯逻辑]"
call :test_pkg "utils"                       "纯逻辑"
call :test_pkg "syslog"                      "纯逻辑"
call :test_pkg "syslog/format"               "纯逻辑"
call :test_pkg "syslog/internal/syslogparser"         "纯逻辑"
call :test_pkg "syslog/internal/syslogparser/rfc3164" "纯逻辑"
call :test_pkg "syslog/internal/syslogparser/rfc5424" "纯逻辑"
call :test_pkg "radius"                      "纯逻辑"
call :test_pkg "log4go"                      "纯逻辑"
call :test_pkg "image"                       "纯逻辑"
call :test_pkg "tcping/ping"                 "纯逻辑"
call :test_pkg "tcping/ping/http"            "纯逻辑"
call :test_pkg "tcping/ping/tcp"             "纯逻辑"
call :test_pkg "execplus"                    "纯逻辑"
call :test_pkg "fsnotify"                    "纯逻辑"
call :test_pkg "timewheel"                   "纯逻辑"
call :test_pkg "ldapserver"                  "纯逻辑"
echo.

REM === 接近目标包 ===
call :log_bold "[接近目标]"
call :test_pkg "network"                     "接近目标"
call :test_pkg "nustdbclient"                "接近目标"
echo.

REM === 外部服务依赖包 ===
call :log_bold "[外部依赖]"
call :test_pkg "ldapclient"                  "外部依赖"
call :test_pkg "dingtalk"                    "外部依赖"
call :test_pkg "etcdclient"                  "外部依赖"
call :test_pkg "redisclient"                 "外部依赖"
call :test_pkg "weixin"                      "外部依赖"
echo.

REM === 辅助/示例包 ===
call :log_bold "[辅助/示例]"
call :test_pkg "log4go/testlog/conn/client"  "辅助/示例"
call :test_pkg "log4go/testlog/conn/server"  "辅助/示例"
call :test_pkg "log4go/testlog/console"      "辅助/示例"
call :test_pkg "log4go/testlog/logger"       "辅助/示例"
call :test_pkg "dingtalk/test_dingtalk_app"  "辅助/示例"
call :test_pkg "execplus/example"            "辅助/示例"
call :test_pkg "timewheel/gtype"             "辅助/示例"
echo.

REM ─── 6. 合并覆盖率 ──────────────────────────────────────────
if "%QUICK_MODE%"=="0" (
    call :log_info "合并覆盖率数据..."
    set "COVERPROFILE=%ROOT_DIR%\.coverage\cover.out"
    echo mode: set > "!COVERPROFILE!"
    for %%f in ("%ROOT_DIR%\.coverage\*.out") do (
        REM 跳过 cover.out 本身
        if not "%%f"=="!COVERPROFILE!" (
            REM 跳过第一行 mode 行，追加其余内容
            for /f "usebackq skip=1" %%l in ("%%f") do echo %%l >> "!COVERPROFILE!"
        )
    )

    if "%HTML_MODE%"=="1" (
        call :log_info "生成 HTML 覆盖率报告..."
        go tool cover -html="!COVERPROFILE!" -o "%ROOT_DIR%\.coverage\cover.html"
        call :log_ok "HTML 报告: %ROOT_DIR%\.coverage\cover.html"
    )
)

REM ─── 7. 汇总报告 ────────────────────────────────────────────
echo.
call :log_bold "═══════════════════════════════════════════════════════════════"
call :log_bold "                       测试汇总报告"
call :log_bold "═══════════════════════════════════════════════════════════════"
echo.

REM 按分组从结果文件输出
set "logic_pass=0"
set "logic_total=0"
for /f "usebackq tokens=1-4 delims=|" %%p %%g %%r %%c in ("%ROOT_DIR%\.coverage\results.txt") do (
    if "%%g"=="纯逻辑" (
        set /a logic_total+=1
        if not "%%c"=="---" (
            REM 提取整数部分
            for /f "tokens=1 delims=." %%i in ("%%c") do (
                if %%i GEQ %COVER_THRESHOLD% set /a logic_pass+=1
            )
        )
    )
)

echo 统计:  总计 %TOTAL%  通过 %PASS_N%  失败 %FAIL_N%  跳过 %SKIP_N%
echo.

if "%QUICK_MODE%"=="0" (
    echo 纯逻辑包达 %COVER_THRESHOLD%%:  %logic_pass%/%logic_total%

    REM 总体覆盖率
    set "total_pct=N/A"
    if exist "!COVERPROFILE!" (
        for /f "usebackq tokens=1,2" %%a %%b in ('go tool cover -func="!COVERPROFILE!" 2^>nul ^| findstr /r "total:"') do (
            set "total_pct=%%b"
        )
    )
    echo 总体覆盖率:  %total_pct%
    echo.
)

REM 失败包列表
if exist "%ROOT_DIR%\.coverage\fail_pkgs.txt" (
    call :log_fail "失败包:"
    for /f "usebackq" %%p in ("%ROOT_DIR%\.coverage\fail_pkgs.txt") do (
        echo   X %%p
    )
    echo.
)

REM ─── 8. 退出码 ──────────────────────────────────────────────
if %FAIL_N% gtr 0 (
    exit /b 1
) else (
    call :log_ok "全部通过"
    exit /b 0
)

REM ═══════════════════════════════════════════════════════════════
REM  子程序
REM ═══════════════════════════════════════════════════════════════

:test_pkg
REM 参数: %1=short_name  %2=group
set "PKG_SHORT=%~1"
set "PKG_GROUP=%~2"
set "PKG_FULL=github.com/tea4go/gh/%PKG_SHORT%"
set /a TOTAL+=1

REM 构建 go test 命令
set "TEST_CMD=go test -v -count=1 -timeout 120s"
if "%RACE_MODE%"=="1" set "TEST_CMD=%TEST_CMD% -race"
if "%QUICK_MODE%"=="0" (
    set "PKG_SLUG=%PKG_SHORT:/=_%"
    set "TEST_CMD=%TEST_CMD% -coverprofile=%ROOT_DIR%\.coverage\%PKG_SLUG%.out"
)
set "TEST_CMD=%TEST_CMD% %PKG_FULL%"

REM 执行测试，捕获输出到临时文件
set "TMP_OUT=%ROOT_DIR%\.coverage\tmp_%PKG_SLUG%.txt"
%TEST_CMD% > "%TMP_OUT%" 2>&1
set "TEST_EXIT=!errorlevel!"

REM 判断结果
set "PKG_RESULT=SKIP"
set "PKG_COV=---"

REM 读取输出判断 PASS/FAIL
findstr /c:"build failure" "%TMP_OUT%" >nul 2>&1
if not errorlevel 1 (
    set "PKG_RESULT=FAIL"
) else (
    findstr /b /c:"FAIL" "%TMP_OUT%" >nul 2>&1
    if not errorlevel 1 (
        set "PKG_RESULT=FAIL"
    ) else (
        findstr /b /c:"ok " "%TMP_OUT%" >nul 2>&1
        if not errorlevel 1 (
            set "PKG_RESULT=PASS"
        )
    )
)

REM 提取覆盖率
if "%PKG_RESULT%"=="PASS" (
    for /f "tokens=2,3" %%a %%b in ('findstr /c:"coverage:" "%TMP_OUT%"') do (
        REM %%a=coverage: %%b=XX.X% (去掉末尾的%)
        set "raw_cov=%%b"
        set "PKG_COV=!raw_cov:%%=!"
    )
)

REM 统计
if "%PKG_RESULT%"=="PASS" (
    set /a PASS_N+=1
    if "%PKG_COV%"=="---" (
        call :log_ok "%PKG_SHORT%  [无测试文件]  [%PKG_GROUP%]"
    ) else (
        for /f "tokens=1 delims=." %%i in ("%PKG_COV%") do set "cov_int=%%i"
        if !cov_int! GEQ %COVER_THRESHOLD% (
            call :log_ok "%PKG_SHORT%  %PKG_COV%%%  [%PKG_GROUP%]"
        ) else if !cov_int! GEQ 80 (
            call :log_warn "%PKG_SHORT%  %PKG_COV%%%  [%PKG_GROUP%]（未达 %COVER_THRESHOLD%%%）"
        ) else (
            call :log_warn "%PKG_SHORT%  %PKG_COV%%%  [%PKG_GROUP%]（外部依赖/结构性限制）"
        )
    )
) else if "%PKG_RESULT%"=="FAIL" (
    set /a FAIL_N+=1
    call :log_fail "%PKG_SHORT%"
    echo %PKG_SHORT%>> "%ROOT_DIR%\.coverage\fail_pkgs.txt"
) else (
    set /a SKIP_N+=1
    call :log_warn "%PKG_SHORT%  [跳过]"
)

REM 记录结果到文件（用于汇总）
echo %PKG_SHORT%|%PKG_GROUP%|%PKG_RESULT%|%PKG_COV%>> "%ROOT_DIR%\.coverage\results.txt"

REM 清理临时文件
if exist "%TMP_OUT%" del "%TMP_OUT%"
goto :eof

:show_help
echo run_tests.bat — gh 项目 Windows 全量回归测试脚本
echo.
echo 用法:
echo   run_tests.bat              默认：测试 + 覆盖率 + 汇总报告
echo   run_tests.bat --race       额外启用 race 检测
echo   run_tests.bat --quick      跳过覆盖率，仅跑测试
echo   run_tests.bat --html       生成 HTML 覆盖率报告
echo.
echo .env 文件:
echo   脚本自动加载项目根目录的 .env 文件。
echo   首次使用: copy .env.example .env && 编辑填入凭证
echo.
echo 退出码: 0=全通过  1=有失败
exit /b 0

REM ─── 日志函数 ────────────────────────────────────────────────
:log_info
echo [INFO]  %~1
goto :eof

:log_ok
echo [PASS]  %~1
goto :eof

:log_warn
echo [WARN]  %~1
goto :eof

:log_fail
echo [FAIL]  %~1
goto :eof

:log_bold
echo %~1
goto :eof