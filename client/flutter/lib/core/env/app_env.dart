/// Runtime configuration from `--dart-define` (see docs/flutter-client-plan.md).
final class AppEnv {
  const AppEnv({
    required this.name,
    required this.apiBaseUrl,
  });

  /// e.g. `dev`, `staging`, `prod`
  final String name;

  /// Blink HTTP API base (no trailing slash required).
  final String apiBaseUrl;
}
