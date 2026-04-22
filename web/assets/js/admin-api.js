/**
 * BlinkAdminAPI — typed wrappers for /admin/api/* endpoints.
 *
 * All calls go through BlinkAPI so BlinkError normalization applies.
 * Requires BlinkAPI to be loaded first.
 */
(function () {
  'use strict';

  const API = window.BlinkAPI;
  if (!API) throw new Error('BlinkAdminAPI requires BlinkAPI');

  function qs(params) {
    const parts = [];
    Object.keys(params || {}).forEach((k) => {
      const v = params[k];
      if (v === undefined || v === null || v === '') return;
      parts.push(encodeURIComponent(k) + '=' + encodeURIComponent(String(v)));
    });
    return parts.length ? '?' + parts.join('&') : '';
  }

  const BlinkAdminAPI = {
    overview: () => API.get('/admin/api/overview'),

    // ---- users ----
    listUsers: (params) => API.get('/admin/api/users' + qs(params || { limit: 50 })),
    patchUser: (id, body) => API.patch('/admin/api/users/' + encodeURIComponent(id), body),
    resetUserPassword: (id, password) =>
      API.post('/admin/api/users/' + encodeURIComponent(id) + '/reset_password', { password }),

    // ---- posts ----
    listPosts: (params) => API.get('/admin/api/posts' + qs(params || { limit: 20, offset: 0 })),
    patchPost: (id, body) => API.patch('/admin/api/posts/' + encodeURIComponent(id), body),
    resolveAppeal: (id, approve, note) =>
      API.post('/admin/api/posts/' + encodeURIComponent(id) + '/resolve_appeal', {
        approve: !!approve,
        note: note || '',
      }),
    listPostReplies: (postId, params) =>
      API.get('/admin/api/posts/' + encodeURIComponent(postId) + '/replies' + qs(params || {})),

    // ---- replies ----
    patchReply: (id, body) => API.patch('/admin/api/replies/' + encodeURIComponent(id), body),
    hideReplyCascade: (id) => API.patch('/admin/api/replies/' + encodeURIComponent(id), { hidden: true }),
    unhideReplyCascade: (id) => API.patch('/admin/api/replies/' + encodeURIComponent(id), { hidden: false }),

    // ---- sensitive words ----
    listSensitiveWords: (params) =>
      API.get('/admin/api/sensitive_words' + qs(params || { limit: 50, offset: 0 })),
    createSensitiveWord: (word) => API.post('/admin/api/sensitive_words', { word }),
    patchSensitiveWord: (id, body) => API.patch('/admin/api/sensitive_words/' + encodeURIComponent(id), body),
    deleteSensitiveWord: (id) => API.del('/admin/api/sensitive_words/' + encodeURIComponent(id)),

    // ---- settings ----
    getSensitivePostMode: () => API.get('/admin/api/settings/sensitive_post_mode'),
    setSensitivePostMode: (mode) => API.put('/admin/api/settings/sensitive_post_mode', { mode }),
    getRegisterEmailVerification: () => API.get('/admin/api/settings/register_email_verification'),
    setRegisterEmailVerification: (required) =>
      API.put('/admin/api/settings/register_email_verification', { required: !!required }),

    // ---- SMTP ----
    getSMTPSettings: () => API.get('/admin/api/settings/smtp'),
    setSMTPSettings: (body) => API.put('/admin/api/settings/smtp', body),
    testSMTP: (to) => API.post('/admin/api/settings/smtp/test', { to }),
  };

  // ---- enum labels ----
  BlinkAdminAPI.modLabel = function (m) {
    if (m === 0) return '正常';
    if (m === 1) return '违规';
    if (m === 2) return '下架';
    return String(m);
  };
  BlinkAdminAPI.postStatusLabel = function (s) {
    if (s === 0) return '草稿';
    if (s === 1) return '已发布';
    if (s === 2) return '隐藏';
    return String(s);
  };
  BlinkAdminAPI.userStatusLabel = function (s) {
    if (s === 0) return '未激活';
    if (s === 1) return '正常';
    if (s === 2) return '封禁';
    return String(s);
  };
  BlinkAdminAPI.replyStatusLabel = function (s) {
    if (s === 0) return '可见';
    if (s === 1) return '隐藏';
    return String(s);
  };

  window.BlinkAdminAPI = BlinkAdminAPI;
})();
