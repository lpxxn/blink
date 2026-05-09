import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../data/auth_repository_provider.dart';
import 'auth_state_provider.dart';

part 'login_submit_provider.g.dart';

@riverpod
class LoginSubmit extends _$LoginSubmit {
  @override
  AsyncValue<void> build() => const AsyncData(null);

  Future<void> submit(String email, String password) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      await ref.read(authRepositoryProvider).loginWithPassword(
            email: email,
            password: password,
          );
      ref.read(authStateProvider.notifier).setAuthenticated();
    });
  }
}
