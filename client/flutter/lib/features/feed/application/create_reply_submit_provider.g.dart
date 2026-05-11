// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'create_reply_submit_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

String _$createReplySubmitHash() => r'f0a99616035e4db2a842f58df2e8e70aa6be674a';

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

abstract class _$CreateReplySubmit
    extends BuildlessAutoDisposeNotifier<AsyncValue<void>> {
  late final String postId;

  AsyncValue<void> build(String postId);
}

/// See also [CreateReplySubmit].
@ProviderFor(CreateReplySubmit)
const createReplySubmitProvider = CreateReplySubmitFamily();

/// See also [CreateReplySubmit].
class CreateReplySubmitFamily extends Family<AsyncValue<void>> {
  /// See also [CreateReplySubmit].
  const CreateReplySubmitFamily();

  /// See also [CreateReplySubmit].
  CreateReplySubmitProvider call(String postId) {
    return CreateReplySubmitProvider(postId);
  }

  @override
  CreateReplySubmitProvider getProviderOverride(
    covariant CreateReplySubmitProvider provider,
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
  String? get name => r'createReplySubmitProvider';
}

/// See also [CreateReplySubmit].
class CreateReplySubmitProvider
    extends
        AutoDisposeNotifierProviderImpl<CreateReplySubmit, AsyncValue<void>> {
  /// See also [CreateReplySubmit].
  CreateReplySubmitProvider(String postId)
    : this._internal(
        () => CreateReplySubmit()..postId = postId,
        from: createReplySubmitProvider,
        name: r'createReplySubmitProvider',
        debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
            ? null
            : _$createReplySubmitHash,
        dependencies: CreateReplySubmitFamily._dependencies,
        allTransitiveDependencies:
            CreateReplySubmitFamily._allTransitiveDependencies,
        postId: postId,
      );

  CreateReplySubmitProvider._internal(
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
  AsyncValue<void> runNotifierBuild(covariant CreateReplySubmit notifier) {
    return notifier.build(postId);
  }

  @override
  Override overrideWith(CreateReplySubmit Function() create) {
    return ProviderOverride(
      origin: this,
      override: CreateReplySubmitProvider._internal(
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
  AutoDisposeNotifierProviderElement<CreateReplySubmit, AsyncValue<void>>
  createElement() {
    return _CreateReplySubmitProviderElement(this);
  }

  @override
  bool operator ==(Object other) {
    return other is CreateReplySubmitProvider && other.postId == postId;
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
mixin CreateReplySubmitRef on AutoDisposeNotifierProviderRef<AsyncValue<void>> {
  /// The parameter `postId` of this provider.
  String get postId;
}

class _CreateReplySubmitProviderElement
    extends
        AutoDisposeNotifierProviderElement<CreateReplySubmit, AsyncValue<void>>
    with CreateReplySubmitRef {
  _CreateReplySubmitProviderElement(super.provider);

  @override
  String get postId => (origin as CreateReplySubmitProvider).postId;
}

// ignore_for_file: type=lint
// ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member, invalid_use_of_visible_for_testing_member, deprecated_member_use_from_same_package
