import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:flutter/foundation.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';

import '../../features/auth/application/auth_state_provider.dart';
import 'blink_cookie_jar_provider.dart';
import 'configure_dio_http.dart';
import '../env/app_env_provider.dart';

part 'dio_provider.g.dart';

@riverpod
Dio dio(DioRef ref) {
  final env = ref.watch(appEnvProvider);
  final dio = Dio(
    BaseOptions(
      baseUrl: env.apiBaseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 20),
      sendTimeout: const Duration(seconds: 20),
      headers: <String, Object?>{
        'Accept': 'application/json',
      },
    ),
  );

  // dio_cookie_manager asserts on web; the browser stores/sends cookies when
  // [BrowserHttpClientAdapter.withCredentials] is true (see configure_dio_http_web.dart).
  if (kIsWeb) {
    configureDioHttpAdapter(dio);
  } else {
    dio.interceptors.add(CookieManager(ref.watch(blinkCookieJarProvider)));
  }

  if (kDebugMode) {
    dio.interceptors.add(
      LogInterceptor(
        requestHeader: true,
        requestBody: true,
        responseHeader: true,
        responseBody: true,
      ),
    );
  }

  dio.interceptors.add(
    InterceptorsWrapper(
      onError: (DioException e, ErrorInterceptorHandler handler) {
        if (e.response?.statusCode == 401) {
          ref.read(authStateProvider.notifier).setGuest();
        }
        return handler.next(e);
      },
    ),
  );

  return dio;
}

