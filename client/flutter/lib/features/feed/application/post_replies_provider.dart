import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/feed_models.dart';
import '../data/social_repository_provider.dart';

part 'post_replies_provider.g.dart';

@riverpod
class PostReplies extends _$PostReplies {
  String? _nextCursor;
  bool _exhausted = false;
  String _postId = '';

  @override
  Future<List<FeedReply>> build(String postId) async {
    _postId = postId;
    _nextCursor = null;
    _exhausted = false;
    final RepliesPage page =
        await ref.read(socialRepositoryProvider).fetchReplies(
              postId,
              cursor: null,
              limit: 30,
            );
    _nextCursor = page.nextCursor;
    _exhausted =
        page.nextCursor == null || page.replies.isEmpty;
    return page.replies;
  }

  Future<void> loadMore() async {
    final List<FeedReply>? current = state.valueOrNull;
    if (current == null || _exhausted || _nextCursor == null) {
      return;
    }
    final String postId = _postId;
    final RepliesPage page =
        await ref.read(socialRepositoryProvider).fetchReplies(
              postId,
              cursor: _nextCursor,
              limit: 30,
            );
    _nextCursor = page.nextCursor;
    if (page.nextCursor == null || page.replies.isEmpty) {
      _exhausted = true;
    }
    state = AsyncData(<FeedReply>[...current, ...page.replies]);
  }
}
