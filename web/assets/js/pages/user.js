/* User profile page — public profile + published posts. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkMD, BlinkSocial } = window;
  const { el, clear, flash, errorText, createCursorPager, fmtTime } = BlinkUI;

  const params = new URLSearchParams(location.search);
  const userId = params.get('id');

  let myUserId = null;
  let pager = null;
  let profileUserId = null;

  function sid(v) { return v != null ? String(v) : ''; }

  function displayName(u) {
    const n = u && u.name != null ? String(u.name).trim() : '';
    return n || ('用户 ' + sid(u && u.user_id));
  }

  function showErr(msg) { flash('err', msg || '', msg ? 'err' : ''); }

  function renderPost(p) {
    const postHref = '/web/post.html?id=' + encodeURIComponent(sid(p.id));
    const children = [
      el('h2', { style: { fontSize: '1.05rem', margin: '0 0 0.35rem' } }, [
        el('a', { href: postHref }, BlinkMD.plainSnippet(p.body, 80)),
      ]),
      el('div', { class: 'meta' }, [
        '帖子 #' + sid(p.id),
        fmtTime(p.created_at) ? ' · ' + fmtTime(p.created_at) : '',
      ]),
    ];
    if (Array.isArray(p.images) && p.images.length) {
      const thumbs = el('div', { class: 'feed-thumbs' });
      p.images.forEach((url) => {
        if (!url) return;
        thumbs.appendChild(el('a', { class: 'feed-thumb-link', href: postHref }, [
          el('img', { src: url, alt: '', loading: 'lazy' }),
        ]));
      });
      if (thumbs.childNodes.length) children.push(thumbs);
    }
    if (BlinkSocial) {
      const actions = el('div', { class: 'post-social-actions' });
      actions.appendChild(BlinkSocial.makeLikeButton(p, (err) => showErr(errorText(err))));
      children.push(actions);
    }
    return el('article', { class: 'feed-item mine-row', style: { padding: '0.85rem 0' } }, children);
  }

  async function loadPage(cursor) {
    const q = new URLSearchParams();
    q.set('limit', '15');
    if (cursor) q.set('cursor', cursor);
    const d = await BlinkAPI.get('/api/users/' + encodeURIComponent(profileUserId) + '/posts?' + q.toString());
    return {
      items: d.posts || [],
      next: d.next_cursor != null && d.next_cursor !== '' ? d.next_cursor : null,
    };
  }

  function appendRows(items, reset) {
    const list = document.getElementById('list');
    if (reset) clear(list);
    if (reset && items.length === 0) {
      list.appendChild(el('p', { class: 'field-hint', style: { margin: '0.5rem 0' } }, '暂无公开帖子。'));
      return;
    }
    items.forEach((p) => list.appendChild(renderPost(p)));
  }

  async function loadProfile() {
    const u = await BlinkAPI.get('/api/users/' + encodeURIComponent(profileUserId));
    profileUserId = sid(u.user_id);
    document.title = displayName(u) + ' — Blink';
    document.getElementById('profile-name').textContent = displayName(u);
    document.getElementById('profile-meta').textContent = '用户 ID #' + profileUserId;

    const st = await BlinkSocial.loadFollowStats(profileUserId);
    const followingEl = document.getElementById('link-following');
    const followersEl = document.getElementById('link-followers');
    const followingCount = st && st.following_count != null ? String(st.following_count) : '0';
    const followerCount = st && st.follower_count != null ? String(st.follower_count) : '0';
    followingEl.textContent = followingCount;
    followersEl.href = '/web/follows.html?tab=following&user_id=' + encodeURIComponent(profileUserId);
    followersEl.href = '/web/follows.html?tab=followers&user_id=' + encodeURIComponent(profileUserId);
    followingEl.href = '/web/follows.html?tab=following&user_id=' + encodeURIComponent(profileUserId);

    const actions = document.getElementById('profile-actions');
    clear(actions);
    if (myUserId && profileUserId !== myUserId && BlinkSocial) {
      actions.appendChild(BlinkSocial.makeFollowButton(
        profileUserId,
        st && st.is_following === true,
        (err) => showErr(errorText(err)),
      ));
    }
    if (myUserId && profileUserId === myUserId) {
      actions.appendChild(el('a', {
        class: 'btn btn-secondary btn-sm',
        href: '/web/mine.html',
      }, '编辑我的资料'));
    }

    document.getElementById('profile-card').hidden = false;
  }

  async function init() {
    if (!userId) {
      showErr('缺少用户 id');
      return;
    }
    profileUserId = sid(userId);
    try {
      const me = await BlinkAPI.me();
      if (me && me.user_id != null) myUserId = sid(me.user_id);
    } catch (_) { /* guest */ }

    pager = createCursorPager({
      loader: loadPage,
      onAppend: appendRows,
      onError: (err) => showErr(errorText(err)),
      moreButton: 'more',
    });

    try {
      await loadProfile();
      pager.reset();
    } catch (err) {
      if (err && err.status === 404) showErr('用户不存在或不可见');
      else showErr(errorText(err));
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
