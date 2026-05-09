import 'package:dio/dio.dart';

import 'configure_dio_http_stub.dart'
    if (dart.library.html) 'configure_dio_http_web.dart' as impl;

void configureDioHttpAdapter(Dio dio) => impl.configureDioHttpAdapter(dio);
