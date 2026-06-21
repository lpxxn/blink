/**
 * Trending page — post like rankings (day / week / month).
 */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkMD, BlinkSocial } = window;
  const { el, flash, clear, fmtTime, authorLink } = BlinkUI;

  const PERIODS = [
    { key: 'day', label: '今日' },
    { key: 'week', label: '近 7 天' },
    { key: 'month', label: '本月' },
  ];

  let currentPeriod = 'week';

  function postMeta(p) {
    const meta = el('div', { class: 'meta' });
    meta.appendChild(document.createTextNode('帖子 #' + p.id + ' · '));
    meta.appendChild(authorLink(p.user_id, p.user_name));
    const t = fmtTime(p.created_at);
    if (t) meta.appendChild(document.createTextNode(' · ' + t));
    return meta;
  }

  function renderPost(p) {
    const postHref = '/web/post.html?id=' + encodeURIComponent(p.id);
    const rank = p.rank != null ? p.rank : '';
    const likeCount = p.period_like_count != null ? p.period_like_count : '0';
    const children = [
      el('div', { class: 'meta' }, '#' + rank + ' · ' + likeCount + ' 赞'),
      el('h2', {}, [el('a', { href: postHref }, BlinkMD.plainSnippet(p.body, 80))]),
      postMeta(p),
    ];
    if (Array.isArray(p.images) && p.images.length) {
      const thumbs = el('div', { class: 'feed-thumbs' });
      p.images.forEach(function (url) {
        if (!url) return;
        thumbs.appendChild(
          el('a', { class: 'feed-thumb-link', href: postHref }, [
            el('img', { src: url, alt: '', loading: 'lazy' }),
          ])
        );
      });
      if (thumbs.childNodes.length) children.push(thumbs);
    }
    const actions = el('div', { class: 'post-social-actions' });
    if (BlinkSocial) {
      actions.appendChild(BlinkSocial.makeLikeButton(p, function (err) {
        flash(document.getElementById('err'), BlinkUI.errorText(err));
      }));
    }
    children.push(actions);
    return el('article', { class: 'feed-item' }, children);
  }

  async function loadRankings(period) {
    const errEl = document.getElementById('err');
    flash(errEl, '');
    const listEl = document.getElementById('list');
    clear(listEl);
    listEl.appendChild(el('p', { class: 'empty-hint' }, '加载中…'));
    try {
      const d = await BlinkAPI.get('/api/posts/like_rankings?period=' + encodeURIComponent(period) + '&limit=20');
      clear(listEl);
      const posts = d.posts || [];
      if (!posts.length) {
        listEl.appendChild(el('p', { class: 'empty-hint' }, '暂无数据'));
        return;
      }
      posts.forEach(function (p) {
        listEl.appendChild(renderPost(p));
      });
    } catch (err) {
      clear(listEl);
      flash(errEl, BlinkUI.errorText(err));
    }
  }

  function renderTabs() {
    const wrap = document.getElementById('period-tabs');
    clear(wrap);
    PERIODS.forEach(function (p) {
      const btn = el('button', {
        type: 'button',
        class: p.key === currentPeriod ? 'active' : '',
        onClick: function () {
          if (currentPeriod === p.key) return;
          currentPeriod = p.key;
          renderTabs();
          loadRankings(currentPeriod);
        },
      }, p.label);
      wrap.appendChild(btn);
    });
  }

  renderTabs();
  loadRankings(currentPeriod);
})();
