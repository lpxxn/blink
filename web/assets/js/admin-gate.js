/**
 * BlinkAdmin.gate({ onReady }) — reusable super_admin gate.
 *
 * HTML contract (defaults, all overridable):
 *   <div id="gate">
 *     <p id="gate-lead">正在验证权限…</p>
 *   </div>
 *   <div id="app" hidden>…admin content…</div>
 *
 * Requires BlinkAPI to be loaded before this script.
 */
(function () {
  'use strict';

  async function gate(opts) {
    const cfg = Object.assign(
      { gateId: 'gate', appId: 'app', leadId: 'gate-lead', onReady: null },
      opts || {}
    );
    const gateEl = document.getElementById(cfg.gateId);
    const appEl = document.getElementById(cfg.appId);
    const leadEl = document.getElementById(cfg.leadId);

    function deny(message) {
      if (leadEl) leadEl.textContent = message;
      if (gateEl) gateEl.hidden = false;
      if (appEl) appEl.hidden = true;
    }

    function allow() {
      if (gateEl) gateEl.hidden = true;
      if (appEl) appEl.hidden = false;
    }

    deny('正在验证权限…');

    let me;
    try {
      me = await window.BlinkAPI.me();
    } catch (err) {
      deny('无法验证登录状态：' + (err && err.message ? err.message : err));
      return null;
    }

    if (!me) {
      deny('请先登录后再访问管理后台。');
      return null;
    }
    if (me.role !== 'super_admin') {
      deny('当前账号不是超级管理员，无法使用此页面。');
      return null;
    }

    allow();
    document.querySelectorAll('[data-blink-nav="admin"]').forEach((el) => {
      el.hidden = false;
    });
    if (typeof cfg.onReady === 'function') {
      try {
        cfg.onReady(me);
      } catch (err) {
        console.error('admin onReady failed', err);
      }
    }
    return me;
  }

  window.BlinkAdmin = { gate };
})();
