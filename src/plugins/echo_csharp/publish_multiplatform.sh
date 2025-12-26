#!/bin/bash
# 多平台发布脚本

# Windows
cd src/plugins/echo_csharp
dotnet publish -c Release -r win-x64 --self-contained true
cp bin/Release/net6.0/win-x64/publish/echo_csharp.exe ../../plugins/echo_csharp/echo_csharp.exe

# Linux
dotnet publish -c Release -r linux-x64 --self-contained true
cp bin/Release/net6.0/linux-x64/publish/echo_csharp ../../plugins/echo_csharp/echo_csharp_linux

# macOS
dotnet publish -c Release -r osx-x64 --self-contained true
cp bin/Release/net6.0/osx-x64/publish/echo_csharp ../../plugins/echo_csharp/echo_csharp_macos

echo "多平台发布完成"