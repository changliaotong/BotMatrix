import 'package:flutter/material.dart';

class LanguageProvider extends ChangeNotifier {
  Locale _locale = const Locale('zh'); // Default to Chinese as per user preference likely

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
