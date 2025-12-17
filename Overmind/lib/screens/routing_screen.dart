import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_animate/flutter_animate.dart';
import '../services/bot_nexus_service.dart';
import '../models/bot_info.dart';
import '../l10n/app_localizations.dart';

class RoutingScreen extends StatefulWidget {
  const RoutingScreen({super.key});

  @override
  State<RoutingScreen> createState() => _RoutingScreenState();
}

class _RoutingScreenState extends State<RoutingScreen> {
  Map<String, String> _routingRules = {};
  List<BotInfo> _bots = [];
  List<Map<String, dynamic>> _workers = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      final service = context.read<BotNexusService>();
      
      // 获取路由规则
      final routingData = await service.getRoutingRules();
      _routingRules = Map<String, String>.from(routingData['rules'] ?? {});
      
      // 获取bot列表
      _bots = service.bots;
      
      // 获取worker列表
      await _fetchWorkers();
      
      setState(() {
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _error = '加载数据失败: $e';
      });
    }
  }

  Future<void> _fetchWorkers() async {
    try {
      final service = context.read<BotNexusService>();
      
      // 调用API获取worker列表
      final response = await service.getWorkers();
      final List<dynamic> workersData = response['workers'] ?? [];
      
      setState(() {
        _workers = workersData.cast<Map<String, dynamic>>();
      });
    } catch (e) {
      print('获取worker列表失败: $e');
      // 使用默认的worker列表作为后备
      setState(() {
        _workers = [
          {'id': 'worker_1', 'handled_count': 0},
          {'id': 'worker_2', 'handled_count': 0},
          {'id': 'worker_3', 'handled_count': 0},
        ];
      });
    }
  }

  Future<void> _setRoutingRule(String key, String workerId) async {
    try {
      final service = context.read<BotNexusService>();
      final success = await service.setRoutingRule(key, workerId);
      
      if (success) {
        // 重新加载数据
        await _loadData();
        
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('路由规则设置成功')),
          );
        }
      } else {
        throw Exception('设置失败');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('设置失败: $e')),
        );
      }
    }
  }

  void _showAddRuleDialog() {
    String? selectedKey;
    String? selectedWorker;
    bool isCustomKey = false;
    final keyController = TextEditingController();
    final workerController = TextEditingController();

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('添加路由规则'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // 选择路由键类型
                Row(
                  children: [
                    Expanded(
                      child: RadioListTile<bool>(
                        title: const Text('选择已有Bot/群'),
                        value: false,
                        groupValue: isCustomKey,
                        onChanged: (value) => setState(() => isCustomKey = value!),
                      ),
                    ),
                    Expanded(
                      child: RadioListTile<bool>(
                        title: const Text('自定义键'),
                        value: true,
                        groupValue: isCustomKey,
                        onChanged: (value) => setState(() => isCustomKey = value!),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                
                // 路由键选择
                if (!isCustomKey) ...[
                  DropdownButtonFormField<String>(
                    value: selectedKey,
                    decoration: const InputDecoration(
                      labelText: '选择Bot/群',
                      border: OutlineInputBorder(),
                    ),
                    items: [
                      // Bot列表
                      if (_bots.isNotEmpty) ...[
                        const DropdownMenuItem(
                          value: null,
                          enabled: false,
                          child: Text('已连接的机器人', style: TextStyle(fontWeight: FontWeight.bold)),
                        ),
                        ..._bots.map((bot) => DropdownMenuItem(
                          value: bot.id,
                          child: Text('🤖 ${bot.id} (${bot.platform})'),
                        )),
                      ],
                      // 群列表（如果有的话）
                      const DropdownMenuItem(
                        value: null,
                        enabled: false,
                        child: Text('已知的群', style: TextStyle(fontWeight: FontWeight.bold)),
                      ),
                      // 这里可以添加群列表
                    ],
                    onChanged: (value) => selectedKey = value,
                    validator: (value) => value == null ? '请选择一个键' : null,
                  ),
                ] else ...[
                  TextFormField(
                    controller: keyController,
                    decoration: const InputDecoration(
                      labelText: '自定义键 (group_id/bot_id)',
                      border: OutlineInputBorder(),
                      hintText: '例如: 123456 或 bot_123',
                    ),
                    validator: (value) => value?.isEmpty ?? true ? '请输入键' : null,
                  ),
                ],
                
                const SizedBox(height: 16),
                
                // Worker选择
                DropdownButtonFormField<String>(
                  value: selectedWorker,
                  decoration: const InputDecoration(
                    labelText: '选择目标Worker',
                    border: OutlineInputBorder(),
                  ),
                  items: [
                    const DropdownMenuItem(
                      value: null,
                      enabled: false,
                      child: Text('可用的Workers', style: TextStyle(fontWeight: FontWeight.bold)),
                    ),
                    ..._workers.map((worker) => DropdownMenuItem(
                      value: worker['id'] as String,
                      child: Text('⚙️ ${worker['id']} (处理: ${worker['handled_count']})'),
                    )),
                  ],
                  onChanged: (value) => selectedWorker = value,
                  validator: (value) => value == null ? '请选择一个worker' : null,
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('取消'),
            ),
            ElevatedButton(
              onPressed: () {
                final key = isCustomKey ? keyController.text : selectedKey;
                if (key != null && key.isNotEmpty && selectedWorker != null) {
                  Navigator.pop(context);
                  _setRoutingRule(key, selectedWorker!);
                }
              },
              child: const Text('确定'),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    
    return Scaffold(
      backgroundColor: const Color(0xFF0D1117),
      appBar: AppBar(
        backgroundColor: const Color(0xFF161B22),
        title: const Text('路由规则管理', style: TextStyle(color: Colors.cyanAccent)),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh, color: Colors.cyanAccent),
            onPressed: _loadData,
            tooltip: '刷新',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Colors.cyanAccent))
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_error!, style: const TextStyle(color: Colors.red)),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadData,
                        child: const Text('重试'),
                      ),
                    ],
                  ),
                )
              : Column(
                  children: [
                    // 统计信息
                    Container(
                      margin: const EdgeInsets.all(16),
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: const Color(0xFF161B22),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.cyanAccent.withOpacity(0.3)),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Column(
                              children: [
                                Text(
                                  '${_routingRules.length}',
                                  style: const TextStyle(
                                    fontSize: 24,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.cyanAccent,
                                  ),
                                ),
                                const Text('路由规则', style: TextStyle(color: Colors.grey)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Column(
                              children: [
                                Text(
                                  '${_bots.length}',
                                  style: const TextStyle(
                                    fontSize: 24,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.greenAccent,
                                  ),
                                ),
                                const Text('已连接机器人', style: TextStyle(color: Colors.grey)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Column(
                              children: [
                                Text(
                                  '${_workers.length}',
                                  style: const TextStyle(
                                    fontSize: 24,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.orangeAccent,
                                  ),
                                ),
                                const Text('可用Workers', style: TextStyle(color: Colors.grey)),
                                Text(
                                  '总处理: ${_workers.fold<int>(0, (sum, w) => sum + (w['handled_count'] as int? ?? 0))}',
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: Colors.grey.withOpacity(0.7),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    
                    // 路由规则列表
                    Expanded(
                      child: _routingRules.isEmpty
                          ? Center(
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(
                                    Icons.route,
                                    size: 64,
                                    color: Colors.grey.withOpacity(0.5),
                                  ),
                                  const SizedBox(height: 16),
                                  Text(
                                    '暂无路由规则',
                                    style: TextStyle(
                                      fontSize: 18,
                                      color: Colors.grey.withOpacity(0.7),
                                    ),
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    '点击右下角的 + 按钮添加规则',
                                    style: TextStyle(
                                      color: Colors.grey.withOpacity(0.5),
                                    ),
                                  ),
                                ],
                              ),
                            )
                          : ListView.builder(
                              padding: const EdgeInsets.symmetric(horizontal: 16),
                              itemCount: _routingRules.length,
                              itemBuilder: (context, index) {
                                final entry = _routingRules.entries.elementAt(index);
                                final key = entry.key;
                                final workerId = entry.value;
                                
                                return Card(
                                  color: const Color(0xFF161B22),
                                  elevation: 2,
                                  margin: const EdgeInsets.only(bottom: 8),
                                  child: ListTile(
                                    leading: Icon(
                                      key.startsWith('bot_') ? Icons.smart_toy : Icons.group,
                                      color: Colors.cyanAccent,
                                    ),
                                    title: Text(
                                      key,
                                      style: const TextStyle(
                                        color: Colors.white,
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    subtitle: Text(
                                      '→ $workerId',
                                      style: TextStyle(
                                        color: Colors.grey.withOpacity(0.8),
                                      ),
                                    ),
                                    trailing: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        IconButton(
                                          icon: const Icon(Icons.edit, color: Colors.blue),
                                          onPressed: () => _showEditRuleDialog(key, workerId),
                                        ),
                                        IconButton(
                                          icon: const Icon(Icons.delete, color: Colors.red),
                                          onPressed: () => _confirmDeleteRule(key),
                                        ),
                                      ],
                                    ),
                                  ),
                                ).animate()
                                  .fadeIn(duration: 300.ms)
                                  .slideX(begin: -0.2, duration: 300.ms);
                              },
                            ),
                    ),
                  ],
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showAddRuleDialog,
        backgroundColor: Colors.cyanAccent,
        child: const Icon(Icons.add),
      ),
    );
  }

  void _showEditRuleDialog(String key, String currentWorkerId) {
    String? selectedWorker = currentWorkerId;

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('编辑规则: $key'),
        content: DropdownButtonFormField<String>(
          value: selectedWorker,
          decoration: const InputDecoration(
            labelText: '选择目标Worker',
            border: OutlineInputBorder(),
          ),
          items: [
            const DropdownMenuItem(
              value: null,
              enabled: false,
              child: Text('可用的Workers', style: TextStyle(fontWeight: FontWeight.bold)),
            ),
            ..._workers.map((worker) => DropdownMenuItem(
              value: worker['id'] as String,
              child: Text('⚙️ ${worker['id']}'),
            )),
          ],
          onChanged: (value) => selectedWorker = value,
          validator: (value) => value == null ? '请选择一个worker' : null,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              if (selectedWorker != null) {
                Navigator.pop(context);
                _setRoutingRule(key, selectedWorker!);
              }
            },
            child: const Text('保存'),
          ),
        ],
      ),
    );
  }

  void _confirmDeleteRule(String key) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('确认删除'),
        content: Text('确定要删除路由规则 "$key" 吗？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _setRoutingRule(key, ''); // 空worker_id表示删除
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('删除'),
          ),
        ],
      ),
    );
  }
}