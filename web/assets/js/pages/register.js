/**
 * Register page — adapts to register config and posts the register request.
 */
(function () {
  'use strict';

  const { flash, errorText } = window.BlinkUI;
  const API = window.BlinkAPI;

  function humanize(err, fallback) {
    if (!err) return fallback || '操作失败';
    const raw = errorText(err).trim();
    const status = err && err.status;
    if (status === 409 || raw.indexOf('email already') !== -1) return '该邮箱已被注册';
    if (status === 400 && raw.indexOf('password too short') !== -1) return '密码至少 8 位';
    if (status === 400 && raw.indexOf('invalid email') !== -1) return '邮箱格式无效';
    if (status === 400 && raw.indexOf('verification code') !== -1) return '验证码无效或已过期';
    if (status === 429) return '发送过于频繁，请稍后再试';
    return raw || fallback || '操作失败';
  }

  const form = document.getElementById('reg');
  const msg = document.getElementById('msg');
  const sendBtn = document.getElementById('send-code');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');
  const confirmPasswordInput = document.getElementById('confirm-password');
  const nameInput = document.getElementById('name');
  const codeInput = document.getElementById('code');
  const codeGroup = document.getElementById('code-group');
  const submitBtn = form.querySelector('button[type="submit"]');
  let countdownTimer = null;
  let configLoaded = false;
  let verificationRequired = false;
  let actionsDisabled = true;

  function updateButtonState() {
    if (submitBtn) submitBtn.disabled = actionsDisabled;
    if (sendBtn) sendBtn.disabled = actionsDisabled || !verificationRequired || countdownTimer !== null;
  }

  function setActionState(disabled) {
    actionsDisabled = !!disabled;
    updateButtonState();
  }

  function setVerificationMode(required) {
    verificationRequired = !!required;
    if (codeGroup) codeGroup.hidden = !verificationRequired;
    if (codeInput) codeInput.disabled = !verificationRequired;
    if (sendBtn) sendBtn.hidden = !verificationRequired;
    if (!verificationRequired && codeInput) codeInput.value = '';
    updateButtonState();
  }

  function startCountdown(seconds) {
    let remaining = seconds;
    sendBtn.textContent = `${remaining}s 后可重发`;
    countdownTimer = setInterval(() => {
      remaining -= 1;
      if (remaining <= 0) {
        clearInterval(countdownTimer);
        countdownTimer = null;
        sendBtn.textContent = '发送验证码';
        updateButtonState();
        return;
      }
      sendBtn.textContent = `${remaining}s 后可重发`;
    }, 1000);
    updateButtonState();
  }

  async function loadRegisterConfig() {
    flash(msg, '正在加载注册配置…', '');
    setActionState(true);
    if (codeGroup) codeGroup.hidden = true;
    if (codeInput) codeInput.disabled = true;
    if (sendBtn) sendBtn.hidden = true;

    try {
      const config = await API.get('/auth/register/config');
      configLoaded = true;
      setVerificationMode(!!(config && config.email_verification_required));
      // enable actions on successful config load so buttons become usable
      setActionState(false);
      flash(msg, '', '');
    } catch (err) {
      setActionState(true);
      if (codeGroup) codeGroup.hidden = true;
      if (codeInput) codeInput.disabled = true;
      if (sendBtn) sendBtn.hidden = true;
      flash(msg, '加载注册配置失败，请刷新页面重试', 'err');
    }
  }

  sendBtn.addEventListener('click', async () => {
    if (!configLoaded || !verificationRequired) return;
    flash(msg, '', '');
    const email = emailInput.value.trim();
    if (!email || email.indexOf('@') === -1) {
      return flash(msg, '请输入有效邮箱', 'err');
    }
    setActionState(true);
    try {
      await API.post('/auth/register/send_code', { email });
      flash(msg, '验证码已发送，请查收邮件（未配置 SMTP 时写入服务日志）', 'ok');
      // Re-enable actions before starting countdown so centralized state+countdown
      // produce the correct disabled state via updateButtonState.
      setActionState(false);
      startCountdown(60);
    } catch (err) {
      // Re-enable actions via centralized state so UI state is consistent.
      setActionState(false);
      flash(msg, humanize(err, '发送失败'), 'err');
    }
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    if (!configLoaded) {
      flash(msg, '注册配置尚未加载完成，请稍后再试', 'err');
      return;
    }

    flash(msg, '', '');

    const email = emailInput.value.trim();
    const password = passwordInput.value;
    const confirmPassword = confirmPasswordInput.value;
    const name = nameInput.value.trim();
    const code = codeInput ? codeInput.value.trim() : '';

    if (!email || !password) return flash(msg, '请填写邮箱和密码', 'err');
    if (password.length < 8) return flash(msg, '密码至少 8 位', 'err');
    if (!confirmPassword) return flash(msg, '请确认密码', 'err');
    if (password !== confirmPassword) return flash(msg, '两次输入的密码不一致', 'err');
    if (verificationRequired && !/^\d{6}$/.test(code)) return flash(msg, '请输入 6 位验证码', 'err');

    const payload = { email, password, name };
    if (verificationRequired) payload.code = code;

    setActionState(true);

    try {
      await API.post('/auth/register', payload);
      flash(msg, '注册成功，正在跳转…', 'ok');
      window.location.href = '/web/feed.html';
    } catch (err) {
      flash(msg, humanize(err, '注册失败'), 'err');
      setActionState(false);
    }
  });

  loadRegisterConfig();
})();
