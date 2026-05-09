import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../features/auth/application/auth_state_provider.dart';
import '../../features/auth/presentation/login_page.dart';
import '../../features/home/presentation/home_page.dart';
import 'auth_router_refresh.dart';

part 'go_router_provider.g.dart';

@Riverpod(keepAlive: true)
AuthRouterRefresh authRouterRefresh(AuthRouterRefreshRef ref) {
  final notifier = AuthRouterRefresh();
  ref.listen(authStateProvider, (AuthStatus? previous, AuthStatus next) {
    notifier.refresh();
  });
  ref.onDispose(notifier.dispose);
  return notifier;
}

@Riverpod(keepAlive: true)
GoRouter goRouter(GoRouterRef ref) {
  final refresh = ref.watch(authRouterRefreshProvider);
  return GoRouter(
    refreshListenable: refresh,
    initialLocation: '/',
    redirect: (BuildContext context, GoRouterState state) {
      final auth = ProviderScope.containerOf(context).read(authStateProvider);
      if (auth == AuthStatus.authenticated &&
          state.matchedLocation == '/login') {
        return '/';
      }
      return null;
    },
    routes: <RouteBase>[
      GoRoute(
        path: '/',
        builder: (BuildContext context, GoRouterState state) {
          return const HomePage();
        },
      ),
      GoRoute(
        path: '/login',
        builder: (BuildContext context, GoRouterState state) {
          return const LoginPage();
        },
      ),
    ],
  );
}
