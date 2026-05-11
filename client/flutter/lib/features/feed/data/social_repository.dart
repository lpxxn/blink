import 'package:dio/dio.dart';

import 'feed_models.dart';

class SocialRepository {
  SocialRepository(this._dio);

  final Dio _dio;

  Future<List<FeedCategory>> fetchCategories() async {
    final Response<Map<String, dynamic>> res =
        await _dio.get<Map<String, dynamic>>('/api/categories');
    final data = res.data;
    if (data == null) {
      throw StateError('Empty categories response');
    }
    final raw = data['categories'];
    if (raw is! List<dynamic>) {
      return const [];
    }
    return raw
        .map((e) => FeedCategory.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<PostsPage> fetchPosts({
    String? categoryId,
    bool uncategorized = false,
    String? cursor,
    int limit = 20,
  }) async {
    final Map<String, dynamic> q = <String, dynamic>{'limit': limit};
    if (categoryId != null && categoryId.isNotEmpty) {
      q['category_id'] = categoryId;
    }
    if (uncategorized) {
      q['uncategorized'] = '1';
    }
    if (cursor != null && cursor.isNotEmpty) {
      q['cursor'] = cursor;
    }
    final Response<Map<String, dynamic>> res = await _dio.get<Map<String, dynamic>>(
      '/api/posts',
      queryParameters: q,
    );
    final data = res.data;
    if (data == null) {
      throw StateError('Empty posts response');
    }
    final raw = data['posts'];
    final List<FeedPost> posts = raw is List<dynamic>
        ? raw
            .map((e) => FeedPost.fromJson(e as Map<String, dynamic>))
            .toList()
        : const [];
    final dynamic nc = data['next_cursor'];
    final String? nextCursor =
        nc == null ? null : (nc is String ? nc : nc.toString());
    return PostsPage(posts: posts, nextCursor: nextCursor);
  }

  /// GET `/api/posts/:id` — same shape as feed post.
  Future<FeedPost> fetchPost(String id) async {
    final Response<Map<String, dynamic>> res =
        await _dio.get<Map<String, dynamic>>('/api/posts/$id');
    final data = res.data;
    if (data == null) {
      throw StateError('Empty post response');
    }
    return FeedPost.fromJson(data);
  }

  /// GET `/api/posts/:id/replies` — `cursor` is last reply id for pagination.
  Future<RepliesPage> fetchReplies(
    String postId, {
    String? cursor,
    int limit = 50,
  }) async {
    final Map<String, dynamic> q = <String, dynamic>{'limit': limit};
    if (cursor != null && cursor.isNotEmpty) {
      q['cursor'] = cursor;
    }
    final Response<Map<String, dynamic>> res =
        await _dio.get<Map<String, dynamic>>(
      '/api/posts/$postId/replies',
      queryParameters: q,
    );
    final data = res.data;
    if (data == null) {
      throw StateError('Empty replies response');
    }
    final raw = data['replies'];
    final List<FeedReply> replies = raw is List<dynamic>
        ? raw
            .map((e) => FeedReply.fromJson(e as Map<String, dynamic>))
            .toList()
        : const [];
    final dynamic nc = data['next_cursor'];
    final String? nextCursor =
        nc == null ? null : (nc is String ? nc : nc.toString());
    return RepliesPage(replies: replies, nextCursor: nextCursor);
  }

  /// POST `/api/posts/:id/replies` — requires session. Returns created reply JSON.
  Future<FeedReply> createReply({
    required String postId,
    required String body,
    String? parentReplyId,
  }) async {
    final Map<String, dynamic> data = <String, dynamic>{'body': body};
    if (parentReplyId != null && parentReplyId.isNotEmpty) {
      data['parent_reply_id'] = parentReplyId;
    }
    final Response<Map<String, dynamic>> res =
        await _dio.post<Map<String, dynamic>>(
      '/api/posts/$postId/replies',
      data: data,
    );
    final Map<String, dynamic>? json = res.data;
    if (json == null) {
      throw StateError('Empty create reply response');
    }
    return FeedReply.fromJson(json);
  }
}
