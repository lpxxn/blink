/* Shows "管理" nav links only when GET /api/me returns role super_admin. */
(function () {
  'use strict';

  function apiPrefix() {
    if (typeof window !== 'undefined' && typeof window.BLINK_API === 'string') {
      return window.BLINK_API;
    }
    if (typeof API !== 'undefined') {
      return API;
    }
    return '';
  }

  function applyAdminNavVisibility(me) {
    var isSuper = me && me.role === 'super_admin';
    document.querySelectorAll('[data-blink-nav="admin"]').forEach(function (el) {
      el.hidden = !isSuper;
    });
  }

  function run() {
    fetch(apiPrefix() + '/api/me', { credentials: 'include' })
      .then(function (r) {
        if (!r.ok) {
          return null;
        }
        return r.json();
      })
      .catch(function () {
        return null;
      })
      .then(applyAdminNavVisibility);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', run);
  } else {
    run();
  }
})();
