/* Mine page: profile (name/logout) + my-posts list with quick actions. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkMD } = window;
  const { el, clear, flash, errorText, createCursorPager } = BlinkUI;

  let me = null;
  let pager = null;

  function sid(v) { return v != null ? String(v) : ''; }

  function statusLabel(s) {
    if (s === 0) return '草稿';
    if (s === 1) return '已发布';
    if (s === 2) return '已隐藏';
    return '状态 ' + s;
  }

  function showProfileMsg(text, ok) {
    flash('profile-msg', text || '', ok ? 'ok' : 'err');
  }

  function applyMe(d) {
    me = d;
    document.getElementById('me-id').textContent = d.user_id != null ? String(d.user_id) : '—';
    document.getElementById('me-email').textContent = d.email || '—';
    document.getElementById('me-role').textContent = d.role || '—';
    document.getElementById('me-name').value = d.name || '';
    document.getElementById('profile-guest').hidden = true;
    document.getElementById('profile-user').hidden = false;
  }

  function showGuest() {
    me = null;
    document.getElementById('profile-guest').hidden = false;
    document.getElementById('profile-user').hidden = true;
    const list = document.getElementById('list');
    clear(list);
    list.appendChild(el('p', {
      class: 'field-hint',
      style: { margin: '0.5rem 0' },
    }, '登录后可在此查看与管理帖子。'));
    document.getElementById('more').hidden = true;
  }

  // ---------- Posts ----------
  function renderRow(p) {
    const pid = sid(p.id);
    const row = el('div', { class: 'mine-row' });

    row.appendChild(el('p', { class: 'mine-snippet' }, [
      el('a', { href: '/web/post.html?id=' + encodeURIComponent(pid) },
        BlinkMD.plainSnippet(p.body, 100)),
    ]));

    const meta = el('div', { class: 'meta' }, '#' + pid + ' · ' + statusLabel(p.status));
    row.appendChild(meta);

    const actions = el('div', { class: 'mine-actions' });
    if (p.status === 0) {
      const pubBtn = el('button', {
        type: 'button', class: 'btn btn-primary btn-sm',
        onClick: async () => {
          flash('err', '');
          try {
            await BlinkAPI.patch('/api/posts/' + encodeURIComponent(pid), { status: 1 });
            meta.textContent = '#' + pid + ' · ' + statusLabel(1);
            pubBtn.remove();
          } catch (err) {
            flash('err', errorText(err));
          }
        },
      }, '发布');
      actions.appendChild(pubBtn);
    }
    actions.appendChild(el('a', {
      class: 'btn btn-secondary btn-sm',
      href: '/web/edit-post.html?id=' + encodeURIComponent(pid),
    }, '编辑'));
    actions.appendChild(el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: async () => {
        if (!window.confirm('确定删除该帖子？不可恢复。')) return;
        try {
          await BlinkAPI.del('/api/posts/' + encodeURIComponent(pid));
          row.remove();
        } catch (err) {
          flash('err', errorText(err));
        }
      },
    }, '删除'));

    row.appendChild(actions);
    return row;
  }

  async function loadPage(cursor) {
    const q = new URLSearchParams();
    q.set('include_draft', '1');
    q.set('limit', '15');
    if (cursor) q.set('cursor', cursor);
    const d = await BlinkAPI.get('/api/me/posts?' + q.toString());
    return {
      items: d.posts || [],
      next: d.next_cursor != null && d.next_cursor !== '' ? d.next_cursor : null,
    };
  }

  function appendRows(items, reset) {
    const list = document.getElementById('list');
    if (reset) clear(list);
    items.forEach((p) => list.appendChild(renderRow(p)));
  }

  // ---------- Profile actions ----------
  async function saveName() {
    const name = document.getElementById('me-name').value.trim();
    if (!name) {
      showProfileMsg('昵称不能为空', false);
      return;
    }
    try {
      const d = await BlinkAPI.patch('/api/me', { name });
      if (d) applyMe(d);
      showProfileMsg('已保存', true);
    } catch (err) {
      if (err && err.status === 401) {
        showProfileMsg('请先登录', false);
        showGuest();
      } else {
        showProfileMsg(errorText(err), false);
      }
    }
  }

  async function doLogout() {
    try { await BlinkAPI.logout(); } catch (_) { /* ignore */ }
    window.location.href = '/web/index.html';
  }

  // ---------- Init ----------
  async function loadProfile() {
    showProfileMsg('', true);
    try {
      const d = await BlinkAPI.me();
      if (!d) {
        showGuest();
        return;
      }
      applyMe(d);
      pager.reset();
    } catch (err) {
      showProfileMsg(errorText(err), false);
    }
  }

  function init() {
    pager = createCursorPager({
      loader: loadPage,
      onAppend: appendRows,
      onError: (err) => flash('err', errorText(err)),
      moreButton: 'more',
    });
    document.getElementById('save-name').addEventListener('click', saveName);
    document.getElementById('logout').addEventListener('click', doLogout);
    loadProfile();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
