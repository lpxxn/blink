import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/feed_models.dart';
import '../data/social_repository_provider.dart';
import 'public_feed_query_provider.dart';

part 'public_feed_provider.g.dart';

@riverpod
class PublicFeed extends _$PublicFeed {
  String? _nextCursor;
  bool _exhausted = false;

  @override
  Future<List<FeedPost>> build() async {
    final PublicFeedFilter query = ref.watch(publicFeedQueryProvider);
    _nextCursor = null;
    _exhausted = false;

    final PostsPage page = await ref.read(socialRepositoryProvider).fetchPosts(
          categoryId: query.categoryId,
          uncategorized: query.uncategorized,
          cursor: null,
          limit: 20,
        );
    _nextCursor = page.nextCursor;
    _exhausted =
        page.nextCursor == null || page.posts.isEmpty;
    return page.posts;
  }

  Future<void> loadMore() async {
    final List<FeedPost>? current = state.valueOrNull;
    if (current == null || _exhausted || _nextCursor == null) {
      return;
    }

    final PublicFeedFilter query = ref.read(publicFeedQueryProvider);
    final PostsPage page = await ref.read(socialRepositoryProvider).fetchPosts(
          categoryId: query.categoryId,
          uncategorized: query.uncategorized,
          cursor: _nextCursor,
          limit: 20,
        );
    _nextCursor = page.nextCursor;
    if (page.nextCursor == null || page.posts.isEmpty) {
      _exhausted = true;
    }
    state = AsyncData(<FeedPost>[...current, ...page.posts]);
  }
}
