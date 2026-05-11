import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/social_repository_provider.dart';
import 'post_replies_provider.dart';

part 'create_reply_submit_provider.g.dart';

@riverpod
class CreateReplySubmit extends _$CreateReplySubmit {
  @override
  AsyncValue<void> build(String postId) => const AsyncData(null);

  Future<void> submit(String body, {String? parentReplyId}) async {
    final String trimmed = body.trim();
    if (trimmed.isEmpty) {
      return;
    }
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      await ref.read(socialRepositoryProvider).createReply(
            postId: postId,
            body: trimmed,
            parentReplyId: parentReplyId,
          );
      ref.invalidate(postRepliesProvider(postId));
    });
  }
}
