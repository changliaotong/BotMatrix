import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:overmind/l10n/app_localizations.dart';
import 'services/bot_nexus_service.dart';
import 'services/language_provider.dart';
import 'screens/dashboard_screen.dart';

void main() {
  print('App starting...');
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) {
          print('Initializing BotNexusService...');
          return BotNexusService()..connect();
        }),
        ChangeNotifierProvider(create: (_) {
          print('Initializing LanguageProvider...');
          return LanguageProvider();
        }),
      ],
      child: const OvermindApp(),
    ),
  );
}

class OvermindApp extends StatelessWidget {
  const OvermindApp({super.key});

  @override
  Widget build(BuildContext context) {
    print('Building OvermindApp...');
    final languageProvider = Provider.of<LanguageProvider>(context);

    return MaterialApp(
      title: 'Overmind',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        primarySwatch: Colors.cyan,
        useMaterial3: true,
      ),
      locale: languageProvider.locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('en'), // English
        Locale('zh'), // Chinese
      ],
      home: const DashboardScreen(),
    );
  }
}
