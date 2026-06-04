/* Follows page — my following / followers lists. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkSocial } = window;
  const { el, clear, flash, errorText, createCursorPager } = BlinkUI;

  const params = new URLSearchParams(location.search);
  let activeTab = params.get('tab') === 'followers' ? 'followers' : 'following';
  let myUserId = null;
  let pager = null;

  function sid(v) { return v != null ? String(v) : ''; }

  function userLabel(u) {
    const n = u.user_name != null ? String(u.user_name).trim() : '';
    return n || ('用户 ' + sid(u.user_id));
  }

  function setTab(tab) {
    activeTab = tab;
    const q = new URLSearchParams(location.search);
    q.set('tab', tab);
    history.replaceState(null, '', location.pathname + '?' + q.toString());
    document.getElementById('tab-following').classList.toggle('btn-primary', tab === 'following');
    document.getElementById('tab-following').classList.toggle('btn-secondary', tab !== 'following');
    document.getElementById('tab-followers').classList.toggle('btn-primary', tab === 'followers');
    document.getElementById('tab-followers').classList.toggle('btn-secondary', tab !== 'followers');
    document.getElementById('page-title').textContent = tab === 'followers' ? '关注我的' : '我关注的';
    document.getElementById('page-lead').textContent = tab === 'followers'
      ? '这里列出关注你的用户。'
      : '这里列出你正在关注的用户。';
    pager.reset();
  }

  function renderRow(u) {
    const uid = sid(u.user_id);
    const row = el('div', { class: 'mine-row' });
    row.appendChild(el('p', { class: 'mine-snippet', style: { marginBottom: '0.35rem' } }, [
      el('strong', {}, userLabel(u)),
      document.createTextNode(' · #' + uid),
    ]));

    const actions = el('div', { class: 'mine-actions' });
    if (myUserId && uid !== myUserId && BlinkSocial) {
      actions.appendChild(BlinkSocial.makeFollowButton(uid, u.is_following === true, (err) => {
        flash('err', errorText(err));
      }));
    }
    row.appendChild(actions);
    return row;
  }

  async function loadPage(cursor) {
    const path = activeTab === 'followers' ? '/api/me/followers' : '/api/me/following';
    const q = new URLSearchParams();
    q.set('limit', '30');
    if (cursor) q.set('cursor', cursor);
    const d = await BlinkAPI.get(path + '?' + q.toString());
    return {
      items: d.users || [],
      next: d.next_cursor != null && d.next_cursor !== '' ? d.next_cursor : null,
    };
  }

  function appendRows(items, reset) {
    const list = document.getElementById('list');
    if (reset) clear(list);
    if (reset && items.length === 0) {
      list.appendChild(el('p', {
        class: 'field-hint',
        style: { margin: '0.5rem 0' },
      }, activeTab === 'followers' ? '还没有人关注你。' : '你还没有关注任何人。'));
      return;
    }
    items.forEach((u) => list.appendChild(renderRow(u)));
  }

  async function init() {
    try {
      const me = await BlinkAPI.me();
      if (!me || me.user_id == null) {
        window.location.href = '/web/login.html?next=' + encodeURIComponent(location.pathname + location.search);
        return;
      }
      myUserId = sid(me.user_id);
    } catch (_) {
      window.location.href = '/web/login.html?next=' + encodeURIComponent(location.pathname + location.search);
      return;
    }

    pager = createCursorPager({
      loader: loadPage,
      onAppend: appendRows,
      onError: (err) => flash('err', errorText(err)),
      moreButton: 'more',
    });

    document.getElementById('tab-following').addEventListener('click', () => setTab('following'));
    document.getElementById('tab-followers').addEventListener('click', () => setTab('followers'));
    setTab(activeTab);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
