// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'post_replies_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$postRepliesHash() => r'af8993a8f010825711eb01f35010da8e2d0d520c';

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

abstract class _$PostReplies
    extends BuildlessAutoDisposeAsyncNotifier<List<FeedReply>> {
  late final String postId;

  FutureOr<List<FeedReply>> build(String postId);
}

/// See also [PostReplies].
@ProviderFor(PostReplies)
const postRepliesProvider = PostRepliesFamily();

/// See also [PostReplies].
class PostRepliesFamily extends Family<AsyncValue<List<FeedReply>>> {
  /// See also [PostReplies].
  const PostRepliesFamily();

  /// See also [PostReplies].
  PostRepliesProvider call(String postId) {
    return PostRepliesProvider(postId);
  }

  @override
  PostRepliesProvider getProviderOverride(
    covariant PostRepliesProvider provider,
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
  String? get name => r'postRepliesProvider';
}

/// See also [PostReplies].
class PostRepliesProvider
    extends AutoDisposeAsyncNotifierProviderImpl<PostReplies, List<FeedReply>> {
  /// See also [PostReplies].
  PostRepliesProvider(String postId)
    : this._internal(
        () => PostReplies()..postId = postId,
        from: postRepliesProvider,
        name: r'postRepliesProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$postRepliesHash,
        dependencies: PostRepliesFamily._dependencies,
        allTransitiveDependencies: PostRepliesFamily._allTransitiveDependencies,
        postId: postId,
      );

  PostRepliesProvider._internal(
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
  FutureOr<List<FeedReply>> runNotifierBuild(covariant PostReplies notifier) {
    return notifier.build(postId);
  }

  @override
  Override overrideWith(PostReplies Function() create) {
    return ProviderOverride(
      origin: this,
      override: PostRepliesProvider._internal(
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
  AutoDisposeAsyncNotifierProviderElement<PostReplies, List<FeedReply>>
  createElement() {
    return _PostRepliesProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is PostRepliesProvider && other.postId == postId;
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
mixin PostRepliesRef on AutoDisposeAsyncNotifierProviderRef<List<FeedReply>> {
  /// The parameter `postId` of this provider.
  String get postId;
}

class _PostRepliesProviderElement
    extends
        AutoDisposeAsyncNotifierProviderElement<PostReplies, List<FeedReply>>
    with PostRepliesRef {
  _PostRepliesProviderElement(super.provider);

  @override
  String get postId => (origin as PostRepliesProvider).postId;
}

// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
