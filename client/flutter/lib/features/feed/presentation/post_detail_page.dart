import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/env/app_env.dart';
import '../../../core/env/app_env_provider.dart';
import '../../../core/network/media_url.dart';
import '../application/post_detail_provider.dart';
import '../application/post_replies_provider.dart';
import '../data/feed_models.dart';

class PostDetailPage extends ConsumerWidget {
  const PostDetailPage({super.key, required this.postId});

  final String postId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final AsyncValue<FeedPost> postAsync = ref.watch(postDetailProvider(postId));
    final AsyncValue<List<FeedReply>> repliesAsync =
        ref.watch(postRepliesProvider(postId));
    final AppEnv env = ref.watch(appEnvProvider);

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.pop(),
        ),
        title: const Text('Post'),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(postDetailProvider(postId));
          ref.invalidate(postRepliesProvider(postId));
          await ref.read(postDetailProvider(postId).future);
          await ref.read(postRepliesProvider(postId).future);
        },
        child: NotificationListener<ScrollNotification>(
          onNotification: (ScrollNotification n) {
            if (n.metrics.extentAfter < 240) {
              ref.read(postRepliesProvider(postId).notifier).loadMore();
            }
            return false;
          },
          child: CustomScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            slivers: [
              SliverToBoxAdapter(
                child: postAsync.when(
                  data: (FeedPost p) => _PostHeader(env: env, post: p),
                  loading: () => const Padding(
                    padding: EdgeInsets.all(24),
                    child: Center(child: CircularProgressIndicator()),
                  ),
                  error: (Object e, StackTrace _) => Padding(
                    padding: const EdgeInsets.all(16),
                    child: Text('$e'),
                  ),
                ),
              ),
              const SliverToBoxAdapter(
                child: Padding(
                  padding: EdgeInsets.fromLTRB(16, 8, 16, 4),
                  child: Text(
                    'Replies',
                    style: TextStyle(fontWeight: FontWeight.w600),
                  ),
                ),
              ),
              repliesAsync.when(
                data: (List<FeedReply> list) {
                  if (list.isEmpty) {
                    return const SliverToBoxAdapter(
                      child: Padding(
                        padding: EdgeInsets.all(16),
                        child: Text('No replies yet'),
                      ),
                    );
                  }
                  return SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (BuildContext context, int i) {
                        final FeedReply r = list[i];
                        return Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            if (i > 0) const Divider(height: 1),
                            ListTile(
                              dense: true,
                              title: Text(
                                r.userName,
                                style: Theme.of(context).textTheme.titleSmall,
                              ),
                              subtitle: Text(r.body),
                            ),
                          ],
                        );
                      },
                      childCount: list.length,
                    ),
                  );
                },
                loading: () => const SliverToBoxAdapter(
                  child: Padding(
                    padding: EdgeInsets.all(24),
                    child: Center(child: CircularProgressIndicator()),
                  ),
                ),
                error: (Object e, StackTrace _) => SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Text('$e'),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _PostHeader extends StatelessWidget {
  const _PostHeader({required this.env, required this.post});

  final AppEnv env;
  final FeedPost post;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(post.userName, style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Text(post.body, style: Theme.of(context).textTheme.bodyLarge),
          if (post.images.isNotEmpty) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: Image.network(
                resolveMediaUrl(env, post.images.first),
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (BuildContext context, Object error, _) =>
                    const SizedBox.shrink(),
              ),
            ),
          ],
          const SizedBox(height: 8),
          Text(
            post.createdAt,
            style: Theme.of(context).textTheme.labelSmall,
          ),
        ],
      ),
    );
  }
}
