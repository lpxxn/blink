/**
 * Admin — settings.
 *
 * Currently exposes the sensitive-post-mode toggle. Uses a radio card so
 * semantics of each option are inline.
 */
(function () {
  'use strict';

  const { el, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;

  const MODE_OPTIONS = [
    {
      value: 'review',
      label: '管理员审查后下架',
      hint: '命中敏感词只标记违规，等待管理员审阅。',
    },
    {
      value: 'auto_remove',
      label: '自动下架',
      hint: '命中敏感词立即下架并通知作者。',
    },
  ];

  async function mount(container, ctx) {
    const errEl = el('p', { class: 'err', role: 'alert' });
    const okEl = el('p', { class: 'ok' });

    const card = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '发布后敏感词处理模式'),
      el('p', { class: 'admin-subtitle' },
        '帖子发布后会由后台异步扫描。选择命中敏感词时的默认处理方式。'),
    ]);

    const radioGroup = el('div', { class: 'modal-radio-group' });
    const radios = {};
    MODE_OPTIONS.forEach((opt) => {
      const id = 'mode-' + opt.value;
      const input = el('input', { type: 'radio', name: 'sensitive-mode', id, value: opt.value });
      input.addEventListener('change', () => { if (input.checked) save(opt.value); });
      radios[opt.value] = input;
      const label = el('label', { class: 'modal-radio', for: id }, [
        input,
        el('span', {}, [
          el('strong', {}, opt.label),
          el('p', { class: 'modal-radio-hint' }, opt.hint),
        ]),
      ]);
      radioGroup.appendChild(label);
    });
    card.appendChild(radioGroup);

    const hint = el('p', { class: 'admin-subtitle' }, '加载中…');
    card.appendChild(hint);

    container.appendChild(errEl);
    container.appendChild(okEl);
    container.appendChild(card);

    function showErr(err) { errEl.textContent = err ? errorText(err) : ''; okEl.textContent = ''; }
    function showOk(msg) { errEl.textContent = ''; okEl.textContent = msg || ''; }

    function reflect(mode) {
      const v = mode || 'review';
      Object.keys(radios).forEach((k) => { radios[k].checked = (k === v); });
      const opt = MODE_OPTIONS.find((o) => o.value === v);
      hint.textContent = '当前模式：' + (opt ? opt.label : v);
    }

    async function load() {
      try {
        const d = await AdminAPI.getSensitivePostMode();
        reflect(d && d.mode ? String(d.mode) : 'review');
      } catch (err) {
        hint.textContent = '';
        showErr(err);
      }
    }

    async function save(mode) {
      showOk(null);
      Object.values(radios).forEach((r) => { r.disabled = true; });
      try {
        await AdminAPI.setSensitivePostMode(mode);
        showOk('已保存：' + (MODE_OPTIONS.find((o) => o.value === mode) || {}).label);
        await load();
      } catch (err) {
        showErr(err);
        await load();
      } finally {
        Object.values(radios).forEach((r) => { r.disabled = false; });
      }
    }

    await load();
    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.settings = { mount };
})();
