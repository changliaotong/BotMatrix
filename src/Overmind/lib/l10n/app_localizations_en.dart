// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'OVERMIND';

  @override
  String get tabNexus => 'NEXUS';

  @override
  String get tabFleet => 'FLEET';

  @override
  String get tabLogs => 'LOGS';

  @override
  String get noActiveNodes => 'No Active Nodes';

  @override
  String get disconnectedNexus => 'Disconnected from Nexus';

  @override
  String get msgPrefix => 'MSG';

  @override
  String get noActiveUnits => 'No active units detected';

  @override
  String get scanNetwork => 'SCAN NETWORK';

  @override
  String get restart => 'RESTART';

  @override
  String get stop => 'STOP';

  @override
  String get start => 'START';

  @override
  String get settings => 'Settings';

  @override
  String get language => 'Language';

  @override
  String get switchLanguage => 'Switch Language';
}
