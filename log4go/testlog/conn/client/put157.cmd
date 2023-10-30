@echo off
go build -o testlog.exe testclient.go
echo "1. 编译Linux可执行程序"
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
del testlog
go build -o testlog testclient.go

echo "2. 关闭在 157 机器 testlog"
set run_server=tony@192.168.3.157
ssh %run_server% "pkill testlog"
ssh %run_server% "ps -ef|grep [t]estlog"

echo "3. 发布到 157 机器 testlog"
copy testlog.exe  C:\MyWork\gitcode\application\log4server
ssh %run_server% "rm -rf /opt/bin/testlog"
scp ./testlog %run_server%:/opt/bin/
scp ./testlog %run_server%:/home/share/zcmfiles/3tools/
ssh %run_server% "chmod +x /opt/bin/testlog"
