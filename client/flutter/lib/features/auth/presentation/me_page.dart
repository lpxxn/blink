import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../application/auth_state_provider.dart';
import '../application/me_provider.dart';
import '../data/auth_repository_provider.dart';
import '../data/me.dart';

class MePage extends ConsumerWidget {
  const MePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final meAsync = ref.watch(meProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Me'),
        actions: [
          TextButton(
            onPressed: () async {
              await ref.read(authRepositoryProvider).logout();
              ref.read(authStateProvider.notifier).setGuest();
              if (context.mounted) {
                context.go('/');
              }
            },
            child: const Text('Logout'),
          ),
        ],
      ),
      body: meAsync.when(
        data: (Me? me) {
          if (me == null) {
            return const Center(child: CircularProgressIndicator());
          }
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text('User ID: ${me.userId}'),
              const SizedBox(height: 8),
              Text('Email: ${me.email}'),
              const SizedBox(height: 8),
              Text('Name: ${me.name}'),
              if (me.role != null) ...[
                const SizedBox(height: 8),
                Text('Role: ${me.role}'),
              ],
            ],
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (Object e, StackTrace _) => Center(child: Text('$e')),
      ),
    );
  }
}
