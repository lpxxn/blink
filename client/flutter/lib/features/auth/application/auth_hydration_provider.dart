import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/auth_repository.dart';
import '../data/auth_repository_provider.dart';
import 'auth_state_provider.dart';

part 'auth_hydration_provider.g.dart';

/// On cold start: validate persisted cookies via `GET /api/me`.
@Riverpod(keepAlive: true)
Future<void> authHydration(AuthHydrationRef ref) async {
  try {
    await ref.read(authRepositoryProvider).fetchMe();
    ref.read(authStateProvider.notifier).setAuthenticated();
  } on AuthInvalidCredentialsException {
    ref.read(authStateProvider.notifier).setGuest();
  } catch (_) {
    ref.read(authStateProvider.notifier).setGuest();
  }
}
