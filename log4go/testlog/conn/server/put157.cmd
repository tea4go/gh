@echo off
echo "1. 编译Linux可执行程序"
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
del log4server
go build -o log4server testserver.go

echo "2. 关闭在 157 机器 log4server"
set run_server=tony@192.168.100.157
ssh %run_server% "pkill log4server"
ssh %run_server% "ps -ef|grep [l]og4server"

echo "3. 发布到 157 机器 log4server"
ssh %run_server% "rm -rf /opt/bin/log4server"
scp ./log4server %run_server%:/opt/bin/
ssh %run_server% "chmod +x /opt/bin/log4server"
