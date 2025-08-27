@echo off

setlocal
echo =========================================================================
echo 开始自动提交代码 ...... tea4go/gh
echo =========================================================================

cd C:\MyWork\GitCode\gh
:: 检查是否有未暂存的修改
git diff --quiet
if %errorlevel% equ 0 (
    :: 检查是否有已暂存但未提交的修改
    git diff --cached --quiet
    if %errorlevel% equ 0 (
        echo 没有代码修改，跳过提交。
        exit /b 0
    )
)

git status
git add .
git commit -m "auto::bugfix"
git push origin dev
git push

echo.
echo 所有 tea4go/gh 码提交完成!
echo.

cd C:\MyWork\GitCode\gh

endlocal