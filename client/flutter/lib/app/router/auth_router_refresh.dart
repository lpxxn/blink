import 'package:flutter/foundation.dart';

/// Drives [GoRouter.refresh] when auth changes.
final class AuthRouterRefresh extends ChangeNotifier {
  void refresh() => notifyListeners();
}
