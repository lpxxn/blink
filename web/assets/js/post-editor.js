/**
 * BlinkPostEditor.mount(container, config) — reusable Markdown composer.
 *
 * config = {
 *   mode: 'create' | 'edit',
 *   initial: { body, images, categoryId, status, postId },
 *   categories: [{ id, name }],
 *   onSubmit: async ({ body, images, categoryId, status }, action) => post,
 *     // action: 'publish' | 'draft' | 'save'
 *   onImageUpload: async (file, onProgress) => ({ url }),  // optional
 *   onDelete: async () => void,                            // edit mode
 *   showDelete: boolean,
 * }
 *
 * Returns: { getValue, setValue, focus, destroy, setError, setStatus, setImages }
 */
(function () {
  'use strict';

  const { el, clear } = window.BlinkUI;

  const MAX_IMAGE_BYTES = 5 * 1024 * 1024;
  const ALLOWED_MIME = new Set(['image/jpeg', 'image/png', 'image/webp', 'image/gif']);

  const TOOLS = [
    { key: 'bold',   label: 'B',   title: '加粗 (⌘/Ctrl+B)', kind: 'wrap',   args: ['**', '**', '加粗文本'] },
    { key: 'italic', label: 'I',   title: '斜体 (⌘/Ctrl+I)', kind: 'wrap',   args: ['*', '*', '斜体文本'] },
    { key: 'code',   label: '<>',  title: '行内代码 / 代码块', kind: 'code' },
    { key: 'heading',label: 'H',   title: '标题 (循环 H1 → H2 → H3 → 取消)', kind: 'heading' },
    { key: 'sep1',   kind: 'sep' },
    { key: 'link',   label: '🔗',  title: '插入链接 (⌘/Ctrl+K)', kind: 'link' },
    { key: 'image',  label: '🖼',  title: '插入图片（打开文件选择）', kind: 'image' },
    { key: 'sep2',   kind: 'sep' },
    { key: 'ul',     label: '•',   title: '无序列表', kind: 'prefix', args: ['- '] },
    { key: 'ol',     label: '1.',  title: '有序列表', kind: 'prefix', args: ['1. '] },
    { key: 'quote',  label: '❝',   title: '引用',     kind: 'prefix', args: ['> '] },
  ];

  function mount(container, config) {
    const cfg = config || {};
    const initial = cfg.initial || {};
    const mode = cfg.mode || 'create';

    const state = {
      body: initial.body != null ? String(initial.body) : '',
      images: Array.isArray(initial.images) ? initial.images.slice() : [],
      categoryId: initial.categoryId != null ? String(initial.categoryId) : '',
      status: initial.status != null ? Number(initial.status) : 1,
      postId: initial.postId != null ? String(initial.postId) : null,
    };
    /** pending uploads: [{ id, name, progress, error, file }] */
    const pending = [];
    let pendingSeq = 0;

    // ---------- DOM ----------
    const shell = el('div', { class: 'pe-shell' });

    const statusEl = el('span', { class: 'pe-status', role: 'status', 'aria-live': 'polite' }, '● 新草稿');
    const metaEl = el('div', { class: 'pe-meta' });
    const topbar = el('div', { class: 'pe-topbar' }, [statusEl, metaEl]);

    // Controls row: category + status (edit mode)
    const catSelect = el('select', { id: 'pe-cat' }, [
      el('option', { value: '' }, '（未分类）'),
    ]);
    const catField = el('div', { class: 'field' }, [
      el('label', { for: 'pe-cat' }, '分类'),
      catSelect,
    ]);

    const statusSelect = el('select', { id: 'pe-status' }, [
      el('option', { value: '0' }, '草稿（仅自己可见）'),
      el('option', { value: '1' }, '已发布（公开流可见）'),
      el('option', { value: '2' }, '已隐藏'),
    ]);
    const statusField = el('div', { class: 'field' }, [
      el('label', { for: 'pe-status' }, '状态'),
      statusSelect,
    ]);

    const controls = el('div', { class: 'pe-controls' }, [catField]);
    if (mode === 'edit') controls.appendChild(statusField);

    // Toolbar
    const toolbar = el('div', { class: 'pe-toolbar', role: 'toolbar', 'aria-label': 'Markdown 工具条' });
    const fileInput = el('input', {
      type: 'file',
      accept: 'image/png,image/jpeg,image/webp,image/gif',
      multiple: true,
      class: 'pe-sr-only',
      'aria-hidden': 'true',
      tabindex: '-1',
    });
    TOOLS.forEach((tool) => {
      if (tool.kind === 'sep') {
        toolbar.appendChild(el('span', { class: 'pe-tool-sep', 'aria-hidden': 'true' }));
        return;
      }
      const btn = el('button', {
        type: 'button',
        class: 'pe-tool',
        title: tool.title,
        'aria-label': tool.title,
        onClick: () => runTool(tool),
      }, tool.label);
      toolbar.appendChild(btn);
    });
    toolbar.appendChild(fileInput);

    // Body
    const textarea = el('textarea', {
      class: 'pe-textarea',
      id: 'pe-body',
      placeholder: '支持 Markdown · 拖拽或粘贴图片直接上传 · ⌘/Ctrl+Enter 发布',
      spellcheck: 'false',
      autocomplete: 'off',
    });
    textarea.value = state.body;
    const preview = el('div', { class: 'pe-preview markdown-body is-empty' });

    const tabEdit = el('button', { type: 'button', class: 'is-active' }, '写');
    const tabPreview = el('button', { type: 'button' }, '预览');
    const tabs = el('div', { class: 'pe-tabs', role: 'tablist' }, [tabEdit, tabPreview]);

    const body = el('div', { class: 'pe-body is-mode-edit' }, [
      el('div', {}, [
        el('label', { class: 'pe-pane-label', for: 'pe-body' }, '正文 (Markdown)'),
        textarea,
      ]),
      el('div', {}, [
        el('span', { class: 'pe-pane-label' }, '实时预览'),
        preview,
      ]),
    ]);

    function setMode(mode) {
      body.classList.remove('is-mode-edit', 'is-mode-preview');
      body.classList.add('is-mode-' + mode);
      tabEdit.classList.toggle('is-active', mode === 'edit');
      tabPreview.classList.toggle('is-active', mode === 'preview');
      if (mode === 'preview') renderPreview();
    }
    tabEdit.addEventListener('click', () => setMode('edit'));
    tabPreview.addEventListener('click', () => setMode('preview'));

    // Images panel
    const imgGrid = el('div', { class: 'pe-images', 'aria-live': 'polite' });
    const imgSection = el('div', {}, [
      el('span', { class: 'pe-pane-label' }, '已上传图片'),
      imgGrid,
    ]);

    // Error
    const errorEl = el('p', { class: 'pe-error err', role: 'alert', 'aria-live': 'assertive' });

    // Action buttons
    const publishBtn = el('button', { type: 'button', class: 'btn btn-primary' }, mode === 'edit' ? '发布' : '发布');
    const draftBtn = el('button', { type: 'button', class: 'btn btn-secondary' }, mode === 'edit' ? '保存' : '保存草稿');
    const viewBtn = mode === 'edit'
      ? el('a', { class: 'btn btn-ghost', href: state.postId ? `/web/post.html?id=${encodeURIComponent(state.postId)}` : '#' }, '查看帖子')
      : null;
    const deleteBtn = (mode === 'edit' && cfg.showDelete)
      ? el('button', { type: 'button', class: 'btn btn-ghost' }, '删除帖子')
      : null;

    const actions = el('div', { class: 'pe-actions' }, [
      publishBtn,
      draftBtn,
      el('span', { class: 'spacer' }),
      deleteBtn,
      viewBtn,
    ]);

    shell.append(topbar, controls, toolbar, tabs, body, imgSection, errorEl, actions);
    clear(container);
    container.appendChild(shell);

    // ---------- State renderers ----------
    function setStatus(text, level) {
      statusEl.textContent = text || '';
      statusEl.className = 'pe-status' + (level ? ' is-' + level : '');
    }

    function setError(msg) {
      errorEl.textContent = msg || '';
      errorEl.className = msg ? 'pe-error err' : 'pe-error';
    }

    function setOk(msg) {
      errorEl.textContent = msg || '';
      errorEl.className = msg ? 'pe-error ok' : 'pe-error';
    }

    function renderMeta() {
      const chars = state.body.length;
      const imgs = state.images.length;
      const limit = '每张 ≤ 5 MiB · 支持 JPG/PNG/WebP/GIF';
      metaEl.textContent = `字数 ${chars} · 图片 ${imgs} · ${limit}`;
    }

    let sensitiveHits = [];

    function renderPreview() {
      const md = state.body;
      if (!md.trim()) {
        preview.className = 'pe-preview markdown-body is-empty';
        preview.textContent = '（预览会在这里显示）';
        return;
      }
      preview.className = 'pe-preview markdown-body';
      preview.innerHTML = window.BlinkMD.parse(md);
      if (sensitiveHits.length) highlightSensitive(preview, sensitiveHits);
    }

    function highlightSensitive(root, words) {
      if (!words || !words.length) return;
      const esc = words
        .filter(Boolean)
        .map((w) => String(w).replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
        .filter((w) => w.length);
      if (!esc.length) return;
      const re = new RegExp(esc.join('|'), 'gi');
      const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
      const nodes = [];
      let n;
      while ((n = walker.nextNode())) nodes.push(n);
      for (const node of nodes) {
        const text = node.nodeValue;
        re.lastIndex = 0;
        if (!re.test(text)) continue;
        re.lastIndex = 0;
        const frag = document.createDocumentFragment();
        let last = 0;
        let m;
        while ((m = re.exec(text))) {
          if (m.index > last) frag.appendChild(document.createTextNode(text.slice(last, m.index)));
          const mark = document.createElement('mark');
          mark.className = 'sensitive-hit';
          mark.textContent = m[0];
          frag.appendChild(mark);
          last = m.index + m[0].length;
        }
        if (last < text.length) frag.appendChild(document.createTextNode(text.slice(last)));
        node.parentNode.replaceChild(frag, node);
      }
    }

    function setSensitiveHits(words) {
      sensitiveHits = Array.isArray(words) ? words.slice() : [];
      renderPreview();
    }

    function clearSensitiveHits() {
      if (!sensitiveHits.length) return;
      sensitiveHits = [];
      renderPreview();
    }

    function renderImages() {
      clear(imgGrid);
      state.images.forEach((url) => {
        const thumb = el('div', { class: 'pe-thumb' }, [
          el('img', { src: url, alt: '' }),
          el('button', {
            type: 'button',
            class: 'pe-thumb-remove',
            title: '删除此图（同时移除正文里的引用）',
            'aria-label': '删除图片',
            onClick: () => removeImage(url),
          }, '×'),
        ]);
        imgGrid.appendChild(thumb);
      });
      pending.forEach((p) => {
        if (p.error) {
          const thumb = el('div', { class: 'pe-thumb is-error', 'data-pending-id': String(p.id) }, [
            el('div', { class: 'pe-thumb-error-msg', title: p.error }, p.error),
            el('button', {
              type: 'button',
              class: 'pe-thumb-retry',
              onClick: () => retryPending(p.id),
            }, '重试'),
            el('button', {
              type: 'button',
              class: 'pe-thumb-remove',
              title: '移除',
              'aria-label': '移除',
              onClick: () => cancelPending(p.id),
            }, '×'),
          ]);
          imgGrid.appendChild(thumb);
        } else {
          const pct = Math.round((p.progress || 0) * 100);
          const thumb = el('div', { class: 'pe-thumb is-uploading', 'data-pending-id': String(p.id) }, [
            el('div', { class: 'pe-thumb-name' }, p.name || '上传中'),
            el('div', { class: 'pe-thumb-progress' }, [
              el('div', { class: 'pe-thumb-progress-bar', style: { width: pct + '%' } }),
            ]),
            el('div', {}, pct + '%'),
          ]);
          imgGrid.appendChild(thumb);
        }
      });
    }

    function updatePendingProgress(id, progress) {
      const p = pending.find((x) => x.id === id);
      if (!p) return;
      p.progress = progress;
      const node = imgGrid.querySelector('[data-pending-id="' + id + '"]');
      if (!node || node.classList.contains('is-error')) return;
      const bar = node.querySelector('.pe-thumb-progress-bar');
      const pct = Math.round(progress * 100);
      if (bar) bar.style.width = pct + '%';
      const lbl = node.lastElementChild;
      if (lbl) lbl.textContent = pct + '%';
    }

    function cancelPending(id) {
      const idx = pending.findIndex((x) => x.id === id);
      if (idx !== -1) pending.splice(idx, 1);
      renderImages();
    }

    function retryPending(id) {
      const p = pending.find((x) => x.id === id);
      if (!p) return;
      p.error = null;
      p.progress = 0;
      renderImages();
      runUpload(p);
    }

    async function runUpload(entry) {
      try {
        const res = await cfg.onImageUpload(entry.file, (progress) => {
          updatePendingProgress(entry.id, progress);
        });
        cancelPending(entry.id);
        if (res && res.url) {
          const alt = (entry.file && entry.file.name || '图片').replace(/\.[^.]+$/, '');
          addImage(res.url, alt);
          setOk('已上传 ' + (entry.file.name || '图片'));
        }
      } catch (err) {
        entry.error = err && err.message ? err.message : String(err);
        entry.progress = 0;
        renderImages();
        setError(entry.error);
      }
    }

    let changeTimer = null;
    function fireChange() {
      if (typeof cfg.onChange !== 'function') return;
      if (changeTimer) clearTimeout(changeTimer);
      changeTimer = setTimeout(() => {
        try {
          cfg.onChange({
            body: textarea.value,
            images: state.images.slice(),
            categoryId: state.categoryId,
            status: state.status,
          });
        } catch (e) {
          console.error('onChange threw', e);
        }
      }, 300);
    }

    function syncBodyToState() {
      state.body = textarea.value;
      renderMeta();
      renderPreview();
      fireChange();
    }

    // ---------- Image helpers ----------
    function insertImage(url, alt) {
      window.BlinkMD.insertImageMarkdown(textarea, url, alt);
      syncBodyToState();
    }

    function addImage(url, alt) {
      if (!url) return;
      if (state.images.indexOf(url) === -1) state.images.push(url);
      insertImage(url, alt || '图片');
      renderImages();
    }

    function removeImage(url) {
      const idx = state.images.indexOf(url);
      if (idx !== -1) state.images.splice(idx, 1);
      const esc = url.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const re = new RegExp('[ \\t]*!\\[[^\\]]*\\]\\(<?' + esc + '>?\\)[ \\t]*\\n?', 'g');
      const next = textarea.value.replace(re, '');
      if (next !== textarea.value) {
        textarea.value = next.replace(/\n{3,}/g, '\n\n');
      }
      renderImages();
      syncBodyToState();
    }

    function validateImageClient(file) {
      if (!file) return '未选择文件';
      if (!ALLOWED_MIME.has(file.type)) return `不支持的图片类型：${file.type || '未知'}`;
      if (file.size > MAX_IMAGE_BYTES) return `图片过大（${(file.size / 1024 / 1024).toFixed(2)} MiB > 5 MiB）`;
      return '';
    }

    function enqueueUpload(file) {
      const err = validateImageClient(file);
      if (err) {
        setError(err);
        return null;
      }
      if (typeof cfg.onImageUpload !== 'function') {
        setError('当前页面未配置图片上传');
        return null;
      }
      const entry = {
        id: ++pendingSeq,
        file,
        name: file.name || '未命名',
        progress: 0,
        error: null,
      };
      pending.push(entry);
      renderImages();
      runUpload(entry);
      return entry;
    }

    function enqueueUploads(files) {
      const arr = Array.isArray(files) ? files : Array.from(files || []);
      arr.forEach((f) => enqueueUpload(f));
    }

    // ---------- Toolbar actions ----------
    function wrap(before, after, placeholder) {
      const v = textarea.value;
      const s = textarea.selectionStart;
      const e = textarea.selectionEnd;
      const sel = v.slice(s, e) || placeholder || '';
      const next = v.slice(0, s) + before + sel + after + v.slice(e);
      textarea.value = next;
      const cStart = s + before.length;
      const cEnd = cStart + sel.length;
      textarea.focus();
      textarea.setSelectionRange(cStart, cEnd);
      syncBodyToState();
    }

    function prefixLines(prefix) {
      const v = textarea.value;
      const s = textarea.selectionStart;
      const e = textarea.selectionEnd;
      const lineStart = v.lastIndexOf('\n', s - 1) + 1;
      let lineEnd = v.indexOf('\n', e);
      if (lineEnd === -1) lineEnd = v.length;
      const segment = v.slice(lineStart, lineEnd);
      const lines = segment.split('\n');
      const allHas = lines.every((l) => l.startsWith(prefix));
      const next = allHas
        ? lines.map((l) => l.slice(prefix.length)).join('\n')
        : lines.map((l) => prefix + l).join('\n');
      textarea.value = v.slice(0, lineStart) + next + v.slice(lineEnd);
      const delta = next.length - segment.length;
      textarea.focus();
      textarea.setSelectionRange(lineStart, lineStart + next.length);
      syncBodyToState();
    }

    function cycleHeading() {
      const v = textarea.value;
      const s = textarea.selectionStart;
      const lineStart = v.lastIndexOf('\n', s - 1) + 1;
      let lineEnd = v.indexOf('\n', s);
      if (lineEnd === -1) lineEnd = v.length;
      const line = v.slice(lineStart, lineEnd);
      const m = line.match(/^(#{1,6})\s(.*)$/);
      let next;
      if (!m) {
        next = '# ' + line;
      } else {
        const level = m[1].length;
        if (level < 3) next = '#'.repeat(level + 1) + ' ' + m[2];
        else next = m[2];
      }
      textarea.value = v.slice(0, lineStart) + next + v.slice(lineEnd);
      const caret = lineStart + next.length;
      textarea.focus();
      textarea.setSelectionRange(caret, caret);
      syncBodyToState();
    }

    function toggleCode() {
      const v = textarea.value;
      const s = textarea.selectionStart;
      const e = textarea.selectionEnd;
      const sel = v.slice(s, e);
      if (sel.indexOf('\n') !== -1 || e === s) {
        const block = '\n```\n' + (sel || '在此写代码') + '\n```\n';
        const next = v.slice(0, s) + block + v.slice(e);
        textarea.value = next;
        const caret = s + 5;
        textarea.focus();
        textarea.setSelectionRange(caret, caret);
      } else {
        wrap('`', '`', '代码');
      }
      syncBodyToState();
    }

    function insertLink() {
      const sel = textarea.value.slice(textarea.selectionStart, textarea.selectionEnd);
      const url = window.prompt('链接地址', 'https://');
      if (!url) return;
      const text = sel || '链接文本';
      const inserted = '[' + text + '](' + url + ')';
      const s = textarea.selectionStart;
      const e = textarea.selectionEnd;
      textarea.value = textarea.value.slice(0, s) + inserted + textarea.value.slice(e);
      const caret = s + inserted.length;
      textarea.focus();
      textarea.setSelectionRange(caret, caret);
      syncBodyToState();
    }

    function runTool(tool) {
      switch (tool.kind) {
        case 'wrap':    return wrap.apply(null, tool.args);
        case 'prefix':  return prefixLines.apply(null, tool.args);
        case 'heading': return cycleHeading();
        case 'code':    return toggleCode();
        case 'link':    return insertLink();
        case 'image':   return fileInput.click();
        default: return null;
      }
    }

    // ---------- Events ----------
    let previewTimer = null;
    textarea.addEventListener('input', () => {
      state.body = textarea.value;
      renderMeta();
      if (previewTimer) clearTimeout(previewTimer);
      previewTimer = setTimeout(renderPreview, 120);
      fireChange();
    });

    textarea.addEventListener('keydown', (e) => {
      const mod = e.metaKey || e.ctrlKey;
      if (!mod) return;
      const k = e.key.toLowerCase();
      if (k === 'b') { e.preventDefault(); wrap('**', '**', '加粗文本'); }
      else if (k === 'i') { e.preventDefault(); wrap('*', '*', '斜体文本'); }
      else if (k === 'k') { e.preventDefault(); insertLink(); }
      else if (k === 'enter') { e.preventDefault(); handlePublish(); }
    });

    catSelect.addEventListener('change', () => {
      state.categoryId = catSelect.value;
      fireChange();
    });
    statusSelect.addEventListener('change', () => {
      state.status = parseInt(statusSelect.value, 10);
      fireChange();
    });

    fileInput.addEventListener('change', () => {
      const files = Array.from(fileInput.files || []);
      fileInput.value = '';
      if (files.length) enqueueUploads(files);
    });

    // Drag & drop (whole editor shell)
    let dragDepth = 0;
    function hasFiles(e) {
      const dt = e.dataTransfer;
      if (!dt) return false;
      if (dt.types && Array.from(dt.types).indexOf('Files') !== -1) return true;
      return false;
    }
    shell.addEventListener('dragenter', (e) => {
      if (!hasFiles(e)) return;
      e.preventDefault();
      dragDepth += 1;
      shell.classList.add('is-dragover');
    });
    shell.addEventListener('dragover', (e) => {
      if (!hasFiles(e)) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = 'copy';
    });
    shell.addEventListener('dragleave', (e) => {
      if (!hasFiles(e)) return;
      dragDepth = Math.max(0, dragDepth - 1);
      if (dragDepth === 0) shell.classList.remove('is-dragover');
    });
    shell.addEventListener('drop', (e) => {
      if (!hasFiles(e)) return;
      e.preventDefault();
      dragDepth = 0;
      shell.classList.remove('is-dragover');
      const files = Array.from(e.dataTransfer.files || []).filter((f) => f.type.startsWith('image/'));
      if (files.length) enqueueUploads(files);
    });

    // Paste images
    textarea.addEventListener('paste', (e) => {
      const items = e.clipboardData && e.clipboardData.items;
      if (!items) return;
      const files = [];
      for (let i = 0; i < items.length; i++) {
        const it = items[i];
        if (it.kind === 'file') {
          const f = it.getAsFile();
          if (f && f.type.startsWith('image/')) files.push(f);
        }
      }
      if (files.length) {
        e.preventDefault();
        enqueueUploads(files);
      }
    });

    // ---------- Submit ----------
    async function submitWith(action) {
      setError('');
      if (pending.some((p) => !p.error)) {
        setError('等图片上传完成后再提交');
        return;
      }
      if (!textarea.value.trim() && state.images.length === 0) {
        setError('正文不能为空');
        textarea.focus();
        return;
      }
      publishBtn.disabled = true;
      draftBtn.disabled = true;
      try {
        const payload = {
          body: textarea.value,
          images: state.images.slice(),
          categoryId: state.categoryId,
          status: state.status,
        };
        clearSensitiveHits();
        const post = await cfg.onSubmit(payload, action);
        if (post && post.id != null) setStatus('● 已保存 #' + post.id, 'saved');
      } catch (err) {
        let handled = false;
        if (typeof cfg.onError === 'function') {
          try { handled = !!cfg.onError(err, action); } catch (_) { /* ignore */ }
        }
        if (!handled) setError(err && err.message ? err.message : String(err));
      } finally {
        publishBtn.disabled = false;
        draftBtn.disabled = false;
      }
    }

    function handlePublish() { return submitWith('publish'); }
    function handleDraft() {
      return submitWith(mode === 'edit' ? 'save' : 'draft');
    }

    publishBtn.addEventListener('click', handlePublish);
    draftBtn.addEventListener('click', handleDraft);

    if (deleteBtn && typeof cfg.onDelete === 'function') {
      deleteBtn.addEventListener('click', async () => {
        if (!window.confirm('确定删除该帖子？不可恢复。')) return;
        try {
          await cfg.onDelete();
        } catch (err) {
          setError(err && err.message ? err.message : String(err));
        }
      });
    }

    // ---------- Category population ----------
    function setCategories(list) {
      const keep = state.categoryId;
      while (catSelect.options.length > 1) catSelect.remove(1);
      (list || []).forEach((c) => {
        const opt = el('option', { value: String(c.id) }, c.name);
        catSelect.appendChild(opt);
      });
      if (keep) catSelect.value = keep;
    }

    if (Array.isArray(cfg.categories) && cfg.categories.length) {
      setCategories(cfg.categories);
    }

    if (state.status != null) statusSelect.value = String(state.status);

    // Initial render
    renderMeta();
    renderPreview();
    renderImages();

    return {
      el: shell,
      getValue: () => ({
        body: textarea.value,
        images: state.images.slice(),
        categoryId: state.categoryId,
        status: state.status,
      }),
      setValue: (partial) => {
        if (partial.body != null) {
          textarea.value = partial.body;
          state.body = partial.body;
        }
        if (Array.isArray(partial.images)) state.images = partial.images.slice();
        if (partial.categoryId != null) {
          state.categoryId = String(partial.categoryId);
          catSelect.value = state.categoryId;
        }
        if (partial.status != null) {
          state.status = Number(partial.status);
          statusSelect.value = String(state.status);
        }
        if (partial.postId != null) state.postId = String(partial.postId);
        // Programmatic update → don't mark as dirty / fire onChange
        renderImages();
        renderMeta();
        renderPreview();
      },
      focus: () => textarea.focus(),
      setError,
      setOk,
      setStatus,
      setCategories,
      setSensitiveHits,
      clearSensitiveHits,
      addImage,
      removeImage,
      getImages: () => state.images.slice(),
      destroy: () => clear(container),
    };
  }

  window.BlinkPostEditor = { mount };
})();
