(function () {
  'use strict';
  const { flash, errorText } = window.BlinkUI;
  const API = window.BlinkAPI;

  const form = document.getElementById('forgot');
  const msg = document.getElementById('msg');
  const sendBtn = document.getElementById('send-code');
  const emailInput = document.getElementById('email');
  let timer = null;

  function startCountdown(seconds) {
    let n = seconds;
    sendBtn.disabled = true;
    sendBtn.textContent = `${n}s 后可重发`;
    timer = setInterval(() => {
      n -= 1;
      if (n <= 0) { clearInterval(timer); timer = null; sendBtn.disabled = false; sendBtn.textContent = '发送验证码'; return; }
      sendBtn.textContent = `${n}s 后可重发`;
    }, 1000);
  }

  sendBtn.addEventListener('click', async () => {
    flash(msg, '', '');
    const email = emailInput.value.trim();
    if (!email || email.indexOf('@') === -1) return flash(msg, '请输入有效邮箱', 'err');
    sendBtn.disabled = true;
    try {
      await API.post('/auth/password/send_code', { email });
      flash(msg, '验证码已发送（若该邮箱已注册）。10 分钟内有效。', 'ok');
      startCountdown(60);
    } catch (err) {
      sendBtn.disabled = false;
      flash(msg, errorText(err) || '发送失败', 'err');
    }
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    flash(msg, '', '');
    const email = emailInput.value.trim();
    const code = document.getElementById('code').value.trim();
    const newPassword = document.getElementById('new-password').value;
    if (!email) return flash(msg, '请输入邮箱', 'err');
    if (!/^\d{6}$/.test(code)) return flash(msg, '请输入 6 位验证码', 'err');
    if (newPassword.length < 8) return flash(msg, '新密码至少 8 位', 'err');

    const btn = form.querySelector('button[type="submit"]');
    if (btn) btn.disabled = true;
    try {
      await API.post('/auth/password/reset', { email, code, new_password: newPassword });
      flash(msg, '密码已重置，正在跳转登录…', 'ok');
      setTimeout(() => { window.location.href = '/web/login.html'; }, 600);
    } catch (err) {
      if (btn) btn.disabled = false;
      const raw = errorText(err) || '';
      const m = raw.indexOf('verification code') !== -1 ? '验证码无效或已过期'
              : raw.indexOf('password too short') !== -1 ? '新密码至少 8 位'
              : raw || '重置失败';
      flash(msg, m, 'err');
    }
  });
})();
