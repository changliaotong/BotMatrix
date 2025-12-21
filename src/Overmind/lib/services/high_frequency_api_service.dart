import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io_web_socket_channel.dart';
import 'package:flutter/foundation.dart';

/// 高频功能API服务 - 针对移动端和小程序优化
class HighFrequencyApiService {
  static const String baseUrl = 'http://bot-manager:5000';
  static const String wsUrl = 'ws://bot-manager:3005';
  
  static final HighFrequencyApiService _instance = HighFrequencyApiService._internal();
  factory HighFrequencyApiService() => _instance;
  HighFrequencyApiService._internal();

  WebSocketChannel? _webSocketChannel;
  Function(Map<String, dynamic>)? _messageHandler;

  /// 获取机器人状态 - 高频使用
  Future<Map<String, dynamic>> getBotStatus() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/bots'),
        headers: {'Content-Type': 'application/json'},
      ).timeout(const Duration(seconds: 3)); // 移动端优化超时时间

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('获取机器人状态失败: ${response.statusCode}');
      }
    } catch (e) {
      debugPrint('获取机器人状态错误: $e');
      // 返回缓存数据或默认值
      return {
        'bots': [],
        'error': e.toString(),
        'cached': true,
      };
    }
  }

  /// 获取系统状态 - 高频使用
  Future<Map<String, dynamic>> getSystemStatus() async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/api/stats'),
        headers: {'Content-Type': 'application/json'},
      ).timeout(const Duration(seconds: 3));

      if (response.statusCode == 200) {
        return json.decode(response.body);
      } else {
        throw Exception('获取系统状态失败: ${response.statusCode}');
      }
    } catch (e) {
      debugPrint('获取系统状态错误: $e');
      return {
        'cpu_usage': 0,
        'memory_usage': 0,
        'error': e.toString(),
        'cached': true,
      };
    }
  }

  /// 快速启停机器人 - 高频操作
  Future<bool> toggleBot(String botId, bool enable) async {
    try {
      final response = await http.post(
        Uri.parse('$baseUrl/api/bot/toggle'),
        headers: {'Content-Type': 'application/json'},
        body: json.encode({
          'bot_id': botId,
          'enable': enable,
        }),
      ).timeout(const Duration(seconds: 5));

      return response.statusCode == 200;
    } catch (e) {
      debugPrint('机器人启停操作错误: $e');
      return false;
    }
  }

  /// 连接WebSocket - 实时数据推送
  void connectWebSocket(Function(Map<String, dynamic>) onMessage) {
    try {
      _messageHandler = onMessage;
      _webSocketChannel = IOWebSocketChannel.connect(wsUrl);
      
      _webSocketChannel!.stream.listen(
        (message) {
          try {
            final data = json.decode(message);
            _messageHandler?.call(data);
          } catch (e) {
            debugPrint('WebSocket消息解析错误: $e');
          }
        },
        onError: (error) {
          debugPrint('WebSocket连接错误: $error');
          // 3秒后重连
          Future.delayed(const Duration(seconds: 3), () {
            if (_messageHandler != null) {
              connectWebSocket(_messageHandler!);
            }
          });
        },
        onDone: () {
          debugPrint('WebSocket连接关闭');
          // 断线重连
          Future.delayed(const Duration(seconds: 3), () {
            if (_messageHandler != null) {
              connectWebSocket(_messageHandler!);
            }
          });
        },
      );
    } catch (e) {
      debugPrint('WebSocket连接失败: $e');
    }
  }

  /// 断开WebSocket连接
  void disconnectWebSocket() {
    _messageHandler = null;
    _webSocketChannel?.sink.close();
    _webSocketChannel = null;
  }

  /// 发送WebSocket消息
  void sendWebSocketMessage(Map<String, dynamic> message) {
    if (_webSocketChannel != null) {
      _webSocketChannel!.sink.add(json.encode(message));
    }
  }
}