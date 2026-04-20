/**
 * BlinkModal — promise-based modal dialogs.
 *
 *   BlinkModal.open({
 *     title: '下架帖子 #123',
 *     description: '作者会收到通知，内容在公开流中立即消失。',
 *     fields: [
 *       { name: 'note', label: '下架理由', type: 'textarea', required: true,
 *         maxLength: 500, placeholder: '…', rows: 4 },
 *     ],
 *     confirmLabel: '下架',
 *     danger: true,
 *   }).then((values) => {
 *     if (!values) return;  // user cancelled
 *     // values.note
 *   });
 *
 * Sugar:
 *   BlinkModal.confirm({ title, description, danger, confirmLabel }) → Promise<boolean>
 *   BlinkModal.prompt({ title, label, value, required }) → Promise<string|null>
 *   BlinkModal.alert({ title, description }) → Promise<void>
 *
 * Field types: text, textarea, select, radio, checkbox, password.
 */
(function () {
  'use strict';

  const { el, clear } = window.BlinkUI;

  let stack = [];

  function close(modal, result) {
    if (!modal || modal._closed) return;
    modal._closed = true;
    modal.backdrop.remove();
    document.removeEventListener('keydown', modal._keydown, true);
    stack = stack.filter((m) => m !== modal);
    if (modal.prevFocus && typeof modal.prevFocus.focus === 'function') {
      try { modal.prevFocus.focus(); } catch (_) { /* ignore */ }
    }
    if (modal._resolve) modal._resolve(result);
  }

  function renderField(field, onChange) {
    const id = 'modal-f-' + Math.random().toString(36).slice(2, 9);
    const wrap = el('div', { class: 'modal-field' });

    if (field.type === 'checkbox') {
      const chk = el('input', { type: 'checkbox', id });
      if (field.value) chk.checked = true;
      chk.addEventListener('change', () => onChange(field.name, chk.checked));
      onChange(field.name, !!field.value);
      wrap.appendChild(el('label', { class: 'modal-checkbox', for: id }, [
        chk,
        el('span', {}, field.text || field.label || ''),
      ]));
      if (field.hint) wrap.appendChild(el('p', { class: 'modal-hint' }, field.hint));
      return { wrap, focusable: chk };
    }

    if (field.label) {
      wrap.appendChild(el('label', { for: id }, field.label + (field.required ? ' *' : '')));
    }

    if (field.type === 'select') {
      const sel = el('select', { id });
      (field.options || []).forEach((opt) => {
        const o = el('option', { value: String(opt.value) }, opt.label);
        if (String(opt.value) === String(field.value)) o.selected = true;
        sel.appendChild(o);
      });
      sel.addEventListener('change', () => onChange(field.name, sel.value));
      onChange(field.name, sel.value);
      wrap.appendChild(sel);
      if (field.hint) wrap.appendChild(el('p', { class: 'modal-hint' }, field.hint));
      return { wrap, focusable: sel };
    }

    if (field.type === 'radio') {
      const groupName = id;
      const grp = el('div', { class: 'modal-radio-group' });
      (field.options || []).forEach((opt) => {
        const rid = groupName + '-' + opt.value;
        const r = el('input', { type: 'radio', name: groupName, id: rid, value: String(opt.value) });
        if (String(opt.value) === String(field.value)) r.checked = true;
        r.addEventListener('change', () => {
          if (r.checked) onChange(field.name, r.value);
        });
        grp.appendChild(el('label', { class: 'modal-radio', for: rid }, [
          r,
          el('span', {}, [
            el('strong', {}, opt.label),
            opt.hint ? el('p', { class: 'modal-radio-hint' }, opt.hint) : null,
          ]),
        ]));
      });
      onChange(field.name, field.value);
      wrap.appendChild(grp);
      if (field.hint) wrap.appendChild(el('p', { class: 'modal-hint' }, field.hint));
      return { wrap, focusable: grp.querySelector('input') };
    }

    if (field.type === 'textarea') {
      const ta = el('textarea', {
        id,
        rows: String(field.rows || 4),
        placeholder: field.placeholder || '',
      });
      if (field.maxLength) ta.maxLength = field.maxLength;
      ta.value = field.value != null ? String(field.value) : '';
      let counter = null;
      if (field.maxLength) {
        counter = el('p', { class: 'modal-hint modal-counter' }, '0 / ' + field.maxLength);
      }
      const updateCounter = () => {
        if (counter) counter.textContent = ta.value.length + ' / ' + field.maxLength;
      };
      ta.addEventListener('input', () => { onChange(field.name, ta.value); updateCounter(); });
      onChange(field.name, ta.value);
      updateCounter();
      wrap.appendChild(ta);
      if (field.hint) wrap.appendChild(el('p', { class: 'modal-hint' }, field.hint));
      if (counter) wrap.appendChild(counter);
      return { wrap, focusable: ta };
    }

    // text / password / number
    const input = el('input', {
      id,
      type: field.type || 'text',
      placeholder: field.placeholder || '',
      autocomplete: field.autocomplete || 'off',
    });
    input.value = field.value != null ? String(field.value) : '';
    if (field.maxLength) input.maxLength = field.maxLength;
    input.addEventListener('input', () => onChange(field.name, input.value));
    onChange(field.name, input.value);
    wrap.appendChild(input);
    if (field.hint) wrap.appendChild(el('p', { class: 'modal-hint' }, field.hint));
    return { wrap, focusable: input };
  }

  function open(opts) {
    const cfg = Object.assign({
      title: '',
      description: '',
      fields: [],
      confirmLabel: '确定',
      cancelLabel: '取消',
      danger: false,
      hideCancel: false,
    }, opts || {});

    return new Promise((resolve) => {
      const values = {};
      function handleChange(name, val) { values[name] = val; }

      function validate() {
        for (const f of cfg.fields) {
          if (!f.required) continue;
          const v = values[f.name];
          if (v == null) return f.label + ' 不能为空';
          if (typeof v === 'string' && !v.trim()) return (f.label || f.name) + ' 不能为空';
          if (f.type === 'checkbox' && !v) return '请勾选 “' + (f.text || f.label) + '”';
          if (f.minLength && typeof v === 'string' && v.trim().length < f.minLength) {
            return (f.label || f.name) + ' 至少 ' + f.minLength + ' 字';
          }
        }
        return null;
      }

      const errorEl = el('p', { class: 'modal-error', role: 'alert' });
      const body = el('div', { class: 'modal-body' });
      if (cfg.description) {
        body.appendChild(el('p', { class: 'modal-description' }, cfg.description));
      }

      const focusables = [];
      cfg.fields.forEach((field) => {
        const { wrap, focusable } = renderField(field, handleChange);
        body.appendChild(wrap);
        if (focusable) focusables.push(focusable);
      });
      body.appendChild(errorEl);

      const cancelBtn = cfg.hideCancel ? null : el('button', {
        type: 'button', class: 'btn btn-ghost',
        onClick: () => close(modal, null),
      }, cfg.cancelLabel);

      const confirmBtn = el('button', {
        type: 'button',
        class: 'btn ' + (cfg.danger ? 'btn-danger' : 'btn-primary'),
        onClick: async () => {
          const err = validate();
          if (err) {
            errorEl.textContent = err;
            return;
          }
          errorEl.textContent = '';
          if (typeof cfg.onSubmit === 'function') {
            confirmBtn.disabled = true;
            if (cancelBtn) cancelBtn.disabled = true;
            try {
              const result = await cfg.onSubmit(Object.assign({}, values));
              close(modal, result === undefined ? values : result);
            } catch (e) {
              errorEl.textContent = (e && e.message ? e.message : String(e));
              confirmBtn.disabled = false;
              if (cancelBtn) cancelBtn.disabled = false;
            }
          } else {
            close(modal, Object.assign({}, values));
          }
        },
      }, cfg.confirmLabel);

      const actions = el('div', { class: 'modal-actions' },
        cancelBtn ? [cancelBtn, confirmBtn] : [confirmBtn]);

      const dialog = el('div', {
        class: 'modal-dialog' + (cfg.danger ? ' is-danger' : ''),
        role: 'dialog', 'aria-modal': 'true', 'aria-labelledby': 'modal-title',
      }, [
        el('h3', { class: 'modal-title', id: 'modal-title' }, cfg.title),
        body,
        actions,
      ]);

      const backdrop = el('div', { class: 'modal-backdrop' }, dialog);
      backdrop.addEventListener('click', (e) => {
        if (e.target === backdrop) close(modal, null);
      });

      const modal = {
        backdrop,
        confirmBtn,
        cancelBtn,
        errorEl,
        prevFocus: document.activeElement,
        _resolve: resolve,
        _closed: false,
      };

      modal._keydown = (e) => {
        if (stack[stack.length - 1] !== modal) return;
        if (e.key === 'Escape') {
          e.preventDefault();
          close(modal, null);
        } else if (e.key === 'Enter') {
          const tag = (e.target && e.target.tagName) || '';
          if (tag === 'TEXTAREA' && !(e.metaKey || e.ctrlKey)) return;
          if (tag === 'SELECT') return;
          e.preventDefault();
          confirmBtn.click();
        }
      };
      document.addEventListener('keydown', modal._keydown, true);

      document.body.appendChild(backdrop);
      stack.push(modal);
      setTimeout(() => {
        const first = focusables[0] || confirmBtn;
        if (first && typeof first.focus === 'function') first.focus();
      }, 10);
    });
  }

  async function confirm(opts) {
    const cfg = opts || {};
    const result = await open({
      title: cfg.title || '确认操作',
      description: cfg.description || '',
      fields: [],
      confirmLabel: cfg.confirmLabel || '确定',
      cancelLabel: cfg.cancelLabel || '取消',
      danger: !!cfg.danger,
    });
    return result != null;
  }

  async function prompt(opts) {
    const cfg = opts || {};
    const result = await open({
      title: cfg.title || '输入',
      description: cfg.description || '',
      fields: [{
        name: 'value',
        label: cfg.label || '',
        type: cfg.type || 'text',
        value: cfg.value || '',
        placeholder: cfg.placeholder || '',
        required: !!cfg.required,
        maxLength: cfg.maxLength,
        minLength: cfg.minLength,
        hint: cfg.hint,
      }],
      confirmLabel: cfg.confirmLabel || '确定',
      cancelLabel: cfg.cancelLabel || '取消',
      danger: !!cfg.danger,
    });
    return result ? String(result.value || '') : null;
  }

  async function alert(opts) {
    const cfg = opts || {};
    await open({
      title: cfg.title || '提示',
      description: cfg.description || '',
      fields: [],
      confirmLabel: cfg.confirmLabel || '知道了',
      hideCancel: true,
    });
  }

  window.BlinkModal = { open, confirm, prompt, alert };
})();
