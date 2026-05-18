/**
 * Admin — user feedback (filter, pagination, close).
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;

  const LIMIT = 20;

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  function renderMessage(m) {
    const isAdmin = m.sender_type === 'admin';
    return el('div', { class: 'reply' + (isAdmin ? ' reply-nested' : '') }, [
      el('div', { class: 'meta' }, (isAdmin ? '管理员' : '用户') + ' · ' + fmtTime(m.created_at)),
      el('div', { style: { whiteSpace: 'pre-wrap', marginTop: '0.25rem' } }, m.body || ''),
    ]);
  }

  async function mount(container, ctx) {
    const params = ctx.params || new URLSearchParams();
    const state = {
      offset: 0,
      limit: LIMIT,
      total: 0,
      status: params.get('status') || '',
      userId: params.get('user_id') || '',
    };

    const errEl = el('p', { class: 'err', role: 'alert' });
    const statusFilter = el('select', { 'aria-label': '状态筛选' }, [
      el('option', { value: '' }, '全部状态'),
      el('option', { value: 'open' }, '进行中'),
      el('option', { value: 'closed' }, '已关闭'),
    ]);
    statusFilter.value = state.status;
    const userIdInput = el('input', {
      type: 'search',
      placeholder: '用户 ID（可选）',
      'aria-label': '用户 ID',
      value: state.userId,
    });
    const reloadBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm' }, '刷新');
    const toolbar = el('div', { class: 'admin-toolbar' }, [
      statusFilter,
      userIdInput,
      el('span', { class: 'admin-toolbar-spacer' }),
      reloadBtn,
    ]);

    const listEl = el('div', { class: 'admin-cards' });
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
    container.appendChild(listEl);
    container.appendChild(pager);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }

    function renderList(items) {
      clear(listEl);
      if (!items.length) {
        listEl.appendChild(el('div', { class: 'admin-empty' }, '当前没有意见反馈'));
        return;
      }
      items.forEach((thread) => listEl.appendChild(renderThread(thread)));
    }

    function renderThread(thread) {
      const card = el('div', { class: 'admin-card' });
      const detail = el('div', { class: 'admin-card-section', hidden: true });
      const openBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm' }, '查看/回复');
      const isClosed = thread.status === 'closed';
      const title = '反馈 #' + thread.id;
      const user = ((thread.user_name || '') + ' ' + thread.user_id).trim();

      card.appendChild(el('div', { class: 'admin-card-head' }, [
        el('div', { class: 'admin-card-head-main' }, [
          el('strong', {}, title),
          el('span', { class: 'admin-subtitle' }, user),
        ]),
        isClosed
          ? el('span', { class: 'chip chip-muted' }, '已关闭')
          : el('span', { class: 'chip chip-warn' }, '进行中'),
      ]));
      card.appendChild(el('p', { class: 'admin-subtitle' },
        '用户补充 ' + (thread.user_reply_count || 0) + '/2' +
        ' · 最后更新 ' + fmtTime(thread.last_message_at)
      ));
      const actions = el('div', { class: 'admin-card-actions' }, [openBtn]);
      if (!isClosed) {
        actions.appendChild(el('button', {
          type: 'button', class: 'btn btn-ghost btn-sm',
          onClick: () => closeThread(thread),
        }, '关闭工单'));
      }
      card.appendChild(actions);
      card.appendChild(detail);

      openBtn.addEventListener('click', async () => {
        if (!detail.hidden) {
          detail.hidden = true;
          openBtn.textContent = '查看/回复';
          return;
        }
        openBtn.disabled = true;
        try {
          await loadDetail(thread.id, detail, isClosed);
          detail.hidden = false;
          openBtn.textContent = '收起';
        } catch (err) {
          showErr(err);
        } finally {
          openBtn.disabled = false;
        }
      });
      return card;
    }

    async function closeThread(thread) {
      const ok = await Modal.confirm({
        title: '关闭反馈 #' + thread.id,
        description: '关闭后用户将无法再补充，但仍可查看历史记录。',
        confirmLabel: '关闭工单',
      });
      if (!ok) return;
      try {
        showErr(null);
        await AdminAPI.closeFeedback(thread.id);
        await load();
        if (ctx.refreshBadges) ctx.refreshBadges();
      } catch (err) { showErr(err); }
    }

    async function loadDetail(id, detail, isClosed) {
      const d = await AdminAPI.getFeedback(id);
      clear(detail);
      (d.messages || []).forEach((m) => detail.appendChild(renderMessage(m)));
      if (!isClosed) {
        detail.appendChild(renderReplyForm(id, async () => loadDetail(id, detail, false)));
      }
    }

    function renderReplyForm(id, reloadDetail) {
      const textarea = el('textarea', { maxlength: '4000', placeholder: '回复用户，用户会收到站内消息' });
      const btn = el('button', { type: 'submit', class: 'btn btn-primary btn-sm' }, '发送回复');
      const form = el('form', { style: { marginTop: '0.75rem' } }, [
        el('div', { class: 'field' }, [
          el('label', {}, '管理员回复'),
          textarea,
        ]),
        el('div', { class: 'btn-row' }, [btn]),
      ]);
      form.addEventListener('submit', async (e) => {
        e.preventDefault();
        showErr(null);
        btn.disabled = true;
        try {
          await AdminAPI.replyFeedback(id, textarea.value);
          textarea.value = '';
          await reloadDetail();
        } catch (err) {
          showErr(err);
        } finally {
          btn.disabled = false;
        }
      });
      return form;
    }

    async function load() {
      showErr(null);
      reloadBtn.disabled = true;
      state.status = statusFilter.value;
      state.userId = (userIdInput.value || '').trim();
      try {
        const q = { limit: state.limit, offset: state.offset };
        if (state.status) q.status = state.status;
        if (state.userId) q.user_id = state.userId;
        const d = await AdminAPI.listFeedback(q);
        state.total = typeof d.total === 'number' ? d.total : (Number(d.total) || 0);
        renderList(d.feedback || []);
        const end = Math.min(state.offset + LIMIT, state.total);
        pageMeta.textContent = state.total
          ? '共 ' + state.total + ' 条 · 显示 ' + (state.offset + 1) + '–' + end
          : '暂无数据';
        prevBtn.disabled = state.offset <= 0;
        nextBtn.disabled = state.offset + LIMIT >= state.total;
      } catch (err) {
        showErr(err);
      } finally {
        reloadBtn.disabled = false;
      }
    }

    statusFilter.addEventListener('change', () => { state.offset = 0; load(); });
    userIdInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') { state.offset = 0; load(); }
    });
    reloadBtn.addEventListener('click', () => load());

    await load();
    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.feedback = { mount };
})();
