import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:overmind/l10n/app_localizations.dart';
import '../services/bot_nexus_service.dart';
import '../services/language_provider.dart';
import '../models/bot_info.dart';
import '../models/log_entry.dart';
import 'fleet_screen.dart';
import 'routing_screen.dart';

class DashboardScreen extends StatelessWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    print('Building DashboardScreen...');
    final l10n = AppLocalizations.of(context)!;
    
    return DefaultTabController(
      length: 4,
      child: Scaffold(
        backgroundColor: const Color(0xFF0D1117), // Dark sci-fi background
        appBar: AppBar(
          backgroundColor: const Color(0xFF161B22),
          title: Text(l10n.appTitle, style: const TextStyle(letterSpacing: 2, fontWeight: FontWeight.bold, color: Colors.cyanAccent)),
          bottom: TabBar(
            indicatorColor: Colors.cyanAccent,
            labelColor: Colors.cyanAccent,
            unselectedLabelColor: Colors.grey,
            tabs: [
              Tab(icon: const Icon(Icons.hub), text: l10n.tabNexus),
              Tab(icon: const Icon(Icons.dns), text: l10n.tabFleet),
              const Tab(icon: Icon(Icons.route), text: '路由'), // 添加路由管理标签页
              Tab(icon: const Icon(Icons.terminal), text: l10n.tabLogs),
            ],
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.translate, color: Colors.cyanAccent),
              onPressed: () {
                context.read<LanguageProvider>().toggleLocale();
              },
              tooltip: l10n.switchLanguage,
            ),
            Consumer<BotNexusService>(
              builder: (context, service, _) {
                return Padding(
                  padding: const EdgeInsets.only(right: 16.0),
                  child: Icon(
                    Icons.circle,
                    size: 12,
                    color: service.isConnected ? Colors.greenAccent : Colors.redAccent,
                  ).animate(target: service.isConnected ? 1 : 0).fade(),
                );
              },
            )
          ],
        ),
        body: const TabBarView(
          children: [
            BotListTab(),
            FleetScreen(),
            RoutingScreen(), // 添加路由管理界面
            LogConsoleTab(),
          ],
        ),
      ),
    );
  }
}

class BotListTab extends StatelessWidget {
  const BotListTab({super.key});

