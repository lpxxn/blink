import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'reply_parent_target_provider.g.dart';

typedef ReplyParent = ({String id, String userName});

/// Which reply is being replied to (thread) for a given post detail screen.
@riverpod
class ReplyParentTarget extends _$ReplyParentTarget {
  @override
  ReplyParent? build(String postId) => null;

  void setParent(String id, String userName) => state = (id: id, userName: userName);

  void clear() => state = null;
}
