/**
 * Admin — dashboard.
 *
 * Shows KPI tiles from `/admin/api/overview`, a pending-appeals counter,
 * and shortcut links into the other admin views.
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

    let appealsCount = 0;

    function render() {
      kpiWrap.textContent = '';
      actionsRow.textContent = '';

      // (Filled in below once we have data.)
    }

    render();

    try {
      const [overview, appealsProbe] = await Promise.all([
        AdminAPI.overview(),
        AdminAPI.listPosts({ appeal_pending: 1, limit: 1, offset: 0 }).catch(() => ({ total: 0 })),
      ]);

      appealsCount = num(appealsProbe.total);
      const userCount = num(overview.user_count);
      const postCount = num(overview.post_count);
      const postsToday = num(overview.posts_today);
      const categoryCount = overview.category_count != null ? num(overview.category_count) : null;

      kpiWrap.appendChild(kpi('用户总数', userCount.toLocaleString()));
      kpiWrap.appendChild(kpi('帖子总数', postCount.toLocaleString()));
      kpiWrap.appendChild(kpi('今日新发', postsToday.toLocaleString(), '过去 24 小时'));
      if (categoryCount != null) {
        kpiWrap.appendChild(kpi('分类数', categoryCount.toLocaleString()));
      }
      kpiWrap.appendChild(kpi('待处理申诉', appealsCount.toLocaleString(), '需要审核', appealsCount > 0));

      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn ' + (appealsCount > 0 ? 'btn-primary' : 'btn-secondary'),
        onClick: () => ctx.navigate('appeals'),
      }, '处理申诉（' + appealsCount + '）'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-secondary',
        onClick: () => ctx.navigate('posts', { moderation_flag: 1 }),
      }, '查看违规帖子'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-secondary',
        onClick: () => ctx.navigate('posts', { moderation_flag: 2 }),
      }, '查看已下架'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-ghost',
        onClick: () => ctx.navigate('sensitive'),
      }, '敏感词管理'));
      actionsRow.appendChild(el('button', {
        type: 'button',
        class: 'btn btn-ghost',
        onClick: () => ctx.navigate('settings'),
      }, '后台设置'));
    } catch (err) {
      errEl.textContent = errorText(err);
    }

    return { unmount() {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.dashboard = { mount };
})();