  @override
  Widget build(BuildContext context) {
    final service = Provider.of<BotNexusService>(context);
    final l10n = AppLocalizations.of(context)!;
    
    if (service.bots.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.cloud_off, size: 64, color: Colors.grey),
            const SizedBox(height: 16),
            Text(l10n.noActiveNodes, style: const TextStyle(color: Colors.grey)),
            if (!service.isConnected)
              Text(l10n.disconnectedNexus, style: const TextStyle(color: Colors.redAccent)),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: service.bots.length,
      itemBuilder: (context, index) {
        final bot = service.bots[index];
        return Card(
          color: const Color(0xFF21262D),
          shape: RoundedRectangleBorder(
            side: BorderSide(color: Colors.cyanAccent.withOpacity(0.3)),
            borderRadius: BorderRadius.circular(8),
          ),
          margin: const EdgeInsets.only(bottom: 12),
          child: InkWell(
            onTap: () => _showSalaryDialog(context, bot, service, l10n),
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 8.0),
              child: ListTile(
                leading: CircleAvatar(
                  backgroundImage: NetworkImage(bot.avatarUrl),
                  backgroundColor: Colors.black,
                ),
                title: Row(
                  children: [
                    Text(bot.nickname, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.cyanAccent.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(color: Colors.cyanAccent.withOpacity(0.5)),
                      ),
                      child: Text(
                        '${l10n.kpi}: ${bot.kpiScore.toStringAsFixed(1)}',
                        style: const TextStyle(color: Colors.cyanAccent, fontSize: 10),
                      ),
                    ),
                  ],
                ),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('${bot.platform} • ${bot.id}', style: const TextStyle(color: Colors.grey)),
                    const SizedBox(height: 4),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(2),
                      child: LinearProgressIndicator(
                        value: bot.salaryLimit > 0 ? bot.salaryToken / bot.salaryLimit : 0,
                        backgroundColor: Colors.grey.withOpacity(0.1),
                        valueColor: AlwaysStoppedAnimation<Color>(
                          bot.salaryToken > bot.salaryLimit * 0.8 ? Colors.redAccent : Colors.greenAccent,
                        ),
                        minHeight: 4,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      '${l10n.salary}: ${bot.salaryToken} / ${bot.salaryLimit}',
                      style: const TextStyle(color: Colors.grey, fontSize: 10),
                    ),
                  ],
                ),
                trailing: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text('${l10n.msgPrefix}: ${bot.msgCount}', style: const TextStyle(color: Colors.cyanAccent)),
                    Text(bot.uptime, style: const TextStyle(color: Colors.grey, fontSize: 10)),
                  ],
                ),
              ),
            ),
          ),
        ).animate().fadeIn(delay: (index * 100).ms).slideX();
      },
    );
  }

  void _showSalaryDialog(BuildContext context, BotInfo bot, BotNexusService service, AppLocalizations l10n) {
    final tokenController = TextEditingController(text: '0');
    final limitController = TextEditingController(text: bot.salaryLimit.toString());

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF161B22),
        title: Text('${bot.nickname} - ${l10n.salary}管理', style: const TextStyle(color: Colors.cyanAccent)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: tokenController,
              keyboardType: TextInputType.number,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: '重置已用 Token (设为 0)',
                labelStyle: const TextStyle(color: Colors.grey),
                enabledBorder: UnderlineInputBorder(borderSide: BorderSide(color: Colors.cyanAccent.withOpacity(0.5))),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: limitController,
              keyboardType: TextInputType.number,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: '${l10n.limit} (Token)',
                labelStyle: const TextStyle(color: Colors.grey),
                enabledBorder: UnderlineInputBorder(borderSide: BorderSide(color: Colors.cyanAccent.withOpacity(0.5))),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消', style: TextStyle(color: Colors.grey)),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: Colors.cyanAccent),
            onPressed: () async {
              final newToken = int.tryParse(tokenController.text);
              final newLimit = int.tryParse(limitController.text);
              if (newToken != null || newLimit != null) {
                final success = await service.updateEmployeeSalary(
                  bot.id,
                  salaryToken: newToken,
                  salaryLimit: newLimit,
                );
                if (success) {
                  Navigator.pop(context);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('更新成功'), backgroundColor: Colors.green),
                  );
                }
              }
            },
            child: const Text('保存', style: TextStyle(color: Colors.black)),
          ),
        ],
      ),
    );
  }
}

class LogConsoleTab extends StatelessWidget {
  const LogConsoleTab({super.key});

  @override
  Widget build(BuildContext context) {
    final service = Provider.of<BotNexusService>(context);
    final logs = service.logs.reversed.toList(); // Newest first

    return Container(
      color: const Color(0xFF000000),
      child: ListView.builder(
        padding: const EdgeInsets.all(8),
        itemCount: logs.length,
        itemBuilder: (context, index) {
          final log = logs[index];
          Color logColor = Colors.green;
          if (log.level == 'ERROR') logColor = Colors.red;
          if (log.level == 'WARN') logColor = Colors.orange;
          if (log.level == 'DEBUG') logColor = Colors.grey;

          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 2.0),
            child: RichText(
              text: TextSpan(
                style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
                children: [
                  TextSpan(text: '[${log.time}] ', style: const TextStyle(color: Colors.grey)),
                  TextSpan(text: '${log.level} ', style: TextStyle(color: logColor, fontWeight: FontWeight.bold)),
                  if (log.botId != null)
                    TextSpan(text: '<${log.botId}> ', style: const TextStyle(color: Colors.blueAccent)),
                  TextSpan(text: log.message, style: const TextStyle(color: Colors.white70)),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}
