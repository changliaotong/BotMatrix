import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

/// 移动端核心功能管理器
class MobileCore {
  static final MobileCore _instance = MobileCore._internal();
  factory MobileCore() => _instance;
  MobileCore._internal();

  /// 检测运行平台
  static String get platform {
    if (kIsWeb) return 'web';
    if (defaultTargetPlatform == TargetPlatform.android) return 'android';
    if (defaultTargetPlatform == TargetPlatform.iOS) return 'ios';
    if (defaultTargetPlatform == TargetPlatform.windows) return 'windows';
    if (defaultTargetPlatform == TargetPlatform.macOS) return 'macos';
    if (defaultTargetPlatform == TargetPlatform.linux) return 'linux';
    return 'unknown';
  }

  /// 检测是否为小程序环境
  static bool get isMiniProgram {
    // 检测微信/QQ小程序环境
    if (kIsWeb) {
      // 通过UserAgent检测
      return false; // 暂时返回false，后续实现具体检测逻辑
    }
    return false;
  }

  /// 获取设备信息
  static Future<Map<String, dynamic>> getDeviceInfo() async {
    try {
      final deviceInfo = <String, dynamic>{};
      
      // 平台信息
      deviceInfo['platform'] = platform;
      deviceInfo['isMiniProgram'] = isMiniProgram;
      
      // 如果是移动设备，获取更多信息
      if (platform == 'android' || platform == 'ios') {
        // 这里可以集成device_info_plus包来获取详细信息
        deviceInfo['deviceType'] = 'mobile';
      }
      
      return deviceInfo;
    } catch (e) {
      debugPrint('获取设备信息失败: $e');
      return {'platform': platform, 'error': e.toString()};
    }
  }

  /// 适配不同平台的UI尺寸
  static double adaptSize(double size) {
    // 根据平台调整UI尺寸
    if (isMiniProgram) {
      return size * 0.9; // 小程序稍微缩小
    }
    return size;
  }
}