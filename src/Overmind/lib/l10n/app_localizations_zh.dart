// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appTitle => '主宰';

  @override
  String get tabNexus => '节点';

  @override
  String get tabFleet => '舰队';

  @override
  String get tabLogs => '日志';

  @override
  String get noActiveNodes => '无活跃节点';

  @override
  String get disconnectedNexus => '已断开连接';

  @override
  String get msgPrefix => '消息';

  @override
  String get noActiveUnits => '未检测到活跃单元';

  @override
  String get scanNetwork => '扫描网络';

  @override
  String get restart => '重启';

  @override
  String get stop => '停止';

  @override
  String get start => '启动';

  @override
  String get settings => '设置';

  @override
  String get language => '语言';

  @override
  String get switchLanguage => '切换语言';
}
