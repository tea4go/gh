@echo off
chcp 65001
cls

echo "1. 编译Linux可执行程序"
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
del testlog >nul 2>nul
del testlog.exe >nul 2>nul
go build -o testlog testclient.go

echo "2. 关闭在 157 机器 testlog"
set run_server=tony@192.168.100.1
set run_port=40026
ssh -p %run_port% %run_server% "pkill testlog"
ssh -p %run_port% %run_server% "ps -ef|grep [t]estlog"

echo "3. 发布到 157 机器 testlog"
ssh -p %run_port% %run_server% "rm -rf /opt/bin/testlog"
scp -P %run_port% ./testlog %run_server%:/opt/bin/
ssh -p %run_port% %run_server% "chmod +x /opt/bin/testlog"
