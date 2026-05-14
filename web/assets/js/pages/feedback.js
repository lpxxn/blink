/**
 * Feedback page — submit feedback and follow up on admin replies.
 */
(function () {
  'use strict';

  const { el, clear, flash, errorText, createCursorPager } = window.BlinkUI;
  const API = window.BlinkAPI;

  const msgEl = document.getElementById('feedback-msg');
  const formEl = document.getElementById('feedback-form');
  const bodyEl = document.getElementById('feedback-body');
  const submitBtn = document.getElementById('feedback-submit');
  const listEl = document.getElementById('feedback-list');
  const moreBtn = document.getElementById('feedback-more');

  let rendered = 0;

  function showMsg(message, level) {
    flash(msgEl, message || '', level || (message ? 'err' : ''));
  }

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  function renderMessage(m) {
    const mine = m.sender_type === 'user';
    return el('div', {
      class: 'reply' + (mine ? '' : ' reply-nested'),
      style: { marginTop: '0.65rem' },
    }, [
      el('div', { class: 'meta' }, (mine ? '我' : '管理员') + ' · ' + fmtTime(m.created_at)),
      el('div', { style: { whiteSpace: 'pre-wrap', marginTop: '0.25rem' } }, m.body || ''),
    ]);
  }

  async function openThread(card, thread) {
    const mount = card.querySelector('[data-feedback-detail]');
    const btn = card.querySelector('[data-feedback-toggle]');
    if (!mount || !btn) return;
    if (!mount.hidden) {
      mount.hidden = true;
      btn.textContent = '展开';
      return;
    }
    btn.disabled = true;
    try {
      await loadThreadDetail(thread.id, mount);
      mount.hidden = false;
      btn.textContent = '收起';
    } catch (err) {
      showMsg(errorText(err), 'err');
    } finally {
      btn.disabled = false;
    }
  }

  async function loadThreadDetail(id, mount) {
    const d = await API.get('/api/me/feedback/' + encodeURIComponent(String(id)));
    clear(mount);
    (d.messages || []).forEach((m) => mount.appendChild(renderMessage(m)));
    mount.appendChild(renderReplyForm(d.thread, () => loadThreadDetail(id, mount)));
  }

  function renderReplyForm(thread, reload) {
    if (!thread.can_user_reply) {
      return el('p', { class: 'field-hint' }, '已达到追加反馈次数上限，或反馈已关闭。');
    }
    const textarea = el('textarea', { maxlength: '4000', placeholder: '继续补充你的反馈' });
    const btn = el('button', { type: 'submit', class: 'btn btn-secondary btn-sm' }, '追加反馈');
    const form = el('form', { style: { marginTop: '0.75rem' } }, [
      el('div', { class: 'field' }, [
        el('label', {}, '补充说明'),
        textarea,
        el('p', { class: 'field-hint' }, '还可以补充 ' + Math.max(0, 2 - (thread.user_reply_count || 0)) + ' 次。'),
      ]),
      el('div', { class: 'btn-row' }, [btn]),
    ]);
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      showMsg('');
      btn.disabled = true;
      try {
        await API.post('/api/me/feedback/' + encodeURIComponent(String(thread.id)) + '/replies', {
          body: textarea.value,
        });
        showMsg('已追加反馈', 'ok');
        if (typeof reload === 'function') await reload();
      } catch (err) {
        showMsg(errorText(err), 'err');
      } finally {
        btn.disabled = false;
      }
    });
    return form;
  }

  function renderThread(thread) {
    const card = el('div', { class: 'notif-item' });
    const title = '反馈 #' + thread.id;
    const meta = [
      thread.status === 'open' ? '处理中' : thread.status,
      '用户补充 ' + (thread.user_reply_count || 0) + '/2',
      fmtTime(thread.last_message_at),
    ].filter(Boolean).join(' · ');
    const toggle = el('button', {
      type: 'button',
      class: 'btn btn-ghost btn-sm',
      'data-feedback-toggle': '1',
    }, '展开');
    const detail = el('div', { 'data-feedback-detail': '1', hidden: true });
    toggle.addEventListener('click', () => openThread(card, thread));
    card.appendChild(el('div', {
      style: { display: 'flex', gap: '0.75rem', alignItems: 'baseline', flexWrap: 'wrap' },
    }, [
      el('strong', {}, title),
      el('span', { class: 'meta' }, meta),
    ]));
    card.appendChild(el('div', { class: 'btn-row', style: { marginTop: '0.45rem' } }, [toggle]));
    card.appendChild(detail);
    return card;
  }

  function handleAppend(items, reset) {
    if (reset) {
      clear(listEl);
      rendered = 0;
    }
    (items || []).forEach((thread) => {
      listEl.appendChild(renderThread(thread));
      rendered += 1;
    });
    if (reset && rendered === 0) {
      listEl.appendChild(el('p', { class: 'empty-hint' }, '暂无反馈'));
    }
  }

  async function loader(cursor) {
    const path = '/api/me/feedback?limit=20' + (cursor ? '&cursor=' + encodeURIComponent(cursor) : '');
    const d = await API.get(path);
    return { items: d.feedback || [], next: d.next_cursor || null };
  }

  const pager = createCursorPager({
    loader,
    onAppend: handleAppend,
    onError: (err) => showMsg(errorText(err), 'err'),
    moreButton: moreBtn,
  });

  formEl.addEventListener('submit', async (e) => {
    e.preventDefault();
    showMsg('');
    submitBtn.disabled = true;
    try {
      await API.post('/api/feedback', { body: bodyEl.value });
      bodyEl.value = '';
      showMsg('反馈已提交，管理员会收到站内消息。', 'ok');
      await pager.reset();
    } catch (err) {
      showMsg(errorText(err), 'err');
    } finally {
      submitBtn.disabled = false;
    }
  });

  pager.reset();
})();
