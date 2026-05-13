import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/env/app_env.dart';
import '../../../core/env/app_env_provider.dart';
import '../../../core/network/media_url.dart';
import '../../auth/application/auth_state_provider.dart';
import '../../feed/application/categories_provider.dart';
import '../../feed/application/public_feed_provider.dart';
import '../../feed/application/public_feed_query_provider.dart';
import '../../feed/data/feed_models.dart';

class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final AuthStatus auth = ref.watch(authStateProvider);
    final PublicFeedFilter query = ref.watch(publicFeedQueryProvider);
    final AsyncValue<List<FeedPost>> feed = ref.watch(publicFeedProvider);
    final AsyncValue<List<FeedCategory>> categories =
        ref.watch(feedCategoriesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Blink'),
        actions: [
          IconButton(
            tooltip: 'New post',
            icon: const Icon(Icons.post_add),
            onPressed: () => context.push('/compose'),
          ),
          TextButton(
            onPressed: () => context.push('/login'),
            child: const Text('Login'),
          ),
          TextButton(
            onPressed: () => context.push('/me'),
            child: const Text('Me'),
          ),
        ],
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text(
              switch (auth) {
                AuthStatus.unknown => 'Session: checking…',
                AuthStatus.guest => 'Session: guest',
                AuthStatus.authenticated => 'Session: signed in',
              },
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ),
          SizedBox(
            height: 44,
            child: categories.when(
              data: (List<FeedCategory> list) {
                return ListView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 12),
                  children: [
                    Padding(
                      padding: const EdgeInsets.only(right: 6),
                      child: FilterChip(
                        label: const Text('All'),
                        selected: !query.uncategorized &&
                            query.categoryId == null,
                        onSelected: (_) {
                          ref.read(publicFeedQueryProvider.notifier).selectAll();
                        },
                      ),
                    ),
                    Padding(
                      padding: const EdgeInsets.only(right: 6),
                      child: FilterChip(
                        label: const Text('Uncategorized'),
                        selected: query.uncategorized,
                        onSelected: (_) {
                          ref
                              .read(publicFeedQueryProvider.notifier)
                              .selectUncategorized();
                        },
                      ),
                    ),
                    ...list.map(
                      (FeedCategory c) => Padding(
                        padding: const EdgeInsets.only(right: 6),
                        child: FilterChip(
                          label: Text(c.name),
                          selected: query.categoryId == c.id &&
                              !query.uncategorized,
                          onSelected: (_) {
                            ref
                                .read(publicFeedQueryProvider.notifier)
                                .selectCategory(c.id);
                          },
                        ),
                      ),
                    ),
                  ],
                );
              },
              loading: () => const Center(
                child: SizedBox(
                  width: 22,
                  height: 22,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ),
              error: (Object e, StackTrace _) => const SizedBox.shrink(),
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: RefreshIndicator(
              onRefresh: () async {
                ref.invalidate(feedCategoriesProvider);
                ref.invalidate(publicFeedProvider);
                await ref.read(publicFeedProvider.future);
              },
              child: NotificationListener<ScrollNotification>(
                onNotification: (ScrollNotification n) {
                  if (n.metrics.extentAfter < 280) {
                    ref.read(publicFeedProvider.notifier).loadMore();
                  }
                  return false;
                },
                child: feed.when(
                  skipLoadingOnReload: true,
                  data: (List<FeedPost> posts) {
                    if (posts.isEmpty) {
                      return ListView(
                        physics: const AlwaysScrollableScrollPhysics(),
                        children: const [
                          SizedBox(height: 120),
                          Center(child: Text('No posts yet')),
                        ],
                      );
                    }
                    final AppEnv env = ref.watch(appEnvProvider);
                    return ListView.separated(
                      physics: const AlwaysScrollableScrollPhysics(),
                      itemCount: posts.length,
                      separatorBuilder: (BuildContext context, int index) =>
                          const Divider(height: 1),
                      itemBuilder: (BuildContext context, int i) {
                        final FeedPost p = posts[i];
                        return InkWell(
                          onTap: () => context.push('/posts/${p.id}'),
                          child: _PostTile(env: env, post: p),
                        );
                      },
                    );
                  },
                  loading: () => const Center(
                    child: CircularProgressIndicator(),
                  ),
                  error: (Object e, StackTrace stackTrace) => ListView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    children: [
                      const SizedBox(height: 48),
                      Padding(
                        padding: const EdgeInsets.all(16),
                        child: Text('$e'),
                      ),
                      Center(
                        child: FilledButton(
                          onPressed: () {
                            ref.invalidate(feedCategoriesProvider);
                            ref.invalidate(publicFeedProvider);
                          },
                          child: const Text('Retry'),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PostTile extends StatelessWidget {
  const _PostTile({required this.env, required this.post});

  final AppEnv env;
  final FeedPost post;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            post.userName,
            style: Theme.of(context).textTheme.titleSmall,
          ),
          const SizedBox(height: 4),
          Text(
            post.body,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
          if (post.images.isNotEmpty) ...[
            const SizedBox(height: 8),
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: Image.network(
                resolveMediaUrl(env, post.images.first),
                height: 160,
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (BuildContext context, Object error, _) =>
                    const SizedBox.shrink(),
              ),
            ),
          ],
          const SizedBox(height: 4),
          Text(
            post.createdAt,
            style: Theme.of(context).textTheme.labelSmall,
          ),
        ],
      ),
    );
  }
}
