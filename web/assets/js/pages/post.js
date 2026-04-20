/* Post detail page: post view + threaded replies + reply form + admin actions. */
(function () {
  'use strict';

  const { BlinkAPI, BlinkUI, BlinkMD } = window;
  const { el, clear, flash, errorText } = BlinkUI;

  const params = new URLSearchParams(location.search);
  const postId = params.get('id');

  let myUserId = null;
  let myRole = null;
  let parentReplyId = null;

  function sid(v) { return v != null ? String(v) : ''; }
  function isSuperAdmin() { return myRole === 'super_admin'; }

  function authorLabel(obj) {
    const n = obj && obj.user_name != null ? String(obj.user_name).trim() : '';
    return n || ('用户 ' + sid(obj && obj.user_id));
  }

  function showErr(msg) { flash('err', msg || '', msg ? 'err' : ''); }

  async function loadMe() {
    try {
      const d = await BlinkAPI.me();
      if (d) {
        if (d.user_id != null) myUserId = sid(d.user_id);
        if (d.role != null) myRole = String(d.role);
      }
    } catch (_) { /* ignore */ }
  }

  // ---------- Post ----------
  function renderModerationBanner(p) {
    const mf = p.moderation_flag != null ? Number(p.moderation_flag) : 0;
    if (mf !== 1 && mf !== 2) return null;

    const title = el('strong', {}, mf === 2 ? '该帖已被管理员下架' : '该帖已被标记为违规，公开流中暂不展示');
    const label = el('p', { class: 'field-hint', style: { margin: '0.35rem 0 0' } }, mf === 2 ? '原因' : '说明');
    const noteRaw = (p.moderation_note && String(p.moderation_note).trim()) ? String(p.moderation_note).trim() : '';
    const note = el('pre', { style: { margin: 0, whiteSpace: 'pre-wrap', fontFamily: 'inherit', fontSize: '0.9rem' } },
      noteRaw ? BlinkMD.formatModerationNote(noteRaw) : '（未填写）');

    const children = [title, label, note];
    if (Number(p.appeal_status) === 1) {
      children.push(el('p', { class: 'ok', style: { margin: '0.75rem 0 0' } },
        mf === 2 ? '你的申诉或复核申请正在等待管理员处理。' : '你的恢复申请正在等待管理员处理。'));
    }
    if (myUserId && sid(p.user_id) === myUserId) {
      children.push(el('p', { class: 'field-hint', style: { marginTop: '0.5rem' } }, [
        '如需修改内容或提交申请，请前往 ',
        el('a', { href: '/web/edit-post.html?id=' + encodeURIComponent(sid(p.id)) }, '编辑页'),
        '。',
      ]));
    }
    if (isSuperAdmin()) {
      children.push(el('div', { class: 'btn-row', style: { marginTop: '0.75rem' } }, [
        el('button', {
          type: 'button', class: 'btn btn-primary btn-sm',
          onClick: () => adminRestorePost(sid(p.id), mf),
        }, mf === 2 ? '管理员：恢复上架' : '管理员：解除违规'),
      ]));
    }
    return el('div', { class: 'banner-warn', style: { marginBottom: '0.75rem' } }, children);
  }

  function renderPost(p) {
    const root = document.getElementById('post');
    clear(root);

    const banner = renderModerationBanner(p);
    if (banner) root.appendChild(banner);

    root.appendChild(el('div', { class: 'meta' },
      '#' + sid(p.id) + ' · ' + authorLabel(p) + ' · 状态 ' + sid(p.status)));

    const body = el('div', { class: 'markdown-body' });
    body.innerHTML = BlinkMD.parse(p.body || '');
    root.appendChild(body);

    const orphans = BlinkMD.orphanImages(p.body || '', p.images || []);
    if (orphans.length) {
      root.appendChild(el('div', { class: 'imgs' },
        orphans.map((u) => el('img', { src: u, alt: '' }))));
    }

    if (myUserId && sid(p.user_id) === myUserId) {
      root.appendChild(el('div', { class: 'post-author-actions' }, [
        el('a', {
          class: 'btn btn-secondary btn-sm',
          href: '/web/edit-post.html?id=' + encodeURIComponent(sid(p.id)),
        }, '编辑此帖'),
      ]));
    }
  }

  async function loadPost() {
    try {
      const p = await BlinkAPI.get('/api/posts/' + encodeURIComponent(postId));
      renderPost(p);
    } catch (err) {
      if (err && err.status === 404) showErr('帖子不存在或无权查看');
      else showErr(errorText(err));
    }
  }

  async function adminRestorePost(id, mf) {
    const msg = mf === 2
      ? '确定恢复该帖上架？将清除下架标记、清空备注，并把状态设为「已发布」。'
      : '确定解除违规标记？将清空备注；不改变帖子发布/草稿状态。';
    if (!window.confirm(msg)) return;
    showErr('');
    const body = { moderation_flag: 0, moderation_note: '' };
    if (mf === 2) body.status = 1;
    try {
      await BlinkAPI.patch('/admin/api/posts/' + encodeURIComponent(id), body);
      loadPost();
    } catch (err) {
      showErr(errorText(err));
    }
  }

  // ---------- Replies ----------
  function replyDepth(x, byId) {
    let d = 0;
    let pid = x.parent_reply_id;
    const guard = {};
    while (pid != null && pid !== '' && byId[sid(pid)]) {
      if (guard[sid(pid)]) break;
      guard[sid(pid)] = 1;
      d++;
      pid = byId[sid(pid)].parent_reply_id;
      if (d > 40) break;
    }
    return d;
  }

  function buildReplyTree(list) {
    const byId = {};
    list.forEach((x) => { byId[sid(x.id)] = x; });
    const children = {};
    const roots = [];
    list.forEach((x) => {
      const pid = x.parent_reply_id;
      if (pid != null && pid !== '' && byId[sid(pid)]) {
        const key = sid(pid);
        (children[key] = children[key] || []).push(x);
      } else {
        roots.push(x);
      }
    });
    const byIdAsc = (a, b) => {
      const ai = Number(a.id), bi = Number(b.id);
      if (!isNaN(ai) && !isNaN(bi)) return ai - bi;
      return sid(a.id) < sid(b.id) ? -1 : sid(a.id) > sid(b.id) ? 1 : 0;
    };
    roots.sort(byIdAsc);
    Object.keys(children).forEach((k) => children[k].sort(byIdAsc));
    const ordered = [];
    (function dfs(nodes) {
      nodes.forEach((n) => {
        ordered.push(n);
        const kids = children[sid(n.id)];
        if (kids && kids.length) dfs(kids);
      });
    })(roots);
    return { ordered, byId };
  }

  function renderReply(x, byId) {
    const depth = replyDepth(x, byId);
    const children = [document.createTextNode(x.body)];

    let metaLine = authorLabel(x) + ' · #' + sid(x.id);
    if (x.parent_reply_id != null && x.parent_reply_id !== '') {
      const par = byId[sid(x.parent_reply_id)];
      metaLine = '↪ 回复 ' + (par ? authorLabel(par) : '#' + sid(x.parent_reply_id)) + ' · ' + metaLine;
    }
    children.push(el('div', { class: 'meta' }, metaLine));

    const actions = el('div', { class: 'reply-actions' });
    actions.appendChild(el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => setReplyTarget(x),
    }, '回复'));

    if (isSuperAdmin()) {
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-secondary btn-sm',
        onClick: () => adminHideReply(x),
      }, '隐藏'));
    }
    if (myUserId && sid(x.user_id) === myUserId) {
      actions.appendChild(el('button', {
        type: 'button', class: 'btn btn-secondary btn-sm',
        onClick: () => deleteReply(x),
      }, '删除'));
    }
    children.push(actions);

    return el('div', {
      class: 'reply' + (depth > 0 ? ' reply-nested' : ''),
      style: { marginLeft: (depth * 14) + 'px' },
    }, children);
  }

  function setReplyTarget(x) {
    parentReplyId = sid(x.id);
    const hint = document.getElementById('reply-hint');
    hint.innerHTML = '正在回复 <strong>#' + parentReplyId + '</strong>（' + authorLabel(x) + '）';
    hint.hidden = false;
    document.getElementById('cancel-reply').hidden = false;
    document.getElementById('rbody').focus();
  }

  function clearReplyTarget() {
    parentReplyId = null;
    const hint = document.getElementById('reply-hint');
    hint.hidden = true;
    hint.innerHTML = '';
    document.getElementById('cancel-reply').hidden = true;
  }

  async function adminHideReply(x) {
    if (!window.confirm('隐藏评论 #' + sid(x.id) + '（含所有子评论）？')) return;
    try {
      await BlinkAPI.patch('/admin/api/replies/' + encodeURIComponent(sid(x.id)), { hidden: true });
      loadReplies();
    } catch (err) {
      showErr(errorText(err));
    }
  }

  async function deleteReply(x) {
    if (!window.confirm('删除这条评论？')) return;
    try {
      await BlinkAPI.del('/api/replies/' + encodeURIComponent(sid(x.id)));
      loadReplies();
    } catch (err) {
      showErr(errorText(err));
    }
  }

  async function loadReplies() {
    try {
      const d = await BlinkAPI.get('/api/posts/' + encodeURIComponent(postId) + '/replies?limit=100');
      const { ordered, byId } = buildReplyTree(d.replies || []);
      const box = document.getElementById('replies');
      clear(box);
      ordered.forEach((x) => box.appendChild(renderReply(x, byId)));
    } catch (err) {
      showErr(errorText(err));
    }
  }

  async function sendReply() {
    showErr('');
    const body = document.getElementById('rbody').value.trim();
    if (!body) return;
    const payload = { body };
    if (parentReplyId) payload.parent_reply_id = parentReplyId;
    try {
      await BlinkAPI.post('/api/posts/' + encodeURIComponent(postId) + '/replies', payload);
      document.getElementById('rbody').value = '';
      clearReplyTarget();
      loadReplies();
    } catch (err) {
      showErr(errorText(err));
    }
  }

  async function init() {
    if (!postId) { showErr('缺少 id'); return; }
    document.getElementById('cancel-reply').addEventListener('click', clearReplyTarget);
    document.getElementById('send').addEventListener('click', sendReply);
    await loadMe();
    loadPost();
    loadReplies();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
