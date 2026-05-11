import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../auth/application/auth_state_provider.dart';
import '../application/create_reply_submit_provider.dart';

class ReplyComposerBar extends ConsumerStatefulWidget {
  const ReplyComposerBar({super.key, required this.postId});

  final String postId;

  @override
  ConsumerState<ReplyComposerBar> createState() => _ReplyComposerBarState();
}

class _ReplyComposerBarState extends ConsumerState<ReplyComposerBar> {
  final TextEditingController _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final AuthStatus auth = ref.watch(authStateProvider);
    final AsyncValue<void> submit =
        ref.watch(createReplySubmitProvider(widget.postId));

    ref.listen(createReplySubmitProvider(widget.postId),
        (AsyncValue<void>? previous, AsyncValue<void> next) {
      if (next.hasError) {
        final Object e = next.error!;
        String msg = 'Failed to post reply';
        if (e is DioException) {
          final dynamic d = e.response?.data;
          if (d is Map && d['error'] != null) {
            msg = d['error'].toString();
          } else if (e.response?.statusCode == 401) {
            msg = 'Please sign in to reply';
          } else if (e.message != null && e.message!.isNotEmpty) {
            msg = e.message!;
          }
        } else {
          msg = e.toString();
        }
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
      } else if (!next.isLoading &&
          !next.hasError &&
          next.hasValue &&
          previous?.isLoading == true) {
        _controller.clear();
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Reply posted')),
        );
      }
    });

    if (auth != AuthStatus.authenticated) {
      return Material(
        elevation: 8,
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    'Sign in to reply',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
                TextButton(
                  onPressed: () => context.push('/login'),
                  child: const Text('Login'),
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Material(
      elevation: 8,
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 8, 12, 12),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(
                child: TextField(
                  controller: _controller,
                  minLines: 1,
                  maxLines: 4,
                  decoration: const InputDecoration(
                    hintText: 'Write a reply…',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                  textInputAction: TextInputAction.newline,
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(
                onPressed: submit.isLoading
                    ? null
                    : () async {
                        await ref
                            .read(createReplySubmitProvider(widget.postId).notifier)
                            .submit(_controller.text);
                      },
                child: submit.isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Send'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
