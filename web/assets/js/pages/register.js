/**
 * Register page — posts the email-verification code + register request.
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
  let countdownTimer = null;

  function startCountdown(seconds) {
    let remaining = seconds;
    sendBtn.disabled = true;
    sendBtn.textContent = `${remaining}s 后可重发`;
    countdownTimer = setInterval(() => {
      remaining -= 1;
      if (remaining <= 0) {
        clearInterval(countdownTimer);
        countdownTimer = null;
        sendBtn.disabled = false;
        sendBtn.textContent = '发送验证码';
        return;
      }
      sendBtn.textContent = `${remaining}s 后可重发`;
    }, 1000);
  }

  sendBtn.addEventListener('click', async () => {
    flash(msg, '', '');
    const email = emailInput.value.trim();
    if (!email || email.indexOf('@') === -1) {
      return flash(msg, '请输入有效邮箱', 'err');
    }
    sendBtn.disabled = true;
    try {
      await API.post('/auth/register/send_code', { email });
      flash(msg, '验证码已发送，请查收邮件（未配置 SMTP 时写入服务日志）', 'ok');
      startCountdown(60);
    } catch (err) {
      sendBtn.disabled = false;
      flash(msg, humanize(err, '发送失败'), 'err');
    }
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    flash(msg, '', '');

    const email = emailInput.value.trim();
    const password = document.getElementById('password').value;
    const name = document.getElementById('name').value.trim();
    const code = document.getElementById('code').value.trim();

    if (!email || !password) return flash(msg, '请填写邮箱和密码', 'err');
    if (password.length < 8)   return flash(msg, '密码至少 8 位', 'err');
    if (!/^\d{6}$/.test(code)) return flash(msg, '请输入 6 位验证码', 'err');

    const submitBtn = form.querySelector('button[type="submit"]');
    if (submitBtn) submitBtn.disabled = true;

    try {
      await API.post('/auth/register', { email, password, name, code });
      flash(msg, '注册成功，正在跳转…', 'ok');
      window.location.href = '/web/feed.html';
    } catch (err) {
      flash(msg, humanize(err, '注册失败'), 'err');
      if (submitBtn) submitBtn.disabled = false;
    }
  });
})();
