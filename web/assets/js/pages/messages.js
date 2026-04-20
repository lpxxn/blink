/**
 * Messages page — paginated notifications feed.
 *
 * Uses BlinkUI.createCursorPager for cursor-driven "more"; actions are
 * per-item mark-read and a global mark-all-read, both via BlinkAPI.
 */
(function () {
  'use strict';

  const { el, clear, flash, errorText, createCursorPager } = window.BlinkUI;
  const API = window.BlinkAPI;

  const listEl = document.getElementById('list');
  const errEl = document.getElementById('err');
  const hintEl = document.getElementById('unread-hint');
  const moreBtn = document.getElementById('more');
  const markAllBtn = document.getElementById('mark-all');
  const emptyHintText = '暂无消息';

  let rendered = 0;

  function showErr(err) {
    flash(errEl, err ? errorText(err) : '', err ? 'err' : '');
  }

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  function renderOne(n) {
    const wrap = el('div', { class: 'notif-item' + (n.read ? '' : ' unread') });

    const head = el('div', {
      style: {
        display: 'flex', flexWrap: 'wrap',
        gap: '0.35rem 0.75rem', alignItems: 'baseline',
      },
    }, [
      el('strong', {}, n.title || '通知'),
      el('span', {
        style: { color: 'var(--text-muted)', fontSize: '0.85rem' },
      }, (n.type || '') + ' · ' + fmtTime(n.created_at)),
    ]);
    wrap.appendChild(head);

    wrap.appendChild(el('div', {
      style: { whiteSpace: 'pre-wrap', marginTop: '0.35rem', fontSize: '0.92rem' },
    }, n.body || ''));

    if (n.ref_post_id != null && n.ref_post_id !== '') {
      wrap.appendChild(el('a', {
        href: '/web/post.html?id=' + encodeURIComponent(String(n.ref_post_id)),
        style: { display: 'inline-block', marginTop: '0.5rem', fontSize: '0.9rem' },
      }, '查看帖子'));
    }

    let readBtn = null;
    if (!n.read) {
      readBtn = el('button', {
        type: 'button', class: 'btn btn-ghost btn-sm',
        style: { marginTop: '0.35rem' },
        onClick: async () => {
          readBtn.disabled = true;
          try {
            await API.post('/api/me/notifications/' + encodeURIComponent(String(n.id)) + '/read');
            n.read = true;
            wrap.className = 'notif-item';
            readBtn.remove();
            loadUnread();
          } catch (err) {
            readBtn.disabled = false;
            showErr(err);
          }
        },
      }, '标为已读');
      wrap.appendChild(readBtn);
    }
    return wrap;
  }

  function handleAppend(items, reset) {
    if (reset) {
      clear(listEl);
      rendered = 0;
    }
    (items || []).forEach((n) => {
      listEl.appendChild(renderOne(n));
      rendered += 1;
    });
    if (reset && rendered === 0) {
      listEl.appendChild(el('p', { class: 'empty-hint' }, emptyHintText));
    }
    loadUnread();
  }

  async function loadUnread() {
    try {
      const d = await API.get('/api/me/notifications/unread_count');
      if (!d || d.unread_count == null) { hintEl.textContent = ''; return; }
      hintEl.textContent = '未读 ' + String(d.unread_count);
    } catch (_) { /* ignore — hint is advisory */ }
  }

  async function loader(cursor) {
    const path = '/api/me/notifications?limit=30' + (cursor ? '&cursor=' + encodeURIComponent(cursor) : '');
    const d = await API.get(path);
    return { items: d.notifications || [], next: d.next_cursor || null };
  }

  const pager = createCursorPager({
    loader,
    onAppend: handleAppend,
    onError: showErr,
    moreButton: moreBtn,
  });

  markAllBtn.addEventListener('click', async () => {
    showErr(null);
    markAllBtn.disabled = true;
    try {
      await API.post('/api/me/notifications/read_all');
      await pager.reset();
    } catch (err) {
      showErr(err);
    } finally {
      markAllBtn.disabled = false;
    }
  });

  pager.reset();
})();
