// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'blink_cookie_jar_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$blinkCookieJarHash() => r'cf67d2a97d12c3e496fe420095e0328b121b3eae';

/// Default: in-memory (tests / quick runs). Override in [main] with [PersistCookieJar].
///
/// Copied from [blinkCookieJar].
@ProviderFor(blinkCookieJar)
final blinkCookieJarProvider = Provider<CookieJar>.internal(
  blinkCookieJar,
  name: r'blinkCookieJarProvider',
  debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
      ? null
      : _$blinkCookieJarHash,
  dependencies: null,
  allTransitiveDependencies: null,
);

@Deprecated('Will be removed in 3.0. Use Ref instead')
// ignore: unused_element
typedef BlinkCookieJarRef = ProviderRef<CookieJar>;
// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
