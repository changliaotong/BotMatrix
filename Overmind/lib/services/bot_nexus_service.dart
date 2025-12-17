import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:http/http.dart' as http;
import '../models/log_entry.dart';
import '../models/bot_info.dart';
import '../models/docker_container.dart';

class BotNexusService extends ChangeNotifier {
  WebSocketChannel? _channel;
  bool _isConnected = false;
  List<LogEntry> _logs = [];
  List<BotInfo> _bots = [];
  List<DockerContainer> _containers = [];

  bool get isConnected => _isConnected;
  List<LogEntry> get logs => _logs;
  List<BotInfo> get bots => _bots;
  List<DockerContainer> get containers => _containers;

  // Configuration
  // For Web: Use current window location
  // For Android Emulator: 10.0.2.2
  // For Windows/Real Device: localhost or LAN IP
  String get _wsUrl {
    if (kIsWeb) {
      // In development, assume backend is on localhost:3001
      if (kDebugMode) return 'ws://localhost:3001/?role=subscriber';

      // In production (served by BotNexus), use the same origin
      final scheme = Uri.base.scheme == 'https' ? 'wss' : 'ws';
      final host = Uri.base.host;
      final port = Uri.base.port;
      final effectiveHost = host.isEmpty ? 'localhost' : host;
      return '$scheme://$effectiveHost:$port/ws?role=subscriber';
    }
    if (Platform.isAndroid) return 'ws://10.0.2.2:3001/?role=subscriber';
    return 'ws://localhost:3001/?role=subscriber';
  }

  String get _apiBaseUrl {
    if (kIsWeb) {
      if (kDebugMode) return 'http://localhost:5000/api';
      
      final origin = Uri.base.origin;
      return '$origin/api';
    }
    if (Platform.isAndroid) return 'http://10.0.2.2:5000/api';
    return 'http://localhost:5000/api';
  }

  // Auth Token (Mock for now or retrieved from login)
  String _token = ""; 
  
  void setToken(String token) {
    _token = token;
  }

  void connect() {
    try {
      if (_isConnected) return;
      
      print('Connecting to $_wsUrl');
      _channel = WebSocketChannel.connect(Uri.parse(_wsUrl));
      _isConnected = true;
      notifyListeners();

      _channel!.stream.listen(
        (message) {
          _handleMessage(message);
        },
        onDone: () {
          print('WebSocket Closed');
          _isConnected = false;
          notifyListeners();
          _reconnect();
        },
        onError: (error) {
          print('WebSocket Error: $error');
          _isConnected = false;
          notifyListeners();
          _reconnect();
        },
      );
    } catch (e) {
      print('Connection failed: $e');
      _reconnect();
    }
  }

  void _reconnect() async {
    await Future.delayed(const Duration(seconds: 5));
    connect();
  }

  void _handleMessage(dynamic message) {
    try {
      final data = jsonDecode(message);
      final type = data['type'];
      final payload = data['data'];

      switch (type) {
        case 'initial_logs':
          if (payload is List) {
            _logs = payload.map((e) => LogEntry.fromJson(e)).toList();
            // Sort by time desc if needed, but usually logs are appended
            // We want newest at bottom usually, but for mobile maybe top?
            // Let's keep original order (oldest first usually)
            notifyListeners();
          }
          break;
        
        case 'log':
          if (payload is Map<String, dynamic>) {
            _logs.add(LogEntry.fromJson(payload));
            if (_logs.length > 1000) {
              _logs.removeAt(0); // Keep buffer size manageable
            }
            notifyListeners();
          }
          break;

        case 'bot_list':
          if (payload is List) {
            _bots = payload.map((e) => BotInfo.fromJson(e)).toList();
            notifyListeners();
          }
          break;
          
        case 'initial_stats':
          // Handle stats if needed
          break;
      }
    } catch (e) {
      print('Error parsing message: $e');
    }
  }

  Future<bool> login(String username, String password) async {
    try {
      final response = await http.post(
        Uri.parse('$_apiBaseUrl/login'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'username': username, 'password': password}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        _token = data['token'];
        notifyListeners();
        return true;
      }
      return false;
    } catch (e) {
      print('Login error: $e');
      return false;
    }
  }

  Future<void> fetchContainers() async {
    try {
      // Auto-login if needed (dev convenience)
      if (_token.isEmpty) {
        await login('admin', 'admin888');
      }

      final response = await http.get(
        Uri.parse('$_apiBaseUrl/docker/list'),
        headers: _token.isNotEmpty ? {'Authorization': 'Bearer $_token'} : {},
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = jsonDecode(response.body);
        _containers = data.map((e) => DockerContainer.fromJson(e)).toList();
        notifyListeners();
      } else {
        print('Failed to fetch containers: ${response.statusCode}');
      }
    } catch (e) {
      print('Error fetching containers: $e');
    }
  }

  Future<bool> controlContainer(String id, String action) async {
    try {
      final response = await http.post(
        Uri.parse('$_apiBaseUrl/docker/action'),
        headers: {
          'Content-Type': 'application/json',
          if (_token.isNotEmpty) 'Authorization': 'Bearer $_token',
        },
        body: jsonEncode({
          'container_id': id,
          'action': action,
        }),
      );

      if (response.statusCode == 200) {
        return true;
      }
      return false;
    } catch (e) {
      print('Error controlling container: $e');
      return false;
    }
  }

  // 路由规则管理相关方法
  Future<Map<String, dynamic>> getRoutingRules() async {
    try {
      final response = await http.get(
        Uri.parse('$_apiBaseUrl/admin/routing'),
        headers: _token.isNotEmpty ? {'Authorization': 'Bearer $_token'} : {},
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      }
      throw Exception('Failed to get routing rules: ${response.statusCode}');
    } catch (e) {
      print('Error getting routing rules: $e');
      rethrow;
    }
  }

  Future<bool> setRoutingRule(String key, String workerId) async {
    try {
      final response = await http.post(
        Uri.parse('$_apiBaseUrl/admin/routing'),
        headers: {
          'Content-Type': 'application/json',
          if (_token.isNotEmpty) 'Authorization': 'Bearer $_token',
        },
        body: jsonEncode({
          'key': key,
          'worker_id': workerId,
        }),
      );

      if (response.statusCode == 200) {
        return true;
      }
      return false;
    } catch (e) {
      print('Error setting routing rule: $e');
      return false;
    }
  }

  Future<Map<String, dynamic>> getWorkers() async {
    try {
      final response = await http.get(
        Uri.parse('$_apiBaseUrl/workers/list'),
        headers: _token.isNotEmpty ? {'Authorization': 'Bearer $_token'} : {},
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      }
      throw Exception('Failed to get workers: ${response.statusCode}');
    } catch (e) {
      print('Error getting workers: $e');
      rethrow;
    }
  }
    } catch (e) {
      print('Error controlling container: $e');
      return false;
    }
  }

  @override
  void dispose() {
    _channel?.sink.close();
    super.dispose();
  }
}
