import 'package:riverpod_annotation/riverpod_annotation.dart';

import 'app_env.dart';

part 'app_env_provider.g.dart';

@riverpod
AppEnv appEnv(AppEnvRef ref) {
  const name = String.fromEnvironment('ENV', defaultValue: 'dev');
  const baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://127.0.0.1:11110',
  );
  return AppEnv(name: name, apiBaseUrl: baseUrl);
}
