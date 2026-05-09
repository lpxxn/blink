import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/feed_models.dart';
import '../data/social_repository_provider.dart';

part 'categories_provider.g.dart';

@riverpod
Future<List<FeedCategory>> feedCategories(FeedCategoriesRef ref) {
  return ref.watch(socialRepositoryProvider).fetchCategories();
}
