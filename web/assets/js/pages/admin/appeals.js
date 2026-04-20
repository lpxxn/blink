/**
 * Admin — appeals queue.
 *
 * Card-per-post layout: each card shows appeal body, moderation note and
 * a collapsible post preview. Approve / reject buttons open a modal so
 * admins can leave a rationale before resolving.
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;
  const Modal = window.BlinkModal;
  const MD = window.BlinkMD;

  const LIMIT = 50;

  function fmtTime(s) {
    if (!s) return '';
    try {
      const d = new Date(s);
      if (Number.isNaN(d.getTime())) return s;
      return d.toLocaleString('zh-CN', { hour12: false });
    } catch (_) { return s; }
  }

  function renderMarkdown(body) {
    const host = el('div');
    if (MD && typeof MD.parse === 'function') {
      host.innerHTML = MD.parse(body || '');
    } else {
      host.textContent = String(body || '');
    }
    return host;
  }

  async function mount(container, ctx) {
    let appeals = [];
    const expanded = new Set();

    const errEl = el('p', { class: 'err', role: 'alert' });
    const statusEl = el('p', { class: 'admin-subtitle' }, '正在加载…');
    const reloadBtn = el('button', {
      type: 'button', class: 'btn btn-secondary btn-sm',
      onClick: () => load(),
    }, '刷新');

    const toolbar = el('div', { class: 'admin-toolbar' }, [statusEl, el('span', { class: 'admin-toolbar-spacer' }), reloadBtn]);
    const wrap = el('div', { class: 'admin-cards' });

    container.appendChild(errEl);
    container.appendChild(toolbar);
    container.appendChild(wrap);

    function showErr(err) {
      errEl.textContent = err ? errorText(err) : '';
    }

    function render() {
      clear(wrap);
      if (appeals.length === 0) {
        wrap.appendChild(el('div', { class: 'admin-empty' }, '当前没有待处理申诉'));
        return;
      }
      appeals.forEach((p) => wrap.appendChild(renderCard(p)));
    }

    function renderCard(p) {
      const key = String(p.id);
      const isOpen = expanded.has(key);
      const card = el('div', { class: 'admin-card' });

      const head = el('div', { class: 'admin-card-head' }, [
        el('h3', {}, p.user_name && p.user_name.trim() ? p.user_name : '匿名作者'),
        el('span', { class: 'chip chip-muted' }, '#' + p.id),
      ]);
      card.appendChild(head);

      const meta = el('div', { class: 'admin-subtitle' },
        '创建 ' + fmtTime(p.created_at) + ' · 更新 ' + fmtTime(p.updated_at));
      card.appendChild(meta);

      const appealSection = el('div', { class: 'admin-card-section' }, [
        el('p', { class: 'admin-card-section-title' }, '作者申诉'),
        el('div', { class: 'admin-card-body' }, (p.appeal_body || '').trim() || '（未填写）'),
      ]);
      card.appendChild(appealSection);

      if (p.moderation_note) {
        card.appendChild(el('div', { class: 'admin-card-section' }, [
          el('p', { class: 'admin-card-section-title' }, '下架备注'),
          el('div', { class: 'admin-card-body' }, p.moderation_note),
        ]));
      }

      const previewSection = el('div', { class: 'admin-card-section', hidden: !isOpen });
      previewSection.appendChild(el('p', { class: 'admin-card-section-title' }, '原帖预览'));
      const previewBody = el('div', { class: 'admin-card-body admin-card-rendered' });
      previewBody.appendChild(renderMarkdown(p.body || '（空）'));
      previewSection.appendChild(previewBody);
      card.appendChild(previewSection);

      const toggleBtn = el('button', {
        type: 'button', class: 'btn btn-ghost btn-sm',
        onClick: () => {
          if (expanded.has(key)) expanded.delete(key);
          else expanded.add(key);
          render();
        },
      }, isOpen ? '收起原帖' : '查看原帖');

      const actions = el('div', { class: 'admin-card-actions' }, [
        toggleBtn,
        el('a', {
          class: 'btn btn-ghost btn-sm',
          href: '/web/post.html?id=' + encodeURIComponent(String(p.id)),
          target: '_blank', rel: 'noopener',
        }, '新窗口打开'),
        el('span', { class: 'admin-toolbar-spacer' }),
        el('button', {
          type: 'button', class: 'btn btn-secondary btn-sm',
          onClick: () => resolve(p, false),
        }, '驳回'),
        el('button', {
          type: 'button', class: 'btn btn-primary btn-sm',
          onClick: () => resolve(p, true),
        }, '通过并恢复'),
      ]);
      card.appendChild(actions);
      return card;
    }

    async function resolve(p, approve) {
      const title = approve ? '通过申诉 #' + p.id : '驳回申诉 #' + p.id;
      const description = approve
        ? '帖子将恢复为审核正常并重新对公众可见。作者会收到通知。'
        : '帖子保持下架状态。附说明会写入备注并通知作者。';
      const result = await Modal.open({
        title,
        description,
        fields: [{
          name: 'note',
          label: approve ? '处理说明（可选）' : '驳回说明',
          type: 'textarea',
          required: !approve,
          maxLength: 500,
          placeholder: approve ? '例如：经复核无违规…' : '例如：内容仍含敏感信息…',
        }],
        confirmLabel: approve ? '通过' : '驳回',
        danger: !approve,
      });
      if (!result) return;
      try {
        showErr(null);
        await AdminAPI.resolveAppeal(p.id, approve, result.note || '');
        await load();
        ctx.refreshBadges();
      } catch (err) { showErr(err); }
    }

    async function load() {
      showErr(null);
      reloadBtn.disabled = true;
      statusEl.textContent = '正在加载…';
      try {
        const d = await AdminAPI.listPosts({ appeal_pending: 1, limit: LIMIT, offset: 0 });
        appeals = d.posts || [];
        statusEl.textContent = appeals.length
          ? '待处理 ' + appeals.length + ' 条'
          : '当前没有待处理申诉';
        render();
      } catch (err) {
        statusEl.textContent = '';
        showErr(err);
      } finally {
        reloadBtn.disabled = false;
      }
    }

    await load();

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.appeals = { mount };
})();
