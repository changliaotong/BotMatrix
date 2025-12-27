import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:overmind/l10n/app_localizations.dart';
import '../services/bot_nexus_service.dart';
import '../models/docker_container.dart';

class FleetScreen extends StatefulWidget {
  const FleetScreen({super.key});

  @override
  State<FleetScreen> createState() => _FleetScreenState();
}

class _FleetScreenState extends State<FleetScreen> {
  @override
  void initState() {
    super.initState();
    // Fetch on init
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<BotNexusService>().fetchContainers();
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    
    return Consumer<BotNexusService>(
      builder: (context, service, child) {
        return Scaffold(
          backgroundColor: Colors.transparent,
          body: service.containers.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.dns_outlined, size: 64, color: Colors.grey),
                      const SizedBox(height: 16),
                      Text(
                        l10n.noActiveUnits,
                        style: TextStyle(color: Colors.grey[400]),
                      ),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: () => service.fetchContainers(),
                        child: Text(l10n.scanNetwork),
                      ),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: service.containers.length,
                  itemBuilder: (context, index) {
                    final container = service.containers[index];
                    return _buildContainerCard(context, container, service);
                  },
                ),
          floatingActionButton: FloatingActionButton(
            backgroundColor: Colors.cyanAccent,
            onPressed: () => service.fetchContainers(),
            child: const Icon(Icons.refresh, color: Colors.black),
          ),
        );
      },
    );
  }

  Widget _buildContainerCard(BuildContext context, DockerContainer container, BotNexusService service) {
    final isRunning = container.state == 'running';
    final statusColor = isRunning ? Colors.greenAccent : Colors.redAccent;
    final l10n = AppLocalizations.of(context)!;

    return Card(
      color: const Color(0xFF1E2329),
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: statusColor.withOpacity(0.3), width: 1),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.layers, color: statusColor),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        container.name,
                        style: const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                      Text(
                        container.image,
                        style: TextStyle(color: Colors.grey[400], fontSize: 12),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: statusColor.withOpacity(0.5)),
                  ),
                  child: Text(
                    container.state.toUpperCase(),
                    style: TextStyle(color: statusColor, fontSize: 10, fontWeight: FontWeight.bold),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                if (isRunning) ...[
                  _ActionButton(
                    icon: Icons.restart_alt,
                    label: l10n.restart,
                    color: Colors.orangeAccent,
                    onTap: () => service.controlContainer(container.id, 'restart'),
                  ),
                  const SizedBox(width: 12),
                  _ActionButton(
                    icon: Icons.stop,
                    label: l10n.stop,
                    color: Colors.redAccent,
                    onTap: () => service.controlContainer(container.id, 'stop'),
                  ),
                ] else ...[
                  _ActionButton(
                    icon: Icons.play_arrow,
                    label: l10n.start,
                    color: Colors.greenAccent,
                    onTap: () => service.controlContainer(container.id, 'start'),
                  ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _ActionButton({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(4),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          border: Border.all(color: color.withOpacity(0.5)),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Row(
          children: [
            Icon(icon, size: 16, color: color),
            const SizedBox(width: 6),
            Text(
              label,
              style: TextStyle(color: color, fontWeight: FontWeight.bold, fontSize: 12),
            ),
          ],
        ),
      ),
    );
  }
}
