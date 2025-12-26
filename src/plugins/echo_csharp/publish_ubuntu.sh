#!/bin/bash
# Ubuntu 发布脚本

# 发布Linux版本
dotnet publish -c Release -r linux-x64 --self-contained true

# 复制到插件目录
cp bin/Release/net6.0/linux-x64/publish/echo_csharp ../../plugins/echo_csharp/

# 设置执行权限
chmod +x ../../plugins/echo_csharp/echo_csharp

echo "C#插件已成功发布到Ubuntu平台"