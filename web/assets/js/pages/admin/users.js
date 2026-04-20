/**
 * Admin — users.
 *
 * Server returns at most `limit` users without a total count; we keep the
 * filter client-side (fuzzy match over email / name / id). Actions all
 * go through BlinkModal instead of window.prompt/confirm.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;

  const PAGE_SIZE = 50;

  function fmtTime(s) {
    if (!s) return '—';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  function statusChip(u) {
    const s = typeof u.status === 'number' ? u.status : -1;
    const label = AdminAPI.userStatusLabel(s);
    const cls = s === 1 ? 'chip chip-ok' : (s === 2 ? 'chip chip-danger' : 'chip chip-warn');
    return el('span', { class: cls }, label);
  }

  function roleChip(u) {
    const r = u.role || 'user';
    const cls = r === 'super_admin' ? 'chip chip-info' : (r === 'admin' ? 'chip chip-info' : 'chip chip-muted');
    return el('span', { class: cls }, r);
  }

  async function mount(container, ctx) {
    const state = { all: [], filtered: [], query: '', limit: PAGE_SIZE };

    const errEl = el('p', { class: 'err', role: 'alert' });

    const searchInput = el('input', {
      type: 'search',
      placeholder: '按 邮箱 / 名称 / ID 搜索…',
      'aria-label': '搜索用户',
    });
    const limitSelect = el('select', { 'aria-label': '加载数量' }, [
      el('option', { value: '50' }, '加载 50'),
      el('option', { value: '100' }, '加载 100'),
      el('option', { value: '200' }, '加载 200'),
    ]);
    const reloadBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => load(),
    }, '刷新');

    const toolbar = el('div', { class: 'admin-toolbar' }, [
      searchInput,
      el('span', { class: 'admin-toolbar-spacer' }),
      limitSelect,
      reloadBtn,
    ]);

    const tbody = el('tbody');
    const pager = el('div', { class: 'admin-pager' });
    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', {}, 'ID'),
          el('th', {}, '邮箱'),
          el('th', {}, '名称'),
          el('th', {}, '角色'),
          el('th', {}, '状态'),
          el('th', {}, '最近登录'),
          el('th', { style: 'text-align:right' }, '操作'),
        ])),
        tbody,
      ]),
    ]);

    container.appendChild(errEl);
    container.appendChild(toolbar);
    container.appendChild(tableWrap);
    container.appendChild(pager);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }

    function render() {
      clear(tbody);
      const q = state.query.trim().toLowerCase();
      const items = q
        ? state.all.filter((u) =>
            String(u.id).toLowerCase().includes(q) ||
            (u.email || '').toLowerCase().includes(q) ||
            (u.name || '').toLowerCase().includes(q))
        : state.all.slice();
      state.filtered = items;

      if (items.length === 0) {
        tbody.appendChild(el('tr', {}, el('td', {
          colspan: '7',
          class: 'admin-empty',
        }, state.all.length ? '没有匹配的用户' : '暂无用户数据')));
      } else {
        items.forEach((u) => tbody.appendChild(renderRow(u)));
      }

      pager.textContent = '';
      pager.appendChild(el('span', {},
        state.all.length
          ? '已加载 ' + state.all.length + ' 条' + (q ? '，筛选后 ' + items.length + ' 条' : '')
          : ''));
    }

    function renderRow(u) {
      const tr = el('tr');
      tr.appendChild(el('td', { class: 'mono' }, String(u.id)));
      tr.appendChild(el('td', {}, u.email || '—'));
      tr.appendChild(el('td', {}, u.name || '—'));
      tr.appendChild(el('td', {}, roleChip(u)));
      tr.appendChild(el('td', {}, statusChip(u)));
      tr.appendChild(el('td', {}, fmtTime(u.last_login_at)));

      const actions = el('div', { class: 'row-actions' });
      if (u.status !== 1) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-secondary btn-sm',
          onClick: () => changeStatus(u, 1, '恢复为正常', '该用户将可以重新登录与使用站点。'),
        }, '恢复'));
      }
      if (u.status !== 2) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-secondary btn-sm',
          onClick: () => changeStatus(u, 2, '封禁此用户', '被封禁的用户无法登录。该操作记录日志。'),
        }, '封禁'));
      }
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-secondary btn-sm',
        onClick: () => editRole(u),
      }, '改角色'));
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-ghost btn-sm',
        onClick: () => resetPassword(u),
      }, '重置密码'));
      tr.appendChild(el('td', {}, actions));
      return tr;
    }

    async function changeStatus(u, newStatus, title, desc) {
      const danger = newStatus === 2;
      const ok = await Modal.confirm({
        title: title + ' #' + u.id,
        description: (desc || '') + '\n\n邮箱：' + (u.email || '—') + '\n名称：' + (u.name || '—'),
        danger,
        confirmLabel: danger ? '确认封禁' : '确认恢复',
      });
      if (!ok) return;
      try {
        showErr(null);
        await AdminAPI.patchUser(u.id, { status: newStatus });
        await load();
      } catch (err) { showErr(err); }
    }

    async function editRole(u) {
      const result = await Modal.open({
        title: '修改角色 #' + u.id,
        description: '当前角色：' + (u.role || 'user'),
        fields: [{
          name: 'role',
          label: '角色',
          type: 'radio',
          value: u.role || 'user',
          options: [
            { value: 'user',        label: 'user',        hint: '普通用户' },
            { value: 'admin',       label: 'admin',       hint: '内容审核权限（规划中）' },
            { value: 'super_admin', label: 'super_admin', hint: '全部后台权限，谨慎授予' },
          ],
        }],
        confirmLabel: '保存',
        danger: u.role !== 'super_admin',
      });
      if (!result) return;
      if (result.role === u.role) return;
      try {
        showErr(null);
        await AdminAPI.patchUser(u.id, { role: result.role });
        await load();
      } catch (err) { showErr(err); }
    }

    async function resetPassword(u) {
      const result = await Modal.open({
        title: '重置密码 #' + u.id,
        description: '为 ' + (u.email || u.name || '此用户') + ' 设置新密码（至少 8 位，仅 builtin 登录方式生效）。',
        fields: [{
          name: 'password',
          label: '新密码',
          type: 'password',
          required: true,
          minLength: 8,
          placeholder: '至少 8 位',
          hint: '保存后请通过安全渠道告知用户。',
        }],
        confirmLabel: '重置',
        danger: true,
      });
      if (!result) return;
      try {
        showErr(null);
        await AdminAPI.resetUserPassword(u.id, result.password);
        await Modal.alert({
          title: '已重置',
          description: '用户 ' + (u.email || u.name || '#' + u.id) + ' 的密码已更新。',
        });
      } catch (err) { showErr(err); }
    }

    async function load() {
      showErr(null);
      state.limit = Number(limitSelect.value) || PAGE_SIZE;
      reloadBtn.disabled = true;
      try {
        const d = await AdminAPI.listUsers({ limit: state.limit });
        state.all = d.users || [];
        render();
      } catch (err) {
        showErr(err);
      } finally {
        reloadBtn.disabled = false;
      }
    }

    searchInput.addEventListener('input', () => {
      state.query = searchInput.value;
      render();
    });
    limitSelect.addEventListener('change', () => load());

    await load();

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.users = { mount };
})();
