/**
 * Admin — replies.
 *
 * Two entry points:
 *   - by post_id: loads the post's replies (cursor-paginated, incl. hidden)
 *     and exposes toggle-hide cascade actions per row.
 *   - by reply_id: one-shot cascade hide from a quick form.
 *
 * Supports deep link via `#replies?post_id=…`.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;

  const PAGE_SIZE = 50;

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

  function statusChip(r) {
    const s = typeof r.status === 'number' ? r.status : 0;
    const cls = s === 1 ? 'chip chip-danger' : 'chip chip-ok';
    return el('span', { class: cls }, AdminAPI.replyStatusLabel(s));
  }

  async function mount(container, ctx) {
    const initPostId = (ctx.params && ctx.params.get('post_id')) || '';
    const state = { postId: '', cursor: null, replies: [] };

    const errEl = el('p', { class: 'err', role: 'alert' });

    // --- search card ---
    const postIdInput = el('input', {
      type: 'text',
      placeholder: 'post_id，例如 1234567890',
      value: initPostId,
      'aria-label': 'post_id',
    });
    const loadBtn = el('button', {
      type: 'button', class: 'btn btn-primary btn-sm',
      onClick: () => loadFirst(),
    }, '加载评论');
    const moreBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => loadMore(),
    }, '加载更多');
    moreBtn.hidden = true;

    const searchCard = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '按帖子加载评论'),
      el('p', { class: 'admin-subtitle' }, '展示包含隐藏评论的完整列表。'),
      el('div', { class: 'admin-form-row' }, [postIdInput, loadBtn, moreBtn]),
    ]);

    // --- quick hide by reply id ---
    const quickCard = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '按评论 ID 快速隐藏'),
      el('p', { class: 'admin-subtitle' }, '直接隐藏评论及其所有子评论，无需先加载帖子。'),
    ]);
    const quickInput = el('input', { type: 'text', placeholder: 'reply_id' });
    const quickBtn = el('button', {
      type: 'button', class: 'btn btn-danger btn-sm',
      onClick: () => quickHide(),
    }, '隐藏（级联）');
    quickCard.appendChild(el('div', { class: 'admin-form-row' }, [quickInput, quickBtn]));

    // --- table ---
    const tbody = el('tbody');
    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', { class: 'nowrap' }, 'ID'),
          el('th', {}, '作者'),
          el('th', { class: 'nowrap' }, '父评论'),
          el('th', { class: 'nowrap' }, '状态'),
          el('th', {}, '内容'),
          el('th', { class: 'nowrap' }, '时间'),
          el('th', { class: 'col-actions' }, '操作'),
        ])),
        tbody,
      ]),
    ]);

    const pageMeta = el('p', { class: 'admin-subtitle' });

    container.appendChild(errEl);
    container.appendChild(searchCard);
    container.appendChild(quickCard);
    container.appendChild(pageMeta);
    container.appendChild(tableWrap);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }

    function render() {
      clear(tbody);
      if (state.replies.length === 0) {
        tbody.appendChild(el('tr', {}, el('td', {
          colspan: '7', class: 'admin-empty',
        }, state.postId ? '当前帖子没有评论' : '输入 post_id 并加载以开始')));
      } else {
        state.replies.forEach((r) => tbody.appendChild(renderRow(r)));
      }
      pageMeta.textContent = state.postId
        ? '帖子 #' + state.postId + ' · 已加载 ' + state.replies.length + ' 条' + (state.cursor ? '，还有更多' : '')
        : '';
    }

    function renderRow(r) {
      const tr = el('tr');
      tr.appendChild(el('td', { class: 'mono nowrap' }, String(r.id)));
      tr.appendChild(el('td', {}, [
        el('div', {}, r.user_name || '—'),
        el('div', { class: 'mono' }, String(r.user_id || '—')),
      ]));
      tr.appendChild(el('td', { class: 'mono nowrap' }, r.parent_reply_id != null ? String(r.parent_reply_id) : '—'));
      tr.appendChild(el('td', { class: 'nowrap' }, statusChip(r)));
      tr.appendChild(el('td', { class: 'cell-ellipsis' }, truncate(r.body || '', 80) || '—'));
      tr.appendChild(el('td', { class: 'nowrap' }, fmtTime(r.created_at)));

      const actions = el('div', { class: 'row-actions' });
      if (r.status === 1) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-secondary btn-sm',
          onClick: () => toggleHidden(r, false),
        }, '恢复（级联）'));
      } else {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-danger btn-sm',
          onClick: () => toggleHidden(r, true),
        }, '隐藏（级联）'));
      }
      tr.appendChild(el('td', { class: 'col-actions' }, actions));
      return tr;
    }

    async function toggleHidden(r, hide) {
      const ok = await Modal.confirm({
        title: (hide ? '隐藏' : '恢复') + '评论 #' + r.id,
        description: (hide
          ? '评论及其所有子评论将被设为隐藏。'
          : '评论及其所有子评论将被设为可见。')
          + '\n\n内容摘要：' + truncate(r.body || '', 80),
        danger: hide,
        confirmLabel: hide ? '隐藏' : '恢复',
      });
      if (!ok) return;
      try {
        showErr(null);
        await AdminAPI.patchReply(r.id, { hidden: hide });
        await loadFirst();
      } catch (err) { showErr(err); }
    }

    async function loadFirst() {
      showErr(null);
      const pid = String(postIdInput.value || '').trim();
      if (!pid) return showErr('请填写 post_id');
      state.postId = pid;
      state.cursor = null;
      state.replies = [];
      try {
        const d = await AdminAPI.listPostReplies(pid, { limit: PAGE_SIZE });
        state.replies = d.replies || [];
        state.cursor = d.next_cursor || null;
        moreBtn.hidden = !state.cursor;
        render();
      } catch (err) {
        showErr(err);
        render();
      }
    }

    async function loadMore() {
      if (!state.postId || !state.cursor) return;
      showErr(null);
      try {
        const d = await AdminAPI.listPostReplies(state.postId, { limit: PAGE_SIZE, cursor: state.cursor });
        state.replies = state.replies.concat(d.replies || []);
        state.cursor = d.next_cursor || null;
        moreBtn.hidden = !state.cursor;
        render();
      } catch (err) { showErr(err); }
    }

    async function quickHide() {
      showErr(null);
      const rid = String(quickInput.value || '').trim();
      if (!rid) return showErr('请填写 reply_id');
      const ok = await Modal.confirm({
        title: '隐藏评论 #' + rid,
        description: '该评论及其所有子评论将被设为隐藏。该操作会记录日志。',
        danger: true,
        confirmLabel: '确认隐藏',
      });
      if (!ok) return;
      try {
        await AdminAPI.hideReplyCascade(rid);
        quickInput.value = '';
        if (state.postId) await loadFirst();
      } catch (err) { showErr(err); }
    }

    postIdInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') { e.preventDefault(); loadFirst(); }
    });
    quickInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') { e.preventDefault(); quickHide(); }
    });

    render();
    if (initPostId) await loadFirst();

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.replies = { mount };
})();
