import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';

import 'app/app.dart';
import 'core/network/blink_cookie_jar_provider.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final overrides = <Override>[];
  if (!kIsWeb) {
    final supportDir = await getApplicationSupportDirectory();
    final jar = PersistCookieJar(
      storage: FileStorage('${supportDir.path}/blink_cookies'),
    );
    overrides.add(
      blinkCookieJarProvider.overrideWith((BlinkCookieJarRef ref) => jar),
    );
  }

  runApp(
    ProviderScope(
      overrides: overrides,
      child: const BlinkApp(),
    ),
  );
}
