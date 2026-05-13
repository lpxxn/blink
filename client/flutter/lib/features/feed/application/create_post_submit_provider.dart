import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/feed_models.dart';
import '../data/social_repository_provider.dart';

part 'create_post_submit_provider.g.dart';

@riverpod
class CreatePostSubmit extends _$CreatePostSubmit {
  @override
  AsyncValue<void> build() => const AsyncData(null);

  Future<FeedPost> submit({
    required String body,
    required String categoryId,
    bool draft = false,
  }) async {
    state = const AsyncLoading();
    try {
      final FeedPost post = await ref.read(socialRepositoryProvider).createPost(
            body: body,
            categoryId: categoryId,
            images: const <String>[],
            draft: draft,
          );
      state = const AsyncData(null);
      return post;
    } catch (e, st) {
      state = AsyncError(e, st);
      rethrow;
    }
  }
}
