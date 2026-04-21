/**
 * <blink-nav active="feed"></blink-nav>
 *
 * Renders the site's top navigation. Replaces itself with a <nav.top-nav>
 * so existing CSS selectors keep working. Shows the "管理" link only
 * when the current user has role === 'super_admin'.
 *
 * Requires: /web/assets/js/api.js (for BlinkAPI.me). Falls back to a
 * plain fetch if BlinkAPI is not loaded (e.g. during incremental rollout).
 */
(function () {
  'use strict';

  const LINKS = [
    { key: 'home',     href: '/web/index.html',    label: '首页' },
    { key: 'register', href: '/web/register.html', label: '注册', guestOnly: true },
    { key: 'login',    href: '/web/login.html',    label: '登录', guestOnly: true },
    { key: 'feed',     href: '/web/feed.html',     label: '帖子流' },
    { key: 'compose',  href: '/web/compose.html',  label: '发帖' },
    { key: 'mine',     href: '/web/mine.html',     label: '我的' },
    { key: 'messages', href: '/web/messages.html', label: '消息' },
    { key: 'admin',    href: '/web/admin.html',    label: '管理', adminOnly: true },
  ];

  let pending = null;
  let cachedMe;
  let cacheReady = false;

  function fetchMe() {
    if (cacheReady) return Promise.resolve(cachedMe);
    if (pending) return pending;
    if (window.BlinkAPI && typeof window.BlinkAPI.me === 'function') {
      pending = window.BlinkAPI.me().catch(() => null);
    } else {
      pending = fetch('/api/me', { credentials: 'include' })
        .then((r) => (r.ok ? r.json() : null))
        .catch(() => null);
    }
    return pending.then((me) => {
      cachedMe = me;
      cacheReady = true;
      pending = null;
      return me;
    });
  }

  function applyAdminVisibility(me) {
    const isSuper = !!(me && me.role === 'super_admin');
    document.querySelectorAll('[data-blink-nav="admin"]').forEach((el) => {
      el.hidden = !isSuper;
    });
    const isLogged = !!(me && me.user_id);
    document.querySelectorAll('[data-blink-nav="guest"]').forEach((el) => {
      el.hidden = isLogged;
    });
  }

  class BlinkNav extends HTMLElement {
    connectedCallback() {
      const active = this.getAttribute('active') || '';
      const nav = document.createElement('nav');
      nav.className = 'top-nav';

      const brand = document.createElement('a');
      brand.className = 'brand';
      brand.href = '/web/index.html';
      brand.textContent = 'Blink';
      nav.appendChild(brand);

      for (const link of LINKS) {
        const a = document.createElement('a');
        a.href = link.href;
        a.textContent = link.label;
        if (link.key === active) a.className = 'nav-active';
        if (link.adminOnly) {
          a.dataset.blinkNav = 'admin';
          a.hidden = true;
        } else if (link.guestOnly) {
          a.dataset.blinkNav = 'guest';
        }
        nav.appendChild(a);
      }

      this.replaceWith(nav);
      fetchMe().then(applyAdminVisibility);
    }
  }

  if (!customElements.get('blink-nav')) {
    customElements.define('blink-nav', BlinkNav);
  }
})();
