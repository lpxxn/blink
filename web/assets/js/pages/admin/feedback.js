/**
 * Admin — user feedback conversations.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;

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

  function mount(container) {
    const state = { offset: 0, limit: 30, total: 0 };
    const errEl = el('p', { class: 'err', role: 'alert' });
    const statusEl = el('p', { class: 'admin-subtitle' }, '正在加载...');
    const reloadBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm' }, '刷新');
    const listEl = el('div', { class: 'admin-cards' });
    const toolbar = el('div', { class: 'admin-toolbar' }, [
      statusEl,
      el('span', { class: 'admin-toolbar-spacer' }),
      reloadBtn,
    ]);
    container.appendChild(errEl);
    container.appendChild(toolbar);
    container.appendChild(listEl);

    async function load() {
      errEl.textContent = '';
      reloadBtn.disabled = true;
      try {
        const d = await AdminAPI.listFeedback({ limit: state.limit, offset: state.offset });
        state.total = d.total || 0;
        statusEl.textContent = '共 ' + state.total + ' 条反馈';
        renderList(d.feedback || []);
      } catch (err) {
        errEl.textContent = errorText(err);
      } finally {
        reloadBtn.disabled = false;
      }
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
      const title = '反馈 #' + thread.id;
      const user = (thread.user_name || '') + ' ' + thread.user_id;
      card.appendChild(el('div', { class: 'admin-card-head' }, [
        el('strong', {}, title),
        el('span', { class: 'admin-subtitle' }, user.trim()),
      ]));
      card.appendChild(el('p', { class: 'admin-subtitle' },
        '状态：' + (thread.status || 'open') +
        ' · 用户补充 ' + (thread.user_reply_count || 0) + '/2' +
        ' · 最后更新 ' + fmtTime(thread.last_message_at)
      ));
      card.appendChild(el('div', { class: 'admin-card-actions' }, [openBtn]));
      card.appendChild(detail);
      openBtn.addEventListener('click', async () => {
        if (!detail.hidden) {
          detail.hidden = true;
          openBtn.textContent = '查看/回复';
          return;
        }
        openBtn.disabled = true;
        try {
          await loadDetail(thread.id, detail);
          detail.hidden = false;
          openBtn.textContent = '收起';
        } catch (err) {
          errEl.textContent = errorText(err);
        } finally {
          openBtn.disabled = false;
        }
      });
      return card;
    }

    async function loadDetail(id, detail) {
      const d = await AdminAPI.getFeedback(id);
      clear(detail);
      (d.messages || []).forEach((m) => detail.appendChild(renderMessage(m)));
      detail.appendChild(renderReplyForm(id, async () => loadDetail(id, detail)));
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
        errEl.textContent = '';
        btn.disabled = true;
        try {
          await AdminAPI.replyFeedback(id, textarea.value);
          await reloadDetail();
        } catch (err) {
          errEl.textContent = errorText(err);
        } finally {
          btn.disabled = false;
        }
      });
      return form;
    }

    reloadBtn.addEventListener('click', load);
    load();
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.feedback = { mount };
})();
