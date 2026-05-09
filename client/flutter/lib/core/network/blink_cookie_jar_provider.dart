import 'package:cookie_jar/cookie_jar.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

part 'blink_cookie_jar_provider.g.dart';

/// Default: in-memory (tests / quick runs). Override in [main] with [PersistCookieJar].
@Riverpod(keepAlive: true)
CookieJar blinkCookieJar(BlinkCookieJarRef ref) {
  return CookieJar();
}
