(function () {
  'use strict';
  const { flash, errorText } = window.BlinkUI;
  const API = window.BlinkAPI;

  const form = document.getElementById('login');
  const msg = document.getElementById('msg');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    flash(msg, '', '');
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    if (!email || !password) return flash(msg, '请填写邮箱和密码', 'err');

    const submitBtn = form.querySelector('button[type="submit"]');
    if (submitBtn) submitBtn.disabled = true;
    try {
      await API.post('/auth/login', { email, password });
      const params = new URLSearchParams(window.location.search);
      const next = params.get('next');
      window.location.href = next && next.startsWith('/') ? next : '/web/feed.html';
    } catch (err) {
      const raw = errorText(err).trim();
      const m = err && err.status === 401
        ? '邮箱或密码不正确'
        : err && err.status === 429
          ? '尝试过多，请稍后再试'
          : raw || '登录失败';
      flash(msg, m, 'err');
      if (submitBtn) submitBtn.disabled = false;
    }
  });
})();
