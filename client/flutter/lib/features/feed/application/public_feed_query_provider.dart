import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'public_feed_query_provider.g.dart';

/// Filter for `GET /api/posts`: all / uncategorized / one category.
typedef PublicFeedFilter = ({String? categoryId, bool uncategorized});

@riverpod
class PublicFeedQuery extends _$PublicFeedQuery {
  @override
  PublicFeedFilter build() =>
      (categoryId: null, uncategorized: false);

  void selectAll() =>
      state = (categoryId: null, uncategorized: false);

  void selectUncategorized() =>
      state = (categoryId: null, uncategorized: true);

  void selectCategory(String id) =>
      state = (categoryId: id, uncategorized: false);
}
