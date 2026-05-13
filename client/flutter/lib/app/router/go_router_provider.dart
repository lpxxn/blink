import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../features/auth/application/auth_state_provider.dart';
import '../../features/auth/presentation/login_page.dart';
import '../../features/auth/presentation/me_page.dart';
import '../../features/feed/presentation/compose_post_page.dart';
import '../../features/feed/presentation/post_detail_page.dart';
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
      final String loc = state.matchedLocation;

      if (auth == AuthStatus.authenticated && loc == '/login') {
        return '/';
      }

      const Set<String> protectedPaths = {'/me', '/compose'};
      if (auth == AuthStatus.unknown && protectedPaths.contains(loc)) {
        return '/';
      }
      if (protectedPaths.contains(loc) && auth == AuthStatus.guest) {
        return '/login';
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
      GoRoute(
        path: '/me',
        builder: (BuildContext context, GoRouterState state) {
          return const MePage();
        },
      ),
      GoRoute(
        path: '/compose',
        builder: (BuildContext context, GoRouterState state) {
          return const ComposePostPage();
        },
      ),
      GoRoute(
        path: '/posts/:postId',
        builder: (BuildContext context, GoRouterState state) {
          final String id = state.pathParameters['postId'] ?? '';
          return PostDetailPage(postId: id);
        },
      ),
    ],
  );
}
