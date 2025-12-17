#!/bin/bash

echo "=== WebSocket连接诊断 ==="
echo "测试连接: ws://192.168.0.115:3001/ws/bots"
echo ""

echo "1. 测试网络连通性 (ping)"
ping -c 3 192.168.0.115
echo ""

echo "2. 测试端口开放 (telnet)"
timeout 5 bash -c "</dev/tcp/192.168.0.115/3001" && echo "✅ 端口3001开放" || echo "❌ 端口3001无法连接"
echo ""

echo "3. 测试WebSocket连接 (使用curl)"
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Host: 192.168.0.115:3001" \
  -H "Origin: http://192.168.0.115:3001" \
  -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
  -H "Sec-WebSocket-Version: 13" \
  http://192.168.0.115:3001/ws/bots
echo ""

echo "4. 检查本地IP配置"
ip addr show | grep inet
echo ""

echo "=== 诊断完成 ==="