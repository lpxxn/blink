// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'reply_parent_target_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$replyParentTargetHash() => r'f9e29f605a873c5f28ac8c944f9f9cae1f1c478f';

/// Copied from Dart SDK
class _SystemHash {
  _SystemHash._();

  static int combine(int hash, int value) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + value);
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x0007ffff & hash) << 10));
    return hash ^ (hash >> 6);
  }

  static int finish(int hash) {
    // ignore: parameter_assignments
    hash = 0x1fffffff & (hash + ((0x03ffffff & hash) << 3));
    // ignore: parameter_assignments
    hash = hash ^ (hash >> 11);
    return 0x1fffffff & (hash + ((0x00003fff & hash) << 15));
  }
}

abstract class _$ReplyParentTarget
    extends BuildlessAutoDisposeNotifier<ReplyParent?> {
  late final String postId;

  ReplyParent? build(String postId);
}

/// Which reply is being replied to (thread) for a given post detail screen.
///
/// Copied from [ReplyParentTarget].
@ProviderFor(ReplyParentTarget)
const replyParentTargetProvider = ReplyParentTargetFamily();

/// Which reply is being replied to (thread) for a given post detail screen.
///
/// Copied from [ReplyParentTarget].
class ReplyParentTargetFamily extends Family<ReplyParent?> {
  /// Which reply is being replied to (thread) for a given post detail screen.
  ///
  /// Copied from [ReplyParentTarget].
  const ReplyParentTargetFamily();

  /// Which reply is being replied to (thread) for a given post detail screen.
  ///
  /// Copied from [ReplyParentTarget].
  ReplyParentTargetProvider call(String postId) {
    return ReplyParentTargetProvider(postId);
  }

  @override
  ReplyParentTargetProvider getProviderOverride(
    covariant ReplyParentTargetProvider provider,
  ) {
    return call(provider.postId);
  }

  static const Iterable<ProviderOrFamily>? _dependencies = null;

  @override
  Iterable<ProviderOrFamily>? get dependencies => _dependencies;

  static const Iterable<ProviderOrFamily>? _allTransitiveDependencies = null;

  @override
  Iterable<ProviderOrFamily>? get allTransitiveDependencies =>
      _allTransitiveDependencies;

  @override
  String? get name => r'replyParentTargetProvider';
}

/// Which reply is being replied to (thread) for a given post detail screen.
///
/// Copied from [ReplyParentTarget].
class ReplyParentTargetProvider
    extends AutoDisposeNotifierProviderImpl<ReplyParentTarget, ReplyParent?> {
  /// Which reply is being replied to (thread) for a given post detail screen.
  ///
  /// Copied from [ReplyParentTarget].
  ReplyParentTargetProvider(String postId)
    : this._internal(
        () => ReplyParentTarget()..postId = postId,
        from: replyParentTargetProvider,
        name: r'replyParentTargetProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$replyParentTargetHash,
        dependencies: ReplyParentTargetFamily._dependencies,
        allTransitiveDependencies:
            ReplyParentTargetFamily._allTransitiveDependencies,
        postId: postId,
      );

  ReplyParentTargetProvider._internal(
    super._createNotifier, {
    required super.name,
    required super.dependencies,
    required super.allTransitiveDependencies,
    required super.debugGetCreateSourceHash,
    required super.from,
    required this.postId,
  }) : super.internal();

  final String postId;

  @override
  ReplyParent? runNotifierBuild(covariant ReplyParentTarget notifier) {
    return notifier.build(postId);
  }

  @override
  Override overrideWith(ReplyParentTarget Function() create) {
    return ProviderOverride(
      origin: this,
      override: ReplyParentTargetProvider._internal(
        () => create()..postId = postId,
        from: from,
        name: null,
        dependencies: null,
        allTransitiveDependencies: null,
        debugGetCreateSourceHash: null,
        postId: postId,
      ),
    );
  }

  @override
  AutoDisposeNotifierProviderElement<ReplyParentTarget, ReplyParent?>
  createElement() {
    return _ReplyParentTargetProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is ReplyParentTargetProvider && other.postId == postId;
  }

  @override
  int get hashCode {
    var hash = _SystemHash.combine(0, runtimeType.hashCode);
    hash = _SystemHash.combine(hash, postId.hashCode);

    return _SystemHash.finish(hash);
  }
}

@Deprecated('Will be removed in 3.0. Use Ref instead')
// ignore: unused_element
mixin ReplyParentTargetRef on AutoDisposeNotifierProviderRef<ReplyParent?> {
  /// The parameter `postId` of this provider.
  String get postId;
}

class _ReplyParentTargetProviderElement
    extends AutoDisposeNotifierProviderElement<ReplyParentTarget, ReplyParent?>
    with ReplyParentTargetRef {
  _ReplyParentTargetProviderElement(super.provider);

  @override
  String get postId => (origin as ReplyParentTargetProvider).postId;
}

// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
