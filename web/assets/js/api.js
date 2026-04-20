/**
 * BlinkAPI — unified fetch wrapper.
 *
 * All requests include credentials, serialize JSON bodies, and normalize
 * non-ok responses into `BlinkAPI.Error { status, message, body }` so
 * pages only need to show `err.message`.
 *
 * Usage:
 *   const me = await BlinkAPI.me();
 *   await BlinkAPI.post('/api/posts', { body, images });
 *   try { ... } catch (err) { BlinkUI.flash('msg', err.message); }
 */
(function () {
  'use strict';

  class BlinkError extends Error {
    constructor(status, message, body) {
      super(message);
      this.name = 'BlinkError';
      this.status = status;
      this.body = body != null ? body : null;
    }
  }

  const DEFAULT_ERRORS = {
    400: '请求参数无效',
    401: '请先登录',
    403: '没有权限',
    404: '资源不存在',
    409: '状态冲突',
    500: '服务端错误',
  };

  function apiBase() {
    if (typeof window.BLINK_API === 'string') return window.BLINK_API;
    return '';
  }

  async function parseBody(res) {
    if (res.status === 204) return null;
    const ct = res.headers.get('content-type') || '';
    if (ct.indexOf('application/json') !== -1) {
      try {
        return await res.json();
      } catch (_) {
        return null;
      }
    }
    try {
      const t = await res.text();
      return t === '' ? null : t;
    } catch (_) {
      return null;
    }
  }

  function pickMessage(body, status) {
    if (body && typeof body === 'object' && body.error) return String(body.error);
    if (typeof body === 'string' && body.trim()) return body.trim();
    return DEFAULT_ERRORS[status] || `请求失败 (${status})`;
  }

  async function request(method, path, opts) {
    const options = opts || {};
    const init = {
      method,
      credentials: 'include',
      headers: {},
    };
    const body = options.body;
    if (body instanceof FormData) {
      init.body = body;
    } else if (body !== undefined && body !== null) {
      init.headers['Content-Type'] = 'application/json';
      init.body = JSON.stringify(body);
    }
    if (options.headers) Object.assign(init.headers, options.headers);

    let res;
    try {
      res = await fetch(apiBase() + path, init);
    } catch (netErr) {
      throw new BlinkError(0, '网络异常，请稍后重试', null);
    }
    const parsed = await parseBody(res);
    if (!res.ok) {
      throw new BlinkError(res.status, pickMessage(parsed, res.status), parsed);
    }
    return parsed;
  }

  const API = {
    Error: BlinkError,
    request,
    get: (path, opts) => request('GET', path, opts),
    post: (path, body, opts) => request('POST', path, Object.assign({}, opts, { body })),
    patch: (path, body, opts) => request('PATCH', path, Object.assign({}, opts, { body })),
    put: (path, body, opts) => request('PUT', path, Object.assign({}, opts, { body })),
    del: (path, opts) => request('DELETE', path, opts),
    /** Returns the current user, or null if not logged in (401). */
    me: async () => {
      try {
        return await request('GET', '/api/me');
      } catch (err) {
        if (err instanceof BlinkError && err.status === 401) return null;
        throw err;
      }
    },
    logout: () => request('POST', '/api/logout'),
  };

  window.BlinkAPI = API;
})();
