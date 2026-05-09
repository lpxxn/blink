/// GET `/api/me` — `user_id` is JSON string (snowflake).
final class Me {
  const Me({
    required this.userId,
    required this.email,
    required this.name,
    this.role,
  });

  final String userId;
  final String email;
  final String name;
  final String? role;

  factory Me.fromJson(Map<String, dynamic> json) {
    final id = json['user_id'];
    return Me(
      userId: id is String ? id : id.toString(),
      email: json['email'] as String? ?? '',
      name: json['name'] as String? ?? '',
      role: json['role'] as String?,
    );
  }
}
