class BotInfo {
  final String id;
  final String nickname;
  final String platform;
  final String connectedTime;
  final String uptime;
  final int groupCount;
  final int friendCount;
  final int msgCount;
  final int sentCount;
  final int recvCount;
  final String avatarUrl;
  final double kpiScore;
  final int salaryToken;
  final int salaryLimit;

  BotInfo({
    required this.id,
    required this.nickname,
    required this.platform,
    required this.connectedTime,
    required this.uptime,
    required this.groupCount,
    required this.friendCount,
    required this.msgCount,
    required this.sentCount,
    required this.recvCount,
    required this.avatarUrl,
    this.kpiScore = 0.0,
    this.salaryToken = 0,
    this.salaryLimit = 0,
  });

  factory BotInfo.fromJson(Map<String, dynamic> json) {
    return BotInfo(
      id: json['id'] ?? '',
      nickname: json['nickname'] ?? 'Unknown',
      platform: json['platform'] ?? 'Unknown',
      connectedTime: json['connected'] ?? '',
      uptime: json['uptime'] ?? '',
      groupCount: json['group_count'] ?? 0,
      friendCount: json['friend_count'] ?? 0,
      msgCount: json['msg_count'] ?? 0,
      sentCount: json['sent_count'] ?? 0,
      recvCount: json['recv_count'] ?? 0,
      avatarUrl: json['avatar_url'] ?? '',
      kpiScore: (json['kpi_score'] ?? 0.0).toDouble(),
      salaryToken: json['salary_token'] ?? 0,
      salaryLimit: json['salary_limit'] ?? 0,
    );
  }
}
