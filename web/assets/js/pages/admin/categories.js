/**
 * Admin — categories CRUD.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;

  async function mount(container, ctx) {
    const errEl = el('p', { class: 'err', role: 'alert' });
    const okEl = el('p', { class: 'ok' });

    const slugInput = el('input', { type: 'text', placeholder: 'slug，如 tech', 'aria-label': 'slug' });
    const nameInput = el('input', { type: 'text', placeholder: '显示名称', 'aria-label': '名称' });
    const sortInput = el('input', { type: 'number', value: '0', 'aria-label': '排序' });
    const addBtn = el('button', { type: 'button', class: 'btn btn-primary btn-sm' }, '添加');

    const addCard = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '新增分类'),
      el('div', { class: 'admin-form-row' }, [slugInput, nameInput, sortInput, addBtn]),
    ]);

    const tbody = el('tbody');
    const tableWrap = el('div', { class: 'admin-table-wrap' }, [
      el('table', { class: 'admin-table' }, [
        el('thead', {}, el('tr', {}, [
          el('th', { class: 'nowrap' }, 'ID'),
          el('th', {}, 'Slug'),
          el('th', {}, '名称'),
          el('th', { class: 'nowrap' }, '排序'),
          el('th', { class: 'col-actions' }, '操作'),
        ])),
        tbody,
      ]),
    ]);

    container.appendChild(errEl);
    container.appendChild(okEl);
    container.appendChild(addCard);
    container.appendChild(tableWrap);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
      if (err) okEl.textContent = '';
    }

    function showOk(msg) {
      okEl.textContent = msg || '';
      if (msg) errEl.textContent = '';
    }

    function render(list) {
      clear(tbody);
      if (!list.length) {
        tbody.appendChild(el('tr', {}, el('td', { colspan: '5', class: 'admin-empty' }, '暂无分类')));
        return;
      }
      list.forEach((c) => tbody.appendChild(renderRow(c)));
    }

    function renderRow(c) {
      const tr = el('tr');
      tr.appendChild(el('td', { class: 'mono nowrap' }, String(c.id)));
      tr.appendChild(el('td', { class: 'mono' }, c.slug || '—'));
      tr.appendChild(el('td', {}, c.name || '—'));
      tr.appendChild(el('td', { class: 'nowrap' }, String(c.sort_order != null ? c.sort_order : 0)));

      const actions = el('div', { class: 'row-actions' });
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-secondary btn-sm',
        onClick: () => editCategory(c),
      }, '编辑'));
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-danger btn-sm',
        onClick: () => deleteCategory(c),
      }, '删除'));
      tr.appendChild(el('td', { class: 'col-actions' }, actions));
      return tr;
    }

    async function load() {
      showErr(null);
      try {
        const d = await AdminAPI.listCategories();
        render(d.categories || []);
      } catch (err) { showErr(err); }
    }

    async function addCategory() {
      const slug = (slugInput.value || '').trim();
      const name = (nameInput.value || '').trim();
      const sort = parseInt(sortInput.value, 10) || 0;
      if (!slug || !name) {
        showErr(new Error('请填写 slug 与名称'));
        return;
      }
      try {
        showErr(null);
        await AdminAPI.createCategory({ slug, name, sort_order: sort });
        slugInput.value = '';
        nameInput.value = '';
        sortInput.value = '0';
        showOk('已添加分类');
        await load();
      } catch (err) { showErr(err); }
    }

    async function editCategory(c) {
      const result = await Modal.open({
        title: '编辑分类 #' + c.id,
        fields: [
          { name: 'slug', label: 'Slug', type: 'text', value: c.slug || '', required: true },
          { name: 'name', label: '名称', type: 'text', value: c.name || '', required: true },
          { name: 'sort_order', label: '排序', type: 'number', value: String(c.sort_order != null ? c.sort_order : 0) },
        ],
        confirmLabel: '保存',
      });
      if (!result) return;
      try {
        showErr(null);
        await AdminAPI.patchCategory(c.id, {
          slug: result.slug,
          name: result.name,
          sort_order: parseInt(result.sort_order, 10) || 0,
        });
        showOk('已更新');
        await load();
      } catch (err) { showErr(err); }
    }

    async function deleteCategory(c) {
      const ok = await Modal.confirm({
        title: '删除分类 #' + c.id,
        description: '软删除后前台将不再展示该分类。确定继续？',
        danger: true,
        confirmLabel: '删除',
      });
      if (!ok) return;
      try {
        showErr(null);
        await AdminAPI.deleteCategory(c.id);
        showOk('已删除');
        await load();
      } catch (err) { showErr(err); }
    }

    addBtn.addEventListener('click', addCategory);
    await load();
    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.categories = { mount };
})();
