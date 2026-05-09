import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';

import 'app/app.dart';
import 'core/network/blink_cookie_jar_provider.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final supportDir = await getApplicationSupportDirectory();
  final CookieJar jar = PersistCookieJar(
    storage: FileStorage('${supportDir.path}/blink_cookies'),
  );

  runApp(
    ProviderScope(
      overrides: [
        blinkCookieJarProvider.overrideWith((BlinkCookieJarRef ref) => jar),
      ],
      child: const BlinkApp(),
    ),
  );
}
