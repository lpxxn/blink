import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../../core/network/dio_provider.dart';
import 'social_repository.dart';

part 'social_repository_provider.g.dart';

@riverpod
SocialRepository socialRepository(SocialRepositoryRef ref) {
  return SocialRepository(ref.watch(dioProvider));
}
