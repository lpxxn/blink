/**
 * Admin — posts.
 *
 * Server-side pagination via offset/limit. Moderation filter is a tab row.
 * Row can be expanded to preview body + moderation metadata without leaving
 * the page. Actions use BlinkModal.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;

  const LIMIT = 20;

  const FILTERS = [
    { id: 'all',     label: '全部', params: {} },
    { id: 'flagged', label: '违规', params: { moderation_flag: 1 } },
    { id: 'removed', label: '已下架', params: { moderation_flag: 2 } },
    { id: 'normal',  label: '正常', params: { moderation_flag: 0 } },
  ];

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  function truncate(s, n) {
    s = s == null ? '' : String(s).replace(/\s+/g, ' ').trim();
    return s.length > n ? s.slice(0, n) + '…' : s;
  }

  function modChip(p) {
    const m = typeof p.moderation_flag === 'number' ? p.moderation_flag : 0;
    const cls = m === 0 ? 'chip chip-ok' : (m === 1 ? 'chip chip-warn' : 'chip chip-danger');
    return el('span', { class: cls }, AdminAPI.modLabel(m));
  }

  function statusChip(p) {
    const s = typeof p.status === 'number' ? p.status : 0;
    const cls = s === 1 ? 'chip chip-info' : (s === 2 ? 'chip chip-muted' : 'chip chip-muted');
    return el('span', { class: cls }, AdminAPI.postStatusLabel(s));
  }

  function appealChip(p) {
    if (p.appeal_status === 1) return el('span', { class: 'chip chip-warn' }, '申诉中');
    return null;
  }

  async function mount(container, ctx) {
    const params = ctx.params || new URLSearchParams();
    let activeFilter = 'all';
    if (params.get('moderation_flag') === '1') activeFilter = 'flagged';
    else if (params.get('moderation_flag') === '2') activeFilter = 'removed';
    else if (params.get('moderation_flag') === '0') activeFilter = 'normal';

    const state = {
      offset: 0,
      total: 0,
      posts: [],
      expanded: new Set(),
    };

    const errEl = el('p', { class: 'err', role: 'alert' });

    const tabs = el('div', { class: 'admin-toolbar', role: 'tablist' });
    FILTERS.forEach((f) => {
      const btn = el('button', {
        type: 'button',
        class: 'btn btn-secondary btn-sm',
        role: 'tab',
        'data-filter': f.id,
        onClick: () => { activeFilter = f.id; state.offset = 0; updateTabs(); load(); },
      }, f.label);
      tabs.appendChild(btn);
    });
    tabs.appendChild(el('span', { class: 'admin-toolbar-spacer' }));
    const searchInput = el('input', {
      type: 'search',
      placeholder: '在当前页内按 ID / 用户 / 正文搜索…',
      'aria-label': '页内搜索',
    });
    tabs.appendChild(searchInput);

    function updateTabs() {
      tabs.querySelectorAll('[data-filter]').forEach((b) => {
        b.classList.toggle('btn-primary', b.dataset.filter === activeFilter);
        b.classList.toggle('btn-secondary', b.dataset.filter !== activeFilter);
      });
    }
    updateTabs();

    const tbody = el('tbody');
    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', { class: 'nowrap' }, 'ID'),
          el('th', {}, '作者'),
          el('th', { class: 'nowrap' }, '审核'),
          el('th', { class: 'nowrap' }, '状态'),
          el('th', {}, '摘要'),
          el('th', { class: 'nowrap' }, '创建'),
          el('th', { class: 'col-actions' }, '操作'),
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
    container.appendChild(tabs);
    container.appendChild(tableWrap);
    container.appendChild(pager);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }

    function toggleExpand(id) {
      const key = String(id);
      if (state.expanded.has(key)) state.expanded.delete(key);
      else state.expanded.add(key);
      render();
    }

    function render() {
      clear(tbody);
      const q = (searchInput.value || '').trim().toLowerCase();
      const list = q
        ? state.posts.filter((p) =>
            String(p.id).includes(q) ||
            (p.user_name || '').toLowerCase().includes(q) ||
            String(p.user_id).includes(q) ||
            (p.body || '').toLowerCase().includes(q))
        : state.posts;

      if (list.length === 0) {
        tbody.appendChild(el('tr', {}, el('td', {
          colspan: '7', class: 'admin-empty',
        }, state.posts.length ? '没有匹配的帖子' : '暂无帖子')));
      } else {
        list.forEach((p) => {
          tbody.appendChild(renderRow(p));
          if (state.expanded.has(String(p.id))) {
            tbody.appendChild(renderExpand(p));
          }
        });
      }

      const end = Math.min(state.offset + LIMIT, state.total);
      pageMeta.textContent = state.total
        ? '共 ' + state.total + ' 条 · 显示 ' + (state.offset + 1) + '–' + end
        : '暂无数据';
      prevBtn.disabled = state.offset <= 0;
      nextBtn.disabled = state.offset + LIMIT >= state.total;
    }

    function renderRow(p) {
      const expanded = state.expanded.has(String(p.id));
      const tr = el('tr');
      tr.appendChild(el('td', { class: 'mono nowrap' }, String(p.id)));

      const author = el('td', {}, [
        el('div', {}, p.user_name && p.user_name.trim() ? p.user_name : '—'),
        el('div', { class: 'mono' }, String(p.user_id)),
      ]);
      tr.appendChild(author);

      const modTd = el('td', { class: 'nowrap' }, [modChip(p), appealChip(p)].filter(Boolean));
      tr.appendChild(modTd);
      tr.appendChild(el('td', { class: 'nowrap' }, statusChip(p)));

      const summary = el('td', { class: 'cell-ellipsis' },
        truncate(p.body || '', 100) || '—');
      tr.appendChild(summary);
      tr.appendChild(el('td', { class: 'nowrap' }, fmtTime(p.created_at)));

      const actions = el('div', { class: 'row-actions' });
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-ghost btn-sm',
        onClick: () => toggleExpand(p.id),
      }, expanded ? '收起' : '展开'));

      const mf = p.moderation_flag || 0;
      if (mf !== 2) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-danger btn-sm',
          onClick: () => takeDown(p),
        }, '下架'));
      }
      if (mf === 0) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-secondary btn-sm',
          onClick: () => flag(p),
        }, '标违规'));
      }
      if (mf !== 0 || p.status !== 1) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-primary btn-sm',
          onClick: () => restore(p),
        }, '恢复'));
      }
      tr.appendChild(el('td', { class: 'col-actions' }, actions));
      return tr;
    }

    function renderExpand(p) {
      const grid = el('div', { class: 'admin-expand-grid' }, [
        el('div', {}, [
          el('p', { class: 'admin-expand-title' }, '正文预览'),
          el('pre', { class: 'admin-expand-body' }, p.body || '（空）'),
          el('div', { class: 'admin-expand-shortcuts' }, [
            el('a', {
              class: 'btn btn-ghost btn-sm',
              href: '/web/post.html?id=' + encodeURIComponent(String(p.id)),
              target: '_blank', rel: 'noopener',
            }, '在新窗口查看'),
            el('button', {
              type: 'button', class: 'btn btn-ghost btn-sm',
              onClick: () => ctx.navigate('replies', { post_id: p.id }),
            }, '管理评论'),
          ]),
        ]),
        el('dl', { class: 'admin-expand-meta' }, [
          el('dt', {}, '更新时间'),       el('dd', {}, fmtTime(p.updated_at) || '—'),
          el('dt', {}, '审核备注'),       el('dd', {}, p.moderation_note || '—'),
          el('dt', {}, '申诉正文'),       el('dd', {}, p.appeal_body || '—'),
          el('dt', {}, '图片'),           el('dd', {}, String((p.images || []).length) + ' 张'),
        ]),
      ]);
      return el('tr', { class: 'row-expand' }, el('td', { colspan: '7' }, grid));
    }

    async function takeDown(p) {
      const result = await Modal.open({
        title: '下架帖子 #' + p.id,
        description: '作者会收到通知，内容将从公开流中移除。请写明下架理由。',
        fields: [{
          name: 'note',
          label: '下架理由',
          type: 'textarea',
          required: true,
          maxLength: 500,
          placeholder: '例如：内容含违规信息…',
        }],
        confirmLabel: '确认下架',
        danger: true,
      });
      if (!result) return;
      try {
        showErr(null);
        await AdminAPI.patchPost(p.id, { moderation_flag: 2, moderation_note: result.note });
        await load();
        ctx.refreshBadges();
      } catch (err) { showErr(err); }
    }

    async function flag(p) {
      const result = await Modal.open({
        title: '标记违规 #' + p.id,
        description: '仅做标记，不影响公开可见性。可附说明以便后续处理。',
        fields: [{
          name: 'note',
          label: '说明（可选）',
          type: 'textarea',
          maxLength: 500,
          placeholder: '例如：疑似广告，需进一步审查…',
        }],
        confirmLabel: '标记违规',
      });
      if (!result) return;
      try {
        showErr(null);
        await AdminAPI.patchPost(p.id, {
          moderation_flag: 1,
          moderation_note: result.note || undefined,
        });
        await load();
      } catch (err) { showErr(err); }
    }

    async function restore(p) {
      const ok = await Modal.confirm({
        title: '恢复公开 #' + p.id,
        description: '将帖子恢复为审核正常 + 已发布状态。作者可立即看到。',
      });
      if (!ok) return;
      try {
        showErr(null);
        await AdminAPI.patchPost(p.id, { moderation_flag: 0, status: 1 });
        await load();
        ctx.refreshBadges();
      } catch (err) { showErr(err); }
    }

    async function load() {
      showErr(null);
      const filter = FILTERS.find((f) => f.id === activeFilter) || FILTERS[0];
      const q = Object.assign({ limit: LIMIT, offset: state.offset }, filter.params);
      try {
        const d = await AdminAPI.listPosts(q);
        state.posts = d.posts || [];
        state.total = typeof d.total === 'number' ? d.total : state.posts.length;
        state.expanded.clear();
        render();
      } catch (err) { showErr(err); }
    }

    searchInput.addEventListener('input', render);

    await load();

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.posts = { mount };
})();
