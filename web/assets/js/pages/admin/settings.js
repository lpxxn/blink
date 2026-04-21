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

    // ----- SMTP settings panel -----
    const smtpErr = el('p', { class: 'err', role: 'alert' });
    const smtpOk = el('p', { class: 'ok' });
    const smtpCard = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '邮件服务 (SMTP)'),
      el('p', { class: 'admin-subtitle' },
        '注册与找回密码邮件通过此 SMTP 发送。关闭后邮件会写入服务日志以便本地联调。'),
    ]);

    function field(label, input, hint) {
      const wrap = el('div', { class: 'field' }, [el('label', {}, label), input]);
      if (hint) wrap.appendChild(el('p', { class: 'field-hint' }, hint));
      return wrap;
    }
    const smtpEnabled = el('input', { type: 'checkbox' });
    const smtpHost = el('input', { type: 'text', placeholder: 'smtp.example.com' });
    const smtpPort = el('input', { type: 'number', min: '1', max: '65535', placeholder: '465' });
    const smtpUsername = el('input', { type: 'text', autocomplete: 'off' });
    const smtpPassword = el('input', { type: 'password', autocomplete: 'new-password', placeholder: '留空表示保持当前密码' });
    const smtpFrom = el('input', { type: 'email', placeholder: 'noreply@example.com' });
    const smtpFromName = el('input', { type: 'text', placeholder: 'Blink' });
    const smtpSecurity = el('select', {}, [
      el('option', { value: 'starttls' }, 'STARTTLS (推荐，587)'),
      el('option', { value: 'ssl' }, 'SSL / TLS (465)'),
      el('option', { value: 'plain' }, 'Plain (不加密，25)'),
    ]);

    smtpCard.appendChild(field('启用 SMTP', el('label', { class: 'modal-radio', style: { padding: '.4rem 0' } }, [smtpEnabled, el('span', {}, '开启后才真的发送邮件')])));
    smtpCard.appendChild(field('主机 (host)', smtpHost));
    smtpCard.appendChild(field('端口 (port)', smtpPort));
    smtpCard.appendChild(field('用户名', smtpUsername));
    smtpCard.appendChild(field('密码', smtpPassword, '出于安全考虑，密码不会回显。留空提交表示保持不变。'));
    smtpCard.appendChild(field('发件人地址', smtpFrom));
    smtpCard.appendChild(field('发件人显示名', smtpFromName));
    smtpCard.appendChild(field('加密方式', smtpSecurity));

    const saveBtn = el('button', { type: 'button', class: 'btn btn-primary' }, '保存 SMTP 设置');
    const testTo = el('input', { type: 'email', placeholder: '测试收件人（可填你的邮箱）', style: { flex: 1 } });
    const testBtn = el('button', { type: 'button', class: 'btn btn-secondary' }, '发送测试邮件');
    const testRow = el('div', { style: { display: 'flex', gap: '.5rem', alignItems: 'stretch', marginTop: '.75rem' } }, [testTo, testBtn]);
    smtpCard.appendChild(el('div', { class: 'btn-row' }, [saveBtn]));
    smtpCard.appendChild(testRow);

    container.appendChild(smtpErr);
    container.appendChild(smtpOk);
    container.appendChild(smtpCard);

    function smtpShowErr(e) { smtpErr.textContent = e ? errorText(e) : ''; smtpOk.textContent = ''; }
    function smtpShowOk(m) { smtpErr.textContent = ''; smtpOk.textContent = m || ''; }

    async function loadSMTP() {
      try {
        const v = await AdminAPI.getSMTPSettings();
        smtpEnabled.checked = !!v.enabled;
        smtpHost.value = v.host || '';
        smtpPort.value = v.port ? String(v.port) : '';
        smtpUsername.value = v.username || '';
        smtpPassword.value = '';
        smtpPassword.placeholder = v.has_password ? '已设置，留空保持不变' : '请输入 SMTP 密码';
        smtpFrom.value = v.from || '';
        smtpFromName.value = v.from_name || '';
        smtpSecurity.value = (v.security || 'starttls');
      } catch (e) {
        smtpShowErr(e);
      }
    }

    saveBtn.addEventListener('click', async () => {
      smtpShowOk(null);
      saveBtn.disabled = true;
      try {
        const body = {
          enabled: !!smtpEnabled.checked,
          host: smtpHost.value.trim(),
          port: smtpPort.value ? Number(smtpPort.value) : 0,
          username: smtpUsername.value.trim(),
          from: smtpFrom.value.trim(),
          from_name: smtpFromName.value.trim(),
          security: smtpSecurity.value,
        };
        if (smtpPassword.value) body.password = smtpPassword.value;
        if (!body.port) delete body.port;
        await AdminAPI.setSMTPSettings(body);
        smtpShowOk('已保存');
        await loadSMTP();
      } catch (e) {
        smtpShowErr(e);
      } finally {
        saveBtn.disabled = false;
      }
    });
    testBtn.addEventListener('click', async () => {
      smtpShowOk(null);
      const to = testTo.value.trim();
      if (!to) return smtpShowErr(new Error('请填写测试收件人邮箱'));
      testBtn.disabled = true;
      try {
        await AdminAPI.testSMTP(to);
        smtpShowOk('已发送测试邮件（请查收；未启用 SMTP 时写入服务日志）');
      } catch (e) {
        smtpShowErr(e);
      } finally {
        testBtn.disabled = false;
      }
    });

    await loadSMTP();

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.settings = { mount };
})();
