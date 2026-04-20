/**
 * Admin — sensitive words.
 *
 * Adds page-based pagination, a page-local search, single add + batch add
 * (one per line), and per-row enable/disable/delete actions through
 * BlinkModal instead of prompt/confirm.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;

  const LIMIT = 50;

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  async function mount(container, ctx) {
    const state = { offset: 0, total: 0, words: [] };

    const errEl = el('p', { class: 'err', role: 'alert' });
    const okEl = el('p', { class: 'ok' });

    const newInput = el('input', {
      type: 'text',
      placeholder: '例如：spam',
      'aria-label': '新增敏感词',
    });
    const addBtn = el('button', {
      type: 'button', class: 'btn btn-primary btn-sm',
      onClick: () => addOne(),
    }, '添加');
    const batchBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => batchAdd(),
    }, '批量导入');

    const addCard = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '新增敏感词'),
      el('p', { class: 'admin-subtitle' }, '自动 trim + lower，修改后实时广播到所有实例。'),
      el('div', { class: 'admin-form-row' }, [newInput, addBtn, batchBtn]),
    ]);

    const searchInput = el('input', {
      type: 'search',
      placeholder: '在当前页内搜索…',
      'aria-label': '页内搜索',
    });
    const reloadBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => load(),
    }, '刷新');

    const toolbar = el('div', { class: 'admin-toolbar' }, [
      searchInput, el('span', { class: 'admin-toolbar-spacer' }), reloadBtn,
    ]);

    const tbody = el('tbody');
    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', {}, 'ID'),
          el('th', {}, '词'),
          el('th', {}, '启用'),
          el('th', {}, '更新时间'),
          el('th', { style: 'text-align:right' }, '操作'),
        ])),
        tbody,
      ]),
    ]);

    const pager = el('div', { class: 'admin-pager' });
    const pageMeta = el('span', {});
    const prevBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { if (state.offset > 0) { state.offset = Math.max(0, state.offset - LIMIT); load(); } },
    }, '上一页');
    const nextBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => { if (state.offset + LIMIT < state.total) { state.offset += LIMIT; load(); } },
    }, '下一页');
    pager.appendChild(pageMeta);
    pager.appendChild(el('span', { class: 'admin-pager-spacer' }));
    pager.appendChild(prevBtn);
    pager.appendChild(nextBtn);

    container.appendChild(errEl);
    container.appendChild(okEl);
    container.appendChild(addCard);
    container.appendChild(toolbar);
    container.appendChild(tableWrap);
    container.appendChild(pager);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }
    function showOk(msg) {
      okEl.textContent = msg || '';
    }

    function render() {
      clear(tbody);
      const q = (searchInput.value || '').trim().toLowerCase();
      const items = q
        ? state.words.filter((w) => (w.word || '').toLowerCase().includes(q) || String(w.id).includes(q))
        : state.words;

      if (items.length === 0) {
        tbody.appendChild(el('tr', {}, el('td', {
          colspan: '5', class: 'admin-empty',
        }, state.words.length ? '没有匹配的敏感词' : '词表为空')));
      } else {
        items.forEach((w) => tbody.appendChild(renderRow(w)));
      }

      const end = Math.min(state.offset + LIMIT, state.total);
      pageMeta.textContent = state.total
        ? '共 ' + state.total + ' 条 · 显示 ' + (state.offset + 1) + '–' + end
        : '暂无数据';
      prevBtn.disabled = state.offset <= 0;
      nextBtn.disabled = state.offset + LIMIT >= state.total;
    }

    function renderRow(w) {
      const tr = el('tr');
      const enabled = !!w.enabled;
      tr.appendChild(el('td', { class: 'mono' }, String(w.id)));
      tr.appendChild(el('td', {}, w.word || ''));
      tr.appendChild(el('td', {}, el('span', {
        class: 'chip ' + (enabled ? 'chip-ok' : 'chip-muted'),
      }, enabled ? '启用' : '停用')));
      tr.appendChild(el('td', {}, fmtTime(w.updated_at)));
      const actions = el('div', { class: 'row-actions' }, [
        el('button', {
          type: 'button', class: 'btn btn-secondary btn-sm',
          onClick: () => toggle(w),
        }, enabled ? '停用' : '启用'),
        el('button', {
          type: 'button', class: 'btn btn-danger btn-sm',
          onClick: () => remove(w),
        }, '删除'),
      ]);
      tr.appendChild(el('td', {}, actions));
      return tr;
    }

    async function toggle(w) {
      try {
        showErr(null);
        await AdminAPI.patchSensitiveWord(w.id, { enabled: !w.enabled });
        await load();
      } catch (err) { showErr(err); }
    }

    async function remove(w) {
      const ok = await Modal.confirm({
        title: '删除敏感词 #' + w.id,
        description: '删除后将不再拦截 “' + (w.word || '') + '”。此操作立即生效。',
        danger: true,
        confirmLabel: '删除',
      });
      if (!ok) return;
      try {
        showErr(null);
        await AdminAPI.deleteSensitiveWord(w.id);
        await load();
      } catch (err) { showErr(err); }
    }

    async function addOne() {
      showErr(null);
      showOk(null);
      const word = String(newInput.value || '').trim();
      if (!word) return showErr('请输入敏感词');
      try {
        await AdminAPI.createSensitiveWord(word);
        newInput.value = '';
        state.offset = 0;
        showOk('已添加：' + word);
        await load();
      } catch (err) { showErr(err); }
    }

    async function batchAdd() {
      const result = await Modal.open({
        title: '批量导入敏感词',
        description: '一行一个，自动 trim + lower。已存在的词会被服务端去重。',
        fields: [{
          name: 'words',
          label: '词表',
          type: 'textarea',
          required: true,
          rows: 10,
          placeholder: 'spam\nviagra\n…',
        }],
        confirmLabel: '导入',
      });
      if (!result) return;
      const lines = String(result.words || '').split(/\r?\n/).map((s) => s.trim()).filter(Boolean);
      if (lines.length === 0) return showErr('词表为空');

      showErr(null);
      showOk(null);
      let added = 0;
      let failed = 0;
      for (const w of lines) {
        try {
          await AdminAPI.createSensitiveWord(w);
          added += 1;
        } catch (_) {
          failed += 1;
        }
      }
      state.offset = 0;
      await load();
      showOk('已导入 ' + added + ' 条' + (failed ? '，' + failed + ' 条失败或已存在' : ''));
    }

    async function load() {
      showErr(null);
      try {
        const d = await AdminAPI.listSensitiveWords({ offset: state.offset, limit: LIMIT });
        state.words = d.words || [];
        state.total = typeof d.total === 'number' ? d.total : state.words.length;
        render();
      } catch (err) { showErr(err); }
    }

    searchInput.addEventListener('input', render);
    newInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') { e.preventDefault(); addOne(); }
    });

    await load();

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.sensitive = { mount };
})();
