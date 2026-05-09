import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'auth_state_provider.g.dart';

enum AuthStatus {
  /// Session not resolved yet (e.g. cold start restore).
  unknown,

  /// No authenticated session (guest).
  guest,

  /// Valid session (cookie / token established).
  authenticated,
}

@riverpod
class AuthState extends _$AuthState {
  @override
  AuthStatus build() => AuthStatus.unknown;

  void setUnknown() => state = AuthStatus.unknown;
  void setGuest() => state = AuthStatus.guest;
  void setAuthenticated() => state = AuthStatus.authenticated;
}
