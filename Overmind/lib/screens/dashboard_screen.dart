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
    final l10n = AppLocalizations.of(context)!;
    
    return DefaultTabController(
      length: 3,
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
              Tab(icon: const Icon(Icons.route), text: '路由'), // 添加路由管理标签页
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
        body: TabBarView(
          children: [
            const BotListTab(),
            const FleetScreen(),
            const RoutingScreen(), // 添加路由管理界面
            const LogConsoleTab(),
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
          child: ListTile(
            leading: CircleAvatar(
              backgroundImage: NetworkImage(bot.avatarUrl),
              backgroundColor: Colors.black,
            ),
            title: Text(bot.nickname, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
            subtitle: Text('${bot.platform} • ${bot.id}', style: const TextStyle(color: Colors.grey)),
            trailing: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text('${l10n.msgPrefix}: ${bot.msgCount}', style: const TextStyle(color: Colors.cyanAccent)),
                Text(bot.uptime, style: const TextStyle(color: Colors.grey, fontSize: 10)),
              ],
            ),
          ),
        ).animate().fadeIn(delay: (index * 100).ms).slideX();
      },
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
