/**
 * Admin console bootstrap + router.
 *
 * Modules self-register on `window.BlinkAdminModules.<id> = { mount, unmount }`.
 * Each `mount(container, ctx)` is async; `ctx` exposes:
 *   - params: URLSearchParams parsed from "#route?key=val"
 *   - navigate(hash)     switch route (accepts "users" or "#users?foo=1")
 *   - refreshBadges()    rerun sidebar badge probes
 *
 * Route hash grammar: "#<id>[?<key>=<val>&…]"
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;

  const ROUTES = [
    { id: 'dashboard', label: '概览' },
    { id: 'users',     label: '用户' },
    { id: 'posts',     label: '帖子' },
    { id: 'appeals',   label: '申诉', badgeKey: 'appeals' },
    { id: 'replies',   label: '评论' },
    { id: 'sensitive', label: '敏感词' },
    { id: 'settings',  label: '设置' },
  ];

  const SUBTITLES = {
    dashboard: '整体数据速览与待处理事项。',
    users:     '搜索用户、调整角色与账号状态。',
    posts:     '筛选帖子并处理审核标记。',
    appeals:   '审阅作者提交的申诉与复核。',
    replies:   '按帖子或评论 ID 管理评论可见性。',
    sensitive: '敏感词词表，修改实时生效。',
    settings:  '后台行为开关。',
  };

  const navLinks = {};
  let currentModule = null;

  function parseHash() {
    const raw = (window.location.hash || '').replace(/^#\/?/, '');
    const qIdx = raw.indexOf('?');
    const path = qIdx >= 0 ? raw.slice(0, qIdx) : raw;
    const query = qIdx >= 0 ? raw.slice(qIdx + 1) : '';
    const route = path || 'dashboard';
    return { route, params: new URLSearchParams(query) };
  }

  function formatQuery(params) {
    if (!params) return '';
    if (params instanceof URLSearchParams) {
      const s = params.toString();
      return s ? '?' + s : '';
    }
    const keys = Object.keys(params);
    if (!keys.length) return '';
    const p = new URLSearchParams();
    keys.forEach((k) => {
      if (params[k] != null && params[k] !== '') p.set(k, String(params[k]));
    });
    const s = p.toString();
    return s ? '?' + s : '';
  }

  function navigateTo(hashOrId, params) {
    let target;
    if (typeof hashOrId === 'string' && hashOrId.startsWith('#')) {
      target = hashOrId;
    } else {
      target = '#' + String(hashOrId || 'dashboard') + formatQuery(params);
    }
    if (target === window.location.hash) {
      navigate();
    } else {
      window.location.hash = target;
    }
  }

  function renderSidebar(parent) {
    const list = el('nav', { class: 'admin-nav', 'aria-label': '管理后台导航' });
    ROUTES.forEach((r) => {
      const a = el('a', { href: '#' + r.id, 'data-route': r.id }, r.label);
      if (r.badgeKey) {
        const b = el('span', { class: 'admin-nav-badge', 'data-badge': r.badgeKey }, '0');
        b.hidden = true;
        a.appendChild(b);
      }
      navLinks[r.id] = a;
      list.appendChild(a);
    });
    parent.appendChild(list);
  }

  function setActive(route) {
    Object.keys(navLinks).forEach((k) => {
      navLinks[k].classList.toggle('is-active', k === route);
    });
  }

  function renderTopbar(route) {
    const meta = ROUTES.find((r) => r.id === route) || ROUTES[0];
    const top = el('div', { class: 'admin-topbar' }, [
      el('div', {}, [
        el('h1', { class: 'admin-title' }, meta.label),
        el('p', { class: 'admin-subtitle' }, SUBTITLES[meta.id] || ''),
      ]),
    ]);
    return top;
  }

  async function navigate() {
    const { route, params } = parseHash();
    const target = ROUTES.find((r) => r.id === route) ? route : 'dashboard';
    setActive(target);

    const view = document.getElementById('admin-view');
    if (currentModule && typeof currentModule.unmount === 'function') {
      try { currentModule.unmount(); } catch (err) { console.error(err); }
    }
    currentModule = null;
    clear(view);

    view.appendChild(renderTopbar(target));
    const body = el('div', { class: 'admin-view-body' });
    view.appendChild(body);

    const registry = window.BlinkAdminModules || {};
    const mod = registry[target];
    if (!mod || typeof mod.mount !== 'function') {
      body.appendChild(el('div', { class: 'card' }, '模块未加载：' + target));
      return;
    }

    const ctx = {
      params,
      navigate: navigateTo,
      refreshBadges,
    };
    try {
      const result = await mod.mount(body, ctx);
      currentModule = result || mod;
    } catch (err) {
      body.appendChild(el('p', { class: 'err' }, errorText(err)));
    }
  }

  async function refreshBadges() {
    try {
      const d = await window.BlinkAdminAPI.listPosts({ appeal_pending: 1, limit: 1, offset: 0 });
      const count = typeof d.total === 'number' ? d.total : (d.posts || []).length;
      updateBadge('appeals', count);
    } catch (_) {
      updateBadge('appeals', 0);
    }
  }

  function updateBadge(key, count) {
    document.querySelectorAll('[data-badge="' + key + '"]').forEach((node) => {
      if (!count) { node.hidden = true; return; }
      node.hidden = false;
      node.textContent = String(count);
    });
  }

  async function start() {
    await window.BlinkAdmin.gate({
      gateId: 'admin-gate',
      appId: 'admin-app',
      leadId: 'admin-gate-lead',
      onReady: async () => {
        renderSidebar(document.getElementById('admin-sidebar'));
        window.addEventListener('hashchange', navigate);
        await navigate();
        await refreshBadges();
      },
    });
  }

  start();
})();
