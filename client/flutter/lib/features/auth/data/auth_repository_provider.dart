import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/network/blink_cookie_jar_provider.dart';
import '../../../core/network/dio_provider.dart';
import 'auth_repository.dart';

part 'auth_repository_provider.g.dart';

@riverpod
AuthRepository authRepository(AuthRepositoryRef ref) {
  return AuthRepository(
    ref.watch(dioProvider),
    ref.watch(blinkCookieJarProvider),
  );
}
