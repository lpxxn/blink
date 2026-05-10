import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/feed_models.dart';
import '../data/social_repository_provider.dart';

part 'post_detail_provider.g.dart';

@riverpod
Future<FeedPost> postDetail(PostDetailRef ref, String id) {
  return ref.watch(socialRepositoryProvider).fetchPost(id);
}
