@echo off
chcp 65001
cls
set log_fdebug=
set log_server=
set log_level=
del logger.exe
go build .
if errorlevel 1 (
    echo 编译失败，请检查错误信息。
    exit /b 1
)

logger.exe abc
