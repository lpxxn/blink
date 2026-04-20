/* Edit-post page: reuses BlinkPostEditor (mode='edit') + moderation banners. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkPostEditor, BlinkMD } = window;

  const params = new URLSearchParams(location.search);
  const postId = params.get('id');

  const msgEl = () => document.getElementById('msg');

  function showErr(t) {
    const m = msgEl();
    if (!m) return;
    m.className = 'err';
    m.textContent = t || '';
  }

  function showOk(t) {
    const m = msgEl();
    if (!m) return;
    m.className = 'ok';
    m.textContent = t || '';
  }

  // ---------- Moderation banners (page-level, domain-specific) ----------
  function updateModerationUI(p) {
    const removed = document.getElementById('removed-banner');
    const flagged = document.getElementById('flagged-banner');
    if (removed) removed.hidden = true;
    if (flagged) flagged.hidden = true;
    if (!p) return;
    const mf = p.moderation_flag != null ? Number(p.moderation_flag) : 0;
    const note = (p.moderation_note && String(p.moderation_note).trim())
      ? BlinkMD.formatModerationNote(p.moderation_note) : '（未填写）';

    if (mf === 2 && removed) {
      removed.hidden = false;
      document.getElementById('mod-reason').textContent = note;
      const pending = document.getElementById('appeal-pending');
      const actions = document.getElementById('appeal-actions');
      const inAppeal = p.appeal_status === 1;
      pending.hidden = !inAppeal;
      actions.hidden = inAppeal;
    } else if (mf === 1 && flagged) {
      flagged.hidden = false;
      document.getElementById('flagged-mod-reason').textContent = note;
      const fp = document.getElementById('flagged-appeal-pending');
      const fa = document.getElementById('flagged-appeal-actions');
      const inAppeal = p.appeal_status === 1;
      fp.hidden = !inAppeal;
      fa.hidden = inAppeal;
    }
  }

  async function submitModerationRequest(kind, message) {
    showErr('');
    if (kind === 'appeal' && !message) {
      showErr('提交申诉时请填写说明');
      return;
    }
    try {
      const p = await BlinkAPI.post(
        '/api/posts/' + encodeURIComponent(postId) + '/moderation_request',
        { kind, message }
      );
      showOk(kind === 'appeal' ? '申诉已提交' : '已申请复核');
      updateModerationUI(p);
    } catch (e) {
      showErr(e && e.message ? e.message : String(e));
    }
  }

  function wireModerationButtons() {
    document.getElementById('btn-appeal').addEventListener('click', () => {
      const msg = document.getElementById('appeal-msg').value.trim();
      submitModerationRequest('appeal', msg);
    });
    document.getElementById('btn-resubmit').addEventListener('click', () => {
      const msg = document.getElementById('appeal-msg').value.trim();
      submitModerationRequest('resubmit', msg);
    });
    document.getElementById('btn-flagged-restore').addEventListener('click', () => {
      const msg = document.getElementById('flagged-restore-msg').value.trim();
      if (!msg) {
        showErr('申请恢复时请填写理由');
        return;
      }
      submitModerationRequest('resubmit', msg);
    });
  }

  // ---------- Image upload with progress (XHR) ----------
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
        const errMsg = (parsed && parsed.error) || ('上传失败 (' + xhr.status + ')');
        reject(new BlinkAPI.Error(xhr.status, errMsg, parsed));
      };
      xhr.onerror = () => reject(new BlinkAPI.Error(0, '网络异常，请稍后重试'));
      const fd = new FormData();
      fd.append('file', file);
      xhr.send(fd);
    });
  }

  // ---------- Categories ----------
  async function loadCategories() {
    try {
      const d = await BlinkAPI.get('/api/categories');
      return (d && d.categories) || [];
    } catch (_) { return []; }
  }

  // ---------- Init ----------
  async function init() {
    if (!postId) {
      showErr('缺少 id');
      return;
    }
    wireModerationButtons();

    let post;
    try {
      post = await BlinkAPI.get('/api/posts/' + encodeURIComponent(postId));
    } catch (err) {
      if (err && err.status === 404) showErr('帖子不存在或无权编辑');
      else if (err && err.status === 401) showErr('请先登录');
      else showErr(err && err.message ? err.message : String(err));
      return;
    }

    const mount = document.getElementById('editor-mount');
    if (!mount) return;
    let dirty = false;

    const editor = BlinkPostEditor.mount(mount, {
      mode: 'edit',
      showDelete: true,
      initial: {
        body: post.body || '',
        images: Array.isArray(post.images) ? post.images : [],
        categoryId: post.category_id != null ? String(post.category_id) : '',
        status: post.status != null ? Number(post.status) : 1,
        postId: post.id,
      },
      onImageUpload: uploadImage,
      onChange: () => { dirty = true; },
      onSubmit: async (payload, action) => {
        showErr('');
        const patch = {
          body: payload.body,
          images: payload.images,
          status: action === 'publish' ? 1 : payload.status,
        };
        if (payload.categoryId) patch.category_id = payload.categoryId;
        else patch.clear_category = true;

        const saved = await BlinkAPI.patch(
          '/api/posts/' + encodeURIComponent(postId),
          patch
        );
        if (saved && saved.status != null) {
          editor.setValue({ status: saved.status });
        }
        updateModerationUI(saved);
        dirty = false;
        showOk(action === 'publish' ? '已发布，帖子流可见' : '已保存');
        return saved;
      },
      onError: (err) => {
        if (err && err.body && Array.isArray(err.body.sensitive_words) && err.body.sensitive_words.length) {
          const words = err.body.sensitive_words;
          editor.setSensitiveHits(words);
          editor.setError('内容包含敏感词：' + words.join('、') + '（已在预览里标黄）');
          showErr('内容包含敏感词：' + words.join('、'));
          return true;
        }
        showErr(err && err.message ? err.message : String(err));
        return false;
      },
      onDelete: async () => {
        showErr('');
        try {
          await BlinkAPI.del('/api/posts/' + encodeURIComponent(postId));
          window.removeEventListener('beforeunload', beforeUnloadHandler);
          window.location.href = '/web/mine.html';
        } catch (err) {
          showErr(err && err.message ? err.message : String(err));
        }
      },
    });

    const cats = await loadCategories();
    editor.setCategories(cats);
    editor.setValue({
      categoryId: post.category_id != null ? String(post.category_id) : '',
      status: post.status != null ? Number(post.status) : 1,
    });

    updateModerationUI(post);
    document.getElementById('form-card').hidden = false;

    // beforeunload: warn if unsaved changes
    function beforeUnloadHandler(e) {
      if (dirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    }
    window.addEventListener('beforeunload', beforeUnloadHandler);
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
