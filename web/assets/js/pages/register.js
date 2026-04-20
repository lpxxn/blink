/**
 * Register page — posts to /auth/register via BlinkAPI and maps a few
 * well-known backend error strings to friendlier Chinese messages.
 */
(function () {
  'use strict';

  const { flash, errorText } = window.BlinkUI;
  const API = window.BlinkAPI;

  function humanize(err) {
    if (!err) return '注册失败';
    const raw = errorText(err).trim();
    if (!raw) return '注册失败';
    const status = err && err.status;
    if (status === 409 || raw.indexOf('email already') !== -1) return '该邮箱已被注册';
    if (status === 400 && raw.indexOf('password too short') !== -1) return '密码至少 8 位';
    if (status === 400 && raw.indexOf('invalid email') !== -1) return '邮箱格式无效';
    return raw;
  }

  const form = document.getElementById('reg');
  const msg = document.getElementById('msg');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    flash(msg, '', '');

    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    const name = document.getElementById('name').value.trim();

    if (!email || !password) return flash(msg, '请填写邮箱和密码', 'err');
    if (password.length < 8)   return flash(msg, '密码至少 8 位', 'err');

    const submitBtn = form.querySelector('button[type="submit"]');
    if (submitBtn) submitBtn.disabled = true;

    try {
      await API.post('/auth/register', { email, password, name });
      flash(msg, '注册成功，正在跳转…', 'ok');
      window.location.href = '/web/feed.html';
    } catch (err) {
      flash(msg, humanize(err), 'err');
      if (submitBtn) submitBtn.disabled = false;
    }
  });
})();
