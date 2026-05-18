/**
 * Admin — audit logs.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;

  const LIMIT = 50;

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  async function mount(container) {
    const state = { offset: 0, total: 0, logs: [] };
    const errEl = el('p', { class: 'err', role: 'alert' });
    const reloadBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm' }, '刷新');
    const toolbar = el('div', { class: 'admin-toolbar' }, [
      el('span', { class: 'admin-subtitle' }, '管理员操作记录'),
      el('span', { class: 'admin-toolbar-spacer' }),
      reloadBtn,
    ]);

    const tbody = el('tbody');
    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', { class: 'nowrap' }, '时间'),
          el('th', { class: 'nowrap' }, '操作者'),
          el('th', { class: 'nowrap' }, '动作'),
          el('th', { class: 'nowrap' }, '目标'),
          el('th', {}, '详情'),
        ])),
        tbody,
      ]),
    ]);

    const pager = el('div', { class: 'admin-pager' });
    const pageMeta = el('span', {});
    const prevBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { if (state.offset > 0) { state.offset = Math.max(0, state.offset - LIMIT); load(); } },
    }, '上一页');
    const nextBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { if (state.offset + LIMIT < state.total) { state.offset += LIMIT; load(); } },
    }, '下一页');
    pager.appendChild(pageMeta);
    pager.appendChild(el('span', { class: 'admin-pager-spacer' }));
    pager.appendChild(prevBtn);
    pager.appendChild(nextBtn);

    container.appendChild(errEl);
    container.appendChild(toolbar);
    container.appendChild(tableWrap);
    container.appendChild(pager);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }

    function render() {
      clear(tbody);
      if (!state.logs.length) {
        tbody.appendChild(el('tr', {}, el('td', { colspan: '5', class: 'admin-empty' }, '暂无审计记录')));
        return;
      }
      state.logs.forEach((row) => {
        const tr = el('tr');
        tr.appendChild(el('td', { class: 'nowrap' }, fmtTime(row.created_at)));
        tr.appendChild(el('td', { class: 'mono nowrap' }, String(row.actor_id || '—')));
        tr.appendChild(el('td', { class: 'nowrap' }, row.action || '—'));
        const target = (row.target_type || '') +
          (row.target_id != null ? ' #' + row.target_id : '');
        tr.appendChild(el('td', { class: 'mono' }, target || '—'));
        tr.appendChild(el('td', { class: 'cell-ellipsis' }, row.detail || '—'));
        tbody.appendChild(tr);
      });
      const end = Math.min(state.offset + LIMIT, state.total);
      pageMeta.textContent = state.total
        ? '共 ' + state.total + ' 条 · 显示 ' + (state.offset + 1) + '–' + end
        : '暂无数据';
      prevBtn.disabled = state.offset <= 0;
      nextBtn.disabled = state.offset + LIMIT >= state.total;
    }

    async function load() {
      showErr(null);
      reloadBtn.disabled = true;
      try {
        const d = await AdminAPI.listAuditLogs({ limit: LIMIT, offset: state.offset });
        state.logs = d.logs || [];
        state.total = typeof d.total === 'number' ? d.total : Number(d.total) || state.logs.length;
        render();
      } catch (err) {
        showErr(err);
      } finally {
        reloadBtn.disabled = false;
      }
    }

    reloadBtn.addEventListener('click', load);
    await load();
    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.audit = { mount };
})();
