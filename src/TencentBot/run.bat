@echo off
echo Starting Tencent Bot...
go run main.go > output.log 2>&1
echo Application exited. Output saved to output.log
type output.log
pause