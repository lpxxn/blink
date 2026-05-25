/**
 * Admin — dashboard.
 */
(function () {
  'use strict';

  const { el, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;

  function num(v) {
    if (v == null) return 0;
    const n = Number(v);
    return Number.isFinite(n) ? n : 0;
  }

  function kpi(label, value, hint, alert) {
    return el('div', { class: 'admin-kpi' + (alert ? ' is-alert' : '') }, [
      el('span', { class: 'admin-kpi-label' }, label),
      el('span', { class: 'admin-kpi-value' }, String(value)),
      hint ? el('span', { class: 'admin-kpi-hint' }, hint) : null,
    ]);
  }

  async function mount(container, ctx) {
    const errEl = el('p', { class: 'err', role: 'alert' });
    container.appendChild(errEl);

    const kpiWrap = el('div', { class: 'admin-kpis' });
    container.appendChild(kpiWrap);

    const actionsCard = el('div', { class: 'admin-form-card' }, [
      el('h2', {}, '待处理事项'),
      el('p', { class: 'admin-subtitle' }, '点击跳转到对应视图。'),
    ]);
    const actionsRow = el('div', { class: 'admin-form-row' });
    actionsCard.appendChild(actionsRow);
    container.appendChild(actionsCard);

    try {
      const overview = await AdminAPI.overview();
      const appealsCount = num(overview.pending_appeals);
      const sensitiveCount = num(overview.pending_sensitive_hits);
      const feedbackCount = num(overview.open_feedback);

      kpiWrap.appendChild(kpi('用户总数', num(overview.user_count).toLocaleString()));
      kpiWrap.appendChild(kpi('帖子总数', num(overview.post_count).toLocaleString()));
      kpiWrap.appendChild(kpi('今日新发', num(overview.posts_today).toLocaleString(), '过去 24 小时'));
      if (overview.category_count != null) {
        kpiWrap.appendChild(kpi('分类数', num(overview.category_count).toLocaleString()));
      }
      kpiWrap.appendChild(kpi('待处理申诉', appealsCount.toLocaleString(), '需要审核', appealsCount > 0));
      kpiWrap.appendChild(kpi('敏感词待办', sensitiveCount.toLocaleString(), '命中敏感词', sensitiveCount > 0));
      kpiWrap.appendChild(kpi('待处理反馈', feedbackCount.toLocaleString(), '未关闭工单', feedbackCount > 0));

      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn ' + (appealsCount > 0 ? 'btn-primary' : 'btn-secondary'),
        onClick: () => ctx.navigate('appeals'),
      }, '处理申诉（' + appealsCount + '）'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn ' + (sensitiveCount > 0 ? 'btn-primary' : 'btn-secondary'),
        onClick: () => ctx.navigate('posts', { sensitive_hit_pending: 1 }),
      }, '敏感词待办（' + sensitiveCount + '）'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn ' + (feedbackCount > 0 ? 'btn-primary' : 'btn-secondary'),
        onClick: () => ctx.navigate('feedback', { status: 'open' }),
      }, '处理反馈（' + feedbackCount + '）'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-secondary',
        onClick: () => ctx.navigate('posts', { moderation_flag: 1 }),
      }, '查看违规帖子'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-ghost',
        onClick: () => ctx.navigate('categories'),
      }, '分类管理'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-ghost',
        onClick: () => ctx.navigate('rankings'),
      }, '排名统计'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-ghost',
        onClick: () => ctx.navigate('audit'),
      }, '审计日志'));
    } catch (err) {
      errEl.textContent = errorText(err);
    }

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.dashboard = { mount };
})();
