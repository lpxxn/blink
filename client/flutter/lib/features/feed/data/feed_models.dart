/// GET `/api/categories` item.
final class FeedCategory {
  const FeedCategory({
    required this.id,
    required this.slug,
    required this.name,
    required this.sortOrder,
  });

  final String id;
  final String slug;
  final String name;
  final int sortOrder;

  factory FeedCategory.fromJson(Map<String, dynamic> json) {
    return FeedCategory(
      id: _idToString(json['id']),
      slug: json['slug'] as String? ?? '',
      name: json['name'] as String? ?? '',
      sortOrder: json['sort_order'] as int? ?? 0,
    );
  }
}

/// GET `/api/posts` item (public feed).
final class FeedPost {
  const FeedPost({
    required this.id,
    required this.userId,
    required this.userName,
    required this.body,
    required this.images,
    required this.createdAt,
    this.categoryId,
  });

  final String id;
  final String userId;
  final String userName;
  final String body;
  final List<String> images;
  final String createdAt;
  final String? categoryId;

  factory FeedPost.fromJson(Map<String, dynamic> json) {
    return FeedPost(
      id: _idToString(json['id']),
      userId: _idToString(json['user_id']),
      userName: json['user_name'] as String? ?? '',
      body: json['body'] as String? ?? '',
      images: (json['images'] as List<dynamic>?)
              ?.map((e) => e as String)
              .toList() ??
          const [],
      createdAt: json['created_at'] as String? ?? '',
      categoryId: _optionalId(json['category_id']),
    );
  }
}

final class PostsPage {
  const PostsPage({required this.posts, this.nextCursor});

  final List<FeedPost> posts;
  final String? nextCursor;
}

/// GET `/api/posts/:id/replies` item.
final class FeedReply {
  const FeedReply({
    required this.id,
    required this.postId,
    required this.userId,
    required this.userName,
    required this.body,
    required this.createdAt,
    this.parentReplyId,
  });

  final String id;
  final String postId;
  final String userId;
  final String userName;
  final String body;
  final String createdAt;
  final String? parentReplyId;

  factory FeedReply.fromJson(Map<String, dynamic> json) {
    return FeedReply(
      id: _idToString(json['id']),
      postId: _idToString(json['post_id']),
      userId: _idToString(json['user_id']),
      userName: json['user_name'] as String? ?? '',
      body: json['body'] as String? ?? '',
      createdAt: json['created_at'] as String? ?? '',
      parentReplyId: _optionalId(json['parent_reply_id']),
    );
  }
}

final class RepliesPage {
  const RepliesPage({required this.replies, this.nextCursor});

  final List<FeedReply> replies;
  final String? nextCursor;
}

String _idToString(Object? value) {
  if (value == null) {
    return '';
  }
  if (value is String) {
    return value;
  }
  return value.toString();
}

String? _optionalId(Object? value) {
  if (value == null) {
    return null;
  }
  final s = _idToString(value);
  return s.isEmpty ? null : s;
}
