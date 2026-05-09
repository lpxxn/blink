import '../env/app_env.dart';

/// Resolves relative upload paths (`/uploads/...`) against API base URL.
String resolveMediaUrl(AppEnv env, String path) {
  if (path.isEmpty) {
    return path;
  }
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  final String base = env.apiBaseUrl.replaceAll(RegExp(r'/+$'), '');
  final String p = path.startsWith('/') ? path : '/$path';
  return '$base$p';
}
