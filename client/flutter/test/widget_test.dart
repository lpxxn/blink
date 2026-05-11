// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:blink_client/app/app.dart';
import 'package:blink_client/features/auth/application/auth_hydration_provider.dart';
import 'package:blink_client/features/auth/application/auth_state_provider.dart';
import 'package:blink_client/features/feed/data/feed_models.dart'
    show FeedCategory, FeedPost, FeedReply, PostsPage, RepliesPage;
import 'package:blink_client/features/feed/data/social_repository.dart';
import 'package:blink_client/features/feed/data/social_repository_provider.dart';

class _StubSocialRepository extends SocialRepository {
  _StubSocialRepository() : super(Dio());

  @override
  Future<List<FeedCategory>> fetchCategories() async => const [];

  @override
  Future<PostsPage> fetchPosts({
    String? categoryId,
    bool uncategorized = false,
    String? cursor,
    int limit = 20,
  }) async =>
      const PostsPage(posts: []);

  @override
  Future<FeedPost> fetchPost(String id) async => FeedPost(
        id: id,
        userId: '0',
        userName: 'stub',
        body: 'stub',
        images: const [],
        createdAt: '',
      );

  @override
  Future<RepliesPage> fetchReplies(
    String postId, {
    String? cursor,
    int limit = 50,
  }) async =>
      const RepliesPage(replies: []);

  @override
  Future<FeedReply> createReply({
    required String postId,
    required String body,
    String? parentReplyId,
  }) async =>
      FeedReply(
        id: '0',
        postId: postId,
        userId: '0',
        userName: 'stub',
        body: body,
        createdAt: '',
        parentReplyId: parentReplyId,
      );
}

void main() {
  testWidgets('App boot smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          authHydrationProvider.overrideWith((ref) async {
            ref.read(authStateProvider.notifier).setGuest();
          }),
          socialRepositoryProvider.overrideWith(
            (ref) => _StubSocialRepository(),
          ),
        ],
        child: const BlinkApp(),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('Blink'), findsOneWidget);
    expect(find.text('Session: guest'), findsOneWidget);
    expect(find.text('Login'), findsOneWidget);
    expect(find.text('Me'), findsOneWidget);
  });
}
