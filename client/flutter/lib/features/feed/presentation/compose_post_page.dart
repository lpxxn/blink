import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../application/categories_provider.dart';
import '../application/create_post_submit_provider.dart';
import '../data/feed_models.dart';

class ComposePostPage extends ConsumerStatefulWidget {
  const ComposePostPage({super.key});

  @override
  ConsumerState<ComposePostPage> createState() => _ComposePostPageState();
}

class _ComposePostPageState extends ConsumerState<ComposePostPage> {
  final TextEditingController _body = TextEditingController();
  /// Selected category snowflake id; null until user picks (default = first in list at submit).
  String? _categoryId;
  bool _draft = false;

  String _effectiveCategoryId(List<FeedCategory> list) {
    if (_categoryId != null &&
        list.any((FeedCategory c) => c.id == _categoryId)) {
      return _categoryId!;
    }
    return list.first.id;
  }

  @override
  void dispose() {
    _body.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final AsyncValue<List<FeedCategory>> cats = ref.watch(feedCategoriesProvider);
    final AsyncValue<void> submit = ref.watch(createPostSubmitProvider);

    ref.listen(createPostSubmitProvider, (AsyncValue<void>? previous, AsyncValue<void> next) {
      if (next.hasError) {
        final Object e = next.error!;
        String msg = 'Failed to publish';
        if (e is DioException) {
          final dynamic d = e.response?.data;
          if (d is Map && d['error'] != null) {
            msg = d['error'].toString();
          } else if (e.message != null && e.message!.isNotEmpty) {
            msg = e.message!;
          }
        } else {
          msg = e.toString();
        }
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
      }
    });

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => context.pop(),
        ),
        title: const Text('New post'),
        actions: [
          TextButton(
            onPressed: submit.isLoading
                ? null
                : () async {
                    final List<FeedCategory>? list = cats.valueOrNull;
                    if (list == null || list.isEmpty) {
                      return;
                    }
                    final String cat = _effectiveCategoryId(list);
                    try {
                      final FeedPost post = await ref
                          .read(createPostSubmitProvider.notifier)
                          .submit(
                            body: _body.text,
                            categoryId: cat,
                            draft: _draft,
                          );
                      if (!context.mounted) {
                        return;
                      }
                      context.go('/posts/${post.id}');
                    } catch (_) {}
                  },
            child: submit.isLoading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Publish'),
          ),
        ],
      ),
      body: cats.when(
        data: (List<FeedCategory> list) {
          if (list.isEmpty) {
            return const Center(child: Text('No categories — cannot post'));
          }
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              InputDecorator(
                decoration: const InputDecoration(labelText: 'Category'),
                child: DropdownButtonHideUnderline(
                  child: DropdownButton<String>(
                    isExpanded: true,
                    value: _effectiveCategoryId(list),
                    items: list
                        .map(
                          (FeedCategory c) => DropdownMenuItem<String>(
                            value: c.id,
                            child: Text(c.name),
                          ),
                        )
                        .toList(),
                    onChanged: (String? v) {
                      if (v != null) {
                        setState(() => _categoryId = v);
                      }
                    },
                  ),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _body,
                minLines: 6,
                maxLines: 16,
                decoration: const InputDecoration(
                  labelText: 'Content',
                  alignLabelWithHint: true,
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              SwitchListTile(
                title: const Text('Save as draft'),
                value: _draft,
                onChanged: (bool v) => setState(() => _draft = v),
              ),
            ],
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (Object e, StackTrace _) => Center(child: Text('$e')),
      ),
    );
  }
}
