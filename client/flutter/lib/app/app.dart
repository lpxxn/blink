import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../features/auth/application/auth_hydration_provider.dart';
import 'router/go_router_provider.dart';

class BlinkApp extends ConsumerWidget {
  const BlinkApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    ref.watch(authHydrationProvider);
    final router = ref.watch(goRouterProvider);

    return MaterialApp.router(
      title: 'Blink',
      theme: ThemeData(
        useMaterial3: true,
        colorSchemeSeed: Colors.indigo,
      ),
      routerConfig: router,
    );
  }
}

