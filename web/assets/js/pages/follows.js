/* Follows page — my or another user's following / followers lists. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkSocial } = window;
  const { el, clear, flash, errorText, createCursorPager, authorLink, authorLabelText } = BlinkUI;

  const params = new URLSearchParams(location.search);
  let activeTab = params.get('tab') === 'followers' ? 'followers' : 'following';
  const viewUserIdParam = params.get('user_id');
  let viewUserId = viewUserIdParam ? String(viewUserIdParam) : null;
  let viewUserName = null;
  let myUserId = null;
  let pager = null;

  function sid(v) { return v != null ? String(v) : ''; }

  function subjectLabel() {
    if (!viewUserId || (myUserId && viewUserId === myUserId)) return '你';
    return viewUserName || ('用户 ' + viewUserId);
  }

  function updateHeadings() {
    const subject = subjectLabel();
    const isSelf = !viewUserId || (myUserId && viewUserId === myUserId);
    document.getElementById('page-title').textContent = activeTab === 'followers'
      ? (isSelf ? '关注我的' : subject + ' 的粉丝')
      : (isSelf ? '我关注的' : subject + ' 关注的');
    document.getElementById('page-lead').textContent = activeTab === 'followers'
      ? (isSelf ? '这里列出关注你的用户。' : '这里列出关注 ' + subject + ' 的用户。')
      : (isSelf ? '这里列出你正在关注的用户。' : '这里列出 ' + subject + ' 正在关注的用户。');
  }

  function setTab(tab) {
    activeTab = tab;
    const q = new URLSearchParams(location.search);
    q.set('tab', tab);
    if (viewUserId) q.set('user_id', viewUserId);
    else q.delete('user_id');
    history.replaceState(null, '', location.pathname + '?' + q.toString());
    document.getElementById('tab-following').classList.toggle('btn-primary', tab === 'following');
    document.getElementById('tab-following').classList.toggle('btn-secondary', tab !== 'following');
    document.getElementById('tab-followers').classList.toggle('btn-primary', tab === 'followers');
    document.getElementById('tab-followers').classList.toggle('btn-secondary', tab !== 'followers');
    updateHeadings();
    pager.reset();
  }

  function renderRow(u) {
    const uid = sid(u.user_id);
    const row = el('div', { class: 'mine-row' });
    row.appendChild(el('p', { class: 'mine-snippet', style: { marginBottom: '0.35rem' } }, [
      authorLink(u.user_id, u.user_name),
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
    const segment = activeTab === 'followers' ? 'followers' : 'following';
    const path = viewUserId
      ? '/api/users/' + encodeURIComponent(viewUserId) + '/' + segment
      : '/api/me/' + segment;
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
      const isSelf = !viewUserId || (myUserId && viewUserId === myUserId);
      const empty = activeTab === 'followers'
        ? (isSelf ? '还没有人关注你。' : '还没有人关注 TA。')
        : (isSelf ? '你还没有关注任何人。' : 'TA 还没有关注任何人。');
      list.appendChild(el('p', {
        class: 'field-hint',
        style: { margin: '0.5rem 0' },
      }, empty));
      return;
    }
    items.forEach((u) => list.appendChild(renderRow(u)));
  }

  async function loadViewUser() {
    if (!viewUserId) return;
    try {
      const u = await BlinkAPI.get('/api/users/' + encodeURIComponent(viewUserId));
      viewUserId = sid(u.user_id);
      viewUserName = authorLabelText(u.user_id, u.name);
    } catch (err) {
      flash('err', errorText(err));
      throw err;
    }
  }

  async function init() {
    try {
      const me = await BlinkAPI.me();
      if (me && me.user_id != null) myUserId = sid(me.user_id);
    } catch (_) { /* guest */ }

    if (viewUserId) {
      await loadViewUser();
    } else {
      if (!myUserId) {
        window.location.href = '/web/login.html?next=' + encodeURIComponent(location.pathname + location.search);
        return;
      }
      viewUserId = myUserId;
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
