import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'dart:html' as html;

class LanguageProvider extends ChangeNotifier {
  Locale _locale = const Locale('zh');

  LanguageProvider() {
    _initFromUrl();
  }

  void _initFromUrl() {
    if (kIsWeb) {
      final uri = Uri.parse(html.window.location.href);
      final lang = uri.queryParameters['lang'];
      if (lang != null) {
        if (lang.startsWith('zh')) {
          _locale = const Locale('zh');
        } else if (lang.startsWith('en')) {
          _locale = const Locale('en');
        }
      }
    }
  }

  Locale get locale => _locale;

  void setLocale(Locale locale) {
    if (!['en', 'zh'].contains(locale.languageCode)) return;
    _locale = locale;
    notifyListeners();
  }

  void toggleLocale() {
    if (_locale.languageCode == 'en') {
      _locale = const Locale('zh');
    } else {
      _locale = const Locale('en');
    }
    notifyListeners();
  }
}
