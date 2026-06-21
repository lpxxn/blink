/**
 * BlinkAuth.requireLogin(onReady) — redirect guests to login before page init.
 *
 * Requires BlinkAPI to be loaded before this script.
 */
(function () {
  'use strict';

  async function requireLogin(onReady) {
    const next = location.pathname + location.search;
    let me;
    try {
      me = await window.BlinkAPI.me();
    } catch (_) {
      me = null;
    }
    if (!me || me.user_id == null) {
      window.location.href = '/web/login.html?next=' + encodeURIComponent(next);
      return null;
    }
    if (typeof onReady === 'function') {
      try {
        await onReady(me);
      } catch (err) {
        console.error('auth onReady failed', err);
      }
    }
    return me;
  }

  window.BlinkAuth = { requireLogin };
})();
