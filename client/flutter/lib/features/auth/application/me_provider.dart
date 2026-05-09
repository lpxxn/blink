import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/auth_repository_provider.dart';
import '../data/me.dart';
import 'auth_state_provider.dart';

part 'me_provider.g.dart';

/// Current user profile. `null` when not authenticated (avoids stale data after 401/logout).
@riverpod
Future<Me?> me(MeRef ref) async {
  final auth = ref.watch(authStateProvider);
  if (auth != AuthStatus.authenticated) {
    return null;
  }
  return ref.read(authRepositoryProvider).fetchMe();
}
