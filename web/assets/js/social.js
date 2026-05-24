/**
 * Social actions — follow users and like posts.
 */
(function () {
  'use strict';

  function sid(v) { return v != null ? String(v) : ''; }

  function likeLabel(count, liked) {
    const n = count != null ? Number(count) : 0;
    const heart = liked ? '♥' : '♡';
    return heart + ' ' + (isNaN(n) ? '0' : n);
  }

  async function toggleLike(postId, btn, countEl) {
    if (!postId) return false;
    const liked = btn.getAttribute('data-liked') === '1';
    const path = '/api/posts/' + encodeURIComponent(sid(postId)) + '/like';
    try {
      if (liked) {
        await window.BlinkAPI.del(path);
        btn.setAttribute('data-liked', '0');
      } else {
        await window.BlinkAPI.post(path, {});
        btn.setAttribute('data-liked', '1');
      }
      const d = await window.BlinkAPI.get(path.replace('/like', '/likes'));
      const cnt = d && d.like_count != null ? d.like_count : 0;
      const nowLiked = d && d.liked === true;
      btn.setAttribute('data-liked', nowLiked ? '1' : '0');
      btn.textContent = likeLabel(cnt, nowLiked);
      if (countEl) countEl.textContent = '';
      return true;
    } catch (err) {
      if (err && err.status === 401) {
        window.location.href = '/web/login.html?next=' + encodeURIComponent(location.pathname + location.search);
        return false;
      }
      throw err;
    }
  }

  function makeLikeButton(post, onError) {
    const postId = sid(post.id);
    const liked = post.liked === true;
    const count = post.like_count != null ? post.like_count : 0;
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'btn btn-ghost btn-sm social-like-btn' + (liked ? ' social-liked' : '');
    btn.setAttribute('data-liked', liked ? '1' : '0');
    btn.textContent = likeLabel(count, liked);
    btn.addEventListener('click', async () => {
      try {
        await toggleLike(postId, btn, null);
        btn.classList.toggle('social-liked', btn.getAttribute('data-liked') === '1');
      } catch (err) {
        if (onError) onError(err);
      }
    });
    return btn;
  }

  async function toggleFollow(userId, btn) {
    if (!userId) return false;
    const following = btn.getAttribute('data-following') === '1';
    const path = '/api/users/' + encodeURIComponent(sid(userId)) + '/follow';
    try {
      if (following) {
        await window.BlinkAPI.del(path);
        btn.setAttribute('data-following', '0');
        btn.textContent = '关注';
      } else {
        await window.BlinkAPI.post(path, {});
        btn.setAttribute('data-following', '1');
        btn.textContent = '已关注';
      }
      return true;
    } catch (err) {
      if (err && err.status === 401) {
        window.location.href = '/web/login.html?next=' + encodeURIComponent(location.pathname + location.search);
        return false;
      }
      throw err;
    }
  }

  function makeFollowButton(userId, isFollowing, onError) {
    const btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'btn btn-secondary btn-sm';
    btn.setAttribute('data-following', isFollowing ? '1' : '0');
    btn.textContent = isFollowing ? '已关注' : '关注';
    btn.addEventListener('click', async () => {
      try {
        await toggleFollow(userId, btn);
      } catch (err) {
        if (onError) onError(err);
      }
    });
    return btn;
  }

  async function loadFollowStats(userId) {
    return window.BlinkAPI.get('/api/users/' + encodeURIComponent(sid(userId)) + '/follow-stats');
  }

  window.BlinkSocial = {
    sid,
    likeLabel,
    toggleLike,
    makeLikeButton,
    toggleFollow,
    makeFollowButton,
    loadFollowStats,
  };
})();
