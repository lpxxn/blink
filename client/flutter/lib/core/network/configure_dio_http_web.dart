import 'package:dio/browser.dart';
import 'package:dio/dio.dart';

/// Web: [dio_cookie_manager] is unsupported; use the browser stack and send
/// cookies on cross-origin XHR when the API enables CORS with credentials.
void configureDioHttpAdapter(Dio dio) {
  dio.httpClientAdapter = BrowserHttpClientAdapter(withCredentials: true);
}
