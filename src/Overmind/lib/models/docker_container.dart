class DockerContainer {
  final String id;
  final List<String> names;
  final String image;
  final String state;
  final String status;

  DockerContainer({
    required this.id,
    required this.names,
    required this.image,
    required this.state,
    required this.status,
  });

  factory DockerContainer.fromJson(Map<String, dynamic> json) {
    return DockerContainer(
      id: json['Id'] ?? '',
      names: List<String>.from(json['Names'] ?? []),
      image: json['Image'] ?? '',
      state: json['State'] ?? '',
      status: json['Status'] ?? '',
    );
  }

  String get name => names.isNotEmpty ? names.first.replaceAll('/', '') : id.substring(0, 12);
}
