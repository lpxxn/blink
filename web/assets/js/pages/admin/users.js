/**
 * Admin — users (server-side search + pagination).
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
    const state = { users: [], total: 0, offset: 0, q: '', limit: PAGE_SIZE };

    const errEl = el('p', { class: 'err', role: 'alert' });
    const searchInput = el('input', {
      type: 'search',
      placeholder: '按 邮箱 / 名称 / ID 搜索…',
      'aria-label': '搜索用户',
    });
    const reloadBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { state.offset = 0; load(); },
    }, '刷新');

    const toolbar = el('div', { class: 'admin-toolbar' }, [
      searchInput,
      el('span', { class: 'admin-toolbar-spacer' }),
      reloadBtn,
    ]);

    const tbody = el('tbody');
    const pager = el('div', { class: 'admin-pager' });
    const pageMeta = el('span', {});
    const prevBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { if (state.offset > 0) { state.offset = Math.max(0, state.offset - state.limit); load(); } },
    }, '上一页');
    const nextBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { if (state.offset + state.limit < state.total) { state.offset += state.limit; load(); } },
    }, '下一页');
    pager.appendChild(pageMeta);
    pager.appendChild(el('span', { class: 'admin-pager-spacer' }));
    pager.appendChild(prevBtn);
    pager.appendChild(nextBtn);

    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', { class: 'nowrap' }, 'ID'),
          el('th', {}, '邮箱'),
          el('th', {}, '名称'),
          el('th', { class: 'nowrap' }, '角色'),
          el('th', { class: 'nowrap' }, '状态'),
          el('th', { class: 'nowrap' }, '最近登录'),
          el('th', { class: 'col-actions' }, '操作'),
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
      if (!state.users.length) {
        tbody.appendChild(el('tr', {}, el('td', {
          colspan: '7',
          class: 'admin-empty',
        }, state.q ? '没有匹配的用户' : '暂无用户数据')));
      } else {
        state.users.forEach((u) => tbody.appendChild(renderRow(u)));
      }
      const end = Math.min(state.offset + state.limit, state.total);
      pageMeta.textContent = state.total
        ? '共 ' + state.total + ' 条 · 显示 ' + (state.offset + 1) + '–' + end
        : '暂无数据';
      prevBtn.disabled = state.offset <= 0;
      nextBtn.disabled = state.offset + state.limit >= state.total;
    }

    function renderRow(u) {
      const tr = el('tr');
      tr.appendChild(el('td', { class: 'mono nowrap' }, String(u.id)));
      tr.appendChild(el('td', {}, u.email || '—'));
      tr.appendChild(el('td', {}, u.name || '—'));
      tr.appendChild(el('td', { class: 'nowrap' }, roleChip(u)));
      tr.appendChild(el('td', { class: 'nowrap' }, statusChip(u)));
      tr.appendChild(el('td', { class: 'nowrap' }, fmtTime(u.last_login_at)));

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
      tr.appendChild(el('td', { class: 'col-actions' }, actions));
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
            { value: 'admin',       label: 'admin',       hint: '后台管理权限' },
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

    let searchTimer = null;
    searchInput.addEventListener('input', () => {
      clearTimeout(searchTimer);
      searchTimer = setTimeout(() => {
        state.q = (searchInput.value || '').trim();
        state.offset = 0;
        load();
      }, 300);
    });

    async function load() {
      showErr(null);
      reloadBtn.disabled = true;
      try {
        const params = { limit: state.limit, offset: state.offset };
        if (state.q) params.q = state.q;
        const d = await AdminAPI.listUsers(params);
        state.users = d.users || [];
        state.total = typeof d.total === 'number' ? d.total : (Number(d.total) || state.users.length);
        render();
      } catch (err) {
        showErr(err);
      } finally {
        reloadBtn.disabled = false;
      }
    }

    await load();
    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.users = { mount };
})();
