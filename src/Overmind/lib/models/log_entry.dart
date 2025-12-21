class LogEntry {
  final String time;
  final String level;
  final String message;
  final String? botId;

  LogEntry({
    required this.time,
    required this.level,
    required this.message,
    this.botId,
  });

  factory LogEntry.fromJson(Map<String, dynamic> json) {
    return LogEntry(
      time: json['time'] ?? '',
      level: json['level'] ?? 'INFO',
      message: json['message'] ?? '',
      botId: json['bot_id'],
    );
  }
}
