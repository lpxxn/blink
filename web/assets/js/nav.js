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
    { key: 'trending', href: '/web/trending.html', label: '热门' },
    { key: 'compose',  href: '/web/compose.html',  label: '发帖', authOnly: true },
    { key: 'mine',     href: '/web/mine.html',     label: '我的', authOnly: true },
    { key: 'messages', href: '/web/messages.html', label: '消息', authOnly: true },
    { key: 'feedback', href: '/web/feedback.html', label: '反馈' },
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
    const isAdmin = !!(me && (me.role === 'super_admin' || me.role === 'admin'));
    document.querySelectorAll('[data-blink-nav="admin"]').forEach((el) => {
      el.hidden = !isAdmin;
    });
    const isLogged = !!(me && me.user_id);
    document.querySelectorAll('[data-blink-nav="guest"]').forEach((el) => {
      el.hidden = isLogged;
    });
    document.querySelectorAll('[data-blink-nav="auth"]').forEach((el) => {
      el.hidden = !isLogged;
    });
    if (isLogged) {
      startSSE();
      fetchUnreadCount();
    }
  }

  function fetchUnreadCount() {
    var p = window.BlinkAPI
      ? window.BlinkAPI.get('/api/me/notifications/unread_count')
      : fetch('/api/me/notifications/unread_count', { credentials: 'include' })
          .then(function (r) { return r.ok ? r.json() : null; })
          .catch(function () { return null; });
    p.then(function (d) {
      if (d && d.unread_count != null) {
        updateBadge(Number(d.unread_count));
      }
    }).catch(function () {});
  }

  function updateBadge(count) {
    document.querySelectorAll('[data-blink-unread]').forEach(function (badge) {
      if (count > 0) {
        badge.textContent = count > 99 ? '99+' : String(count);
        badge.hidden = false;
      } else {
        badge.hidden = true;
      }
    });
  }

  var sseSource = null;
  function startSSE() {
    if (sseSource) return;
    if (typeof EventSource === 'undefined') return;
    sseSource = new EventSource('/api/me/notifications/stream', { withCredentials: true });
    sseSource.addEventListener('notification', function (e) {
      try {
        var data = JSON.parse(e.data);
        var count = Number(data.unread_count);
        if (!isNaN(count)) updateBadge(count);
      } catch (_) {}
    });
    sseSource.onerror = function () {
      sseSource.close();
      sseSource = null;
      setTimeout(function () {
        fetchMe().then(function (me) {
          if (me && me.user_id) startSSE();
        });
      }, 5000);
    };
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
        } else if (link.authOnly) {
          a.dataset.blinkNav = 'auth';
          a.hidden = true;
        } else if (link.guestOnly) {
          a.dataset.blinkNav = 'guest';
        }
        if (link.key === 'messages') {
          a.style.position = 'relative';
          const badge = document.createElement('span');
          badge.className = 'nav-unread-badge';
          badge.dataset.blinkUnread = '1';
          badge.hidden = true;
          a.appendChild(badge);
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
