/* Compose page: wires BlinkPostEditor → HTTP + uploads + local autosave. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkPostEditor, BlinkUI } = window;
  const { el } = BlinkUI;

  const DRAFT_KEY = 'blink.compose.draft.v1';

  // ---------- localStorage draft helpers ----------
  function readDraft() {
    try {
      const raw = localStorage.getItem(DRAFT_KEY);
      if (!raw) return null;
      const d = JSON.parse(raw);
      return d && typeof d === 'object' ? d : null;
    } catch (_) { return null; }
  }

  function writeDraft(obj) {
    try { localStorage.setItem(DRAFT_KEY, JSON.stringify(obj)); }
    catch (_) { /* quota — silent */ }
  }

  function clearDraft() {
    try { localStorage.removeItem(DRAFT_KEY); } catch (_) {}
  }

  function timeSince(ts) {
    if (!ts) return '';
    const s = Math.max(0, Math.floor((Date.now() - ts) / 1000));
    if (s < 5) return '刚刚';
    if (s < 60) return s + ' 秒前';
    if (s < 3600) return Math.floor(s / 60) + ' 分钟前';
    if (s < 86400) return Math.floor(s / 3600) + ' 小时前';
    return Math.floor(s / 86400) + ' 天前';
  }

  // ---------- categories ----------
  async function loadCategories() {
    try {
      const d = await BlinkAPI.get('/api/categories');
      return (d && d.categories) || [];
    } catch (_) { return []; }
  }

  // ---------- image upload with progress (XHR) ----------
  function uploadImage(file, onProgress) {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open('POST', '/api/uploads');
      xhr.withCredentials = true;
      if (xhr.upload && typeof onProgress === 'function') {
        xhr.upload.onprogress = (e) => {
          if (e.lengthComputable) onProgress(e.loaded / e.total);
        };
      }
      xhr.onload = () => {
        if (xhr.status === 401) return reject(new BlinkAPI.Error(401, '请先登录'));
        if (xhr.status === 413) return reject(new BlinkAPI.Error(413, '图片过大'));
        let parsed = null;
        try { parsed = JSON.parse(xhr.responseText); } catch (_) { /* ignore */ }
        if (xhr.status >= 200 && xhr.status < 300) {
          if (parsed && parsed.url) return resolve(parsed);
          return reject(new BlinkAPI.Error(xhr.status, '上传响应缺少 url'));
        }
        const msg = (parsed && parsed.error) || ('上传失败 (' + xhr.status + ')');
        reject(new BlinkAPI.Error(xhr.status, msg, parsed));
      };
      xhr.onerror = () => reject(new BlinkAPI.Error(0, '网络异常，请稍后重试'));
      xhr.onabort = () => reject(new BlinkAPI.Error(0, '上传已取消'));
      const fd = new FormData();
      fd.append('file', file);
      xhr.send(fd);
    });
  }

  // ---------- page bootstrap ----------
  async function init() {
    const mount = document.getElementById('editor-mount');
    const bannerMount = document.getElementById('banner-mount');
    if (!mount) return;

    let serverDraftId = null;
    let localSavedAt = null;
    let serverSavedAt = null;
    let dirty = false;
    let statusTicker = null;

    const editor = BlinkPostEditor.mount(mount, {
      mode: 'create',
      initial: { body: '', images: [], status: 1 },
      onImageUpload: uploadImage,
      onChange: (val) => {
        dirty = true;
        localSavedAt = Date.now();
        writeDraft({
          body: val.body,
          images: val.images,
          categoryId: val.categoryId,
          serverDraftId,
          ts: localSavedAt,
        });
        updateStatus();
      },
      onSubmit: submit,
      onError: (err) => {
        if (err && err.body && Array.isArray(err.body.sensitive_words) && err.body.sensitive_words.length) {
          const words = err.body.sensitive_words;
          editor.setSensitiveHits(words);
          editor.setError('内容包含敏感词：' + words.join('、') + '（已在预览里标黄）');
          return true;
        }
        return false;
      },
    });

    function updateStatus() {
      if (!dirty && !localSavedAt && !serverDraftId) {
        editor.setStatus('● 空白草稿');
        return;
      }
      const bits = [];
      if (serverDraftId) {
        bits.push('服务器草稿 #' + serverDraftId);
        if (serverSavedAt) bits.push('上次保存 ' + timeSince(serverSavedAt));
      } else if (localSavedAt) {
        bits.push('本地自动保存 ' + timeSince(localSavedAt));
      }
      const lvl = dirty ? 'dirty' : 'saved';
      const prefix = dirty ? '编辑中 · ' : '● ';
      editor.setStatus(prefix + bits.join(' · '), lvl);
    }

    function startTicker() {
      if (statusTicker) return;
      statusTicker = setInterval(updateStatus, 15000);
    }

    async function submit(payload, action) {
      if (action === 'draft') {
        await saveAsDraft(payload);
        return { id: serverDraftId };
      }
      return publish(payload);
    }

    async function saveAsDraft(payload) {
      if (serverDraftId) {
        const patch = {
          body: payload.body,
          images: payload.images,
          status: 0,
        };
        if (payload.categoryId) patch.category_id = payload.categoryId;
        else patch.clear_category = true;
        const post = await BlinkAPI.patch('/api/posts/' + encodeURIComponent(serverDraftId), patch);
        afterServerSave(payload, post);
        return post;
      }
      const body = {
        body: payload.body,
        images: payload.images,
        draft: true,
      };
      if (payload.categoryId) body.category_id = payload.categoryId;
      const post = await BlinkAPI.post('/api/posts', body);
      serverDraftId = String(post.id);
      afterServerSave(payload, post);
      return post;
    }

    async function publish(payload) {
      let post;
      if (serverDraftId) {
        const patch = {
          body: payload.body,
          images: payload.images,
          status: 1,
        };
        if (payload.categoryId) patch.category_id = payload.categoryId;
        else patch.clear_category = true;
        post = await BlinkAPI.patch('/api/posts/' + encodeURIComponent(serverDraftId), patch);
      } else {
        const body = {
          body: payload.body,
          images: payload.images,
          draft: false,
        };
        if (payload.categoryId) body.category_id = payload.categoryId;
        post = await BlinkAPI.post('/api/posts', body);
      }
      clearDraft();
      dirty = false;
      serverDraftId = null;
      window.removeEventListener('beforeunload', beforeUnloadHandler);
      showPublishedDialog(post);
      return post;
    }

    function showPublishedDialog(post) {
      const postUrl = '/web/post.html?id=' + encodeURIComponent(post.id);
      const backdrop = el('div', { class: 'pe-dialog-backdrop', role: 'dialog', 'aria-modal': 'true' });
      const viewBtn = el('a', { class: 'btn btn-primary', href: postUrl }, '查看帖子');
      const moreBtn = el('button', { type: 'button', class: 'btn btn-secondary' }, '继续发下一条');
      const feedBtn = el('a', { class: 'btn btn-ghost', href: '/web/feed.html' }, '回帖子流');
      const box = el('div', { class: 'pe-dialog' }, [
        el('h3', {}, '发布成功 🎉'),
        el('p', {}, '帖子 #' + post.id + ' 已发布。'),
        el('div', { class: 'btn-row' }, [viewBtn, moreBtn, feedBtn]),
      ]);
      backdrop.appendChild(box);
      document.body.appendChild(backdrop);

      function close() { backdrop.remove(); }
      moreBtn.addEventListener('click', () => {
        close();
        editor.setValue({ body: '', images: [], categoryId: '' });
        editor.setOk('');
        editor.setError('');
        window.addEventListener('beforeunload', beforeUnloadHandler);
        localSavedAt = null;
        serverSavedAt = null;
        dirty = false;
        updateStatus();
        editor.focus();
      });
      backdrop.addEventListener('click', (e) => {
        if (e.target === backdrop) close();
      });
      document.addEventListener('keydown', function escClose(e) {
        if (e.key === 'Escape') {
          close();
          document.removeEventListener('keydown', escClose);
        }
      });
      setTimeout(() => viewBtn.focus(), 50);
    }

    function afterServerSave(payload, post) {
      serverSavedAt = Date.now();
      localSavedAt = serverSavedAt;
      dirty = false;
      writeDraft({
        body: payload.body,
        images: payload.images,
        categoryId: payload.categoryId,
        serverDraftId,
        ts: serverSavedAt,
      });
      editor.setOk('已保存到服务器 · 草稿 #' + (post && post.id));
      updateStatus();
    }

    // ---------- Restore banner ----------
    const stored = readDraft();
    const hasContent = stored && (
      (stored.body && stored.body.trim()) || (Array.isArray(stored.images) && stored.images.length)
    );
    if (hasContent) {
      const banner = el('div', { class: 'banner-warn' }, [
        el('strong', {}, '检测到未发布草稿'),
        el('p', { class: 'field-hint', style: { margin: '0.3rem 0 0.7rem' } },
          [
            (stored.body ? stored.body.length : 0) + ' 字符',
            (stored.images ? stored.images.length : 0) + ' 张图',
            stored.serverDraftId ? '已保存到服务器 #' + stored.serverDraftId : '',
            stored.ts ? '上次编辑 ' + new Date(stored.ts).toLocaleString() : '',
          ].filter(Boolean).join(' · ')
        ),
        el('div', { class: 'btn-row' }, [
          el('button', {
            type: 'button', class: 'btn btn-primary btn-sm',
            onClick: () => {
              editor.setValue({
                body: stored.body || '',
                images: Array.isArray(stored.images) ? stored.images : [],
                categoryId: stored.categoryId || '',
              });
              serverDraftId = stored.serverDraftId || null;
              localSavedAt = stored.ts || Date.now();
              serverSavedAt = serverDraftId ? localSavedAt : null;
              dirty = false;
              banner.remove();
              updateStatus();
            },
          }, '恢复草稿'),
          el('button', {
            type: 'button', class: 'btn btn-secondary btn-sm',
            onClick: () => {
              clearDraft();
              banner.remove();
            },
          }, '丢弃'),
        ]),
      ]);
      bannerMount.appendChild(banner);
    }

    // ---------- beforeunload ----------
    function beforeUnloadHandler(e) {
      const v = editor.getValue();
      if (dirty && ((v.body && v.body.trim()) || v.images.length)) {
        e.preventDefault();
        e.returnValue = '';
      }
    }
    window.addEventListener('beforeunload', beforeUnloadHandler);

    // ---------- Categories ----------
    const cats = await loadCategories();
    editor.setCategories(cats);

    startTicker();
    updateStatus();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
