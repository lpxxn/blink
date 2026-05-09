import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio/dio.dart';

import 'me.dart';

final class AuthRepository {
  AuthRepository(this._dio, this._cookieJar);

  final Dio _dio;
  final CookieJar _cookieJar;

  /// Email + password login; server sets `blink_session` cookie.
  Future<void> loginWithPassword({
    required String email,
    required String password,
  }) async {
    try {
      await _dio.post<void>(
        '/auth/login',
        data: <String, dynamic>{
          'email': email,
          'password': password,
        },
      );
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw AuthInvalidCredentialsException();
      }
      throw AuthRepositoryException(e.message ?? 'Login failed');
    }
  }

  /// Restore session using persisted cookies.
  Future<Me> fetchMe() async {
    try {
      final res = await _dio.get<Map<String, dynamic>>('/api/me');
      final data = res.data;
      if (data == null) {
        throw AuthRepositoryException('Empty response');
      }
      return Me.fromJson(data);
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        throw AuthInvalidCredentialsException();
      }
      throw AuthRepositoryException(e.message ?? 'Request failed');
    }
  }

  /// Clears server session cookie (`POST /api/logout`) and local jar (non-web).
  Future<void> logout() async {
    try {
      await _dio.post<void>('/api/logout');
    } on DioException {
      // Still clear local cookies; web relies on Set-Cookie from this response.
    }
    await _cookieJar.deleteAll();
  }
}

final class AuthInvalidCredentialsException implements Exception {}

final class AuthRepositoryException implements Exception {
  AuthRepositoryException(this.message);
  final String message;
}
