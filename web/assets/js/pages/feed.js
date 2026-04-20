/* Feed page: category tabs + cursor-paginated post list. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkMD } = window;
  const { el, flash, mountList, createCursorPager } = BlinkUI;

  let categoryId = null;     // null = no specific cat; use with `uncategorized` flag
  let uncategorized = false;
  let pager = null;

  function authorLabel(p) {
    const n = p && p.user_name != null ? String(p.user_name).trim() : '';
    return n || ('用户 ' + (p && p.user_id));
  }

  function renderPost(p) {
    const postHref = '/web/post.html?id=' + encodeURIComponent(p.id);
    const children = [
      el('h2', {}, [el('a', { href: postHref }, BlinkMD.plainSnippet(p.body, 80))]),
      el('div', { class: 'meta' }, '帖子 #' + p.id + ' · ' + authorLabel(p)),
    ];
    if (Array.isArray(p.images) && p.images.length) {
      const thumbs = el('div', { class: 'feed-thumbs' });
      p.images.forEach((url) => {
        if (!url) return;
        thumbs.appendChild(
          el('a', { class: 'feed-thumb-link', href: postHref }, [
            el('img', { src: url, alt: '', loading: 'lazy' }),
          ])
        );
      });
      if (thumbs.childNodes.length) children.push(thumbs);
    }
    return el('article', { class: 'feed-item' }, children);
  }

  async function loadPage(cursor) {
    const q = new URLSearchParams();
    q.set('limit', '15');
    if (categoryId != null) q.set('category_id', String(categoryId));
    if (uncategorized) q.set('uncategorized', '1');
    if (cursor) q.set('cursor', cursor);
    const d = await BlinkAPI.get('/api/posts?' + q.toString());
    return {
      items: d.posts || [],
      next: d.next_cursor != null && d.next_cursor !== '' ? d.next_cursor : null,
    };
  }

  function renderList(items, reset) {
    const list = document.getElementById('list');
    if (reset) {
      if (!items.length) {
        mountList(list, [], null, '暂无帖子');
        return;
      }
      while (list.firstChild) list.removeChild(list.firstChild);
    }
    items.forEach((p) => list.appendChild(renderPost(p)));
  }

  async function loadCategories() {
    const wrap = document.getElementById('cats');
    try {
      const d = await BlinkAPI.get('/api/categories');
      const cats = d.categories || [];
      while (wrap.firstChild) wrap.removeChild(wrap.firstChild);

      const makeTab = (label, isActive, onClick) => {
        const btn = el('button', { type: 'button', onClick }, label);
        if (isActive) btn.className = 'active';
        return btn;
      };

      wrap.appendChild(makeTab('全部', categoryId === null && !uncategorized, () => {
        selectTab(null, false);
      }));
      wrap.appendChild(makeTab('未分类', uncategorized, () => {
        selectTab(null, true);
      }));
      cats.forEach((c) => {
        wrap.appendChild(makeTab(c.name, String(categoryId) === String(c.id), () => {
          selectTab(String(c.id), false);
        }));
      });
    } catch (err) {
      flash('err', '分类加载失败：' + BlinkUI.errorText(err));
    }
  }

  function selectTab(catId, uncat) {
    categoryId = catId;
    uncategorized = !!uncat;
    loadCategories();
    pager.reset();
  }

  function init() {
    pager = createCursorPager({
      loader: loadPage,
      onAppend: renderList,
      onError: (err) => flash('err', BlinkUI.errorText(err)),
      moreButton: 'more',
    });
    loadCategories();
    pager.reset();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
