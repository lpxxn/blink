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
}
