/**
 * Admin — rankings (post count / user activity by day/month/year).
 */
(function () {
  'use strict';

  const { el, clear, errorText } = window.BlinkUI;
  const AdminAPI = window.BlinkAdminAPI;

  const PERIOD_LABELS = {
    today: '今日',
    month: '本月',
    year: '本年',
  };

  function table(headers, rows) {
    const thead = el('thead', {}, [
      el('tr', {}, headers.map(function (h) { return el('th', {}, h); })),
    ]);
    const tbody = el('tbody', {}, rows.map(function (cells) {
      return el('tr', {}, cells.map(function (c) { return el('td', {}, String(c)); }));
    }));
    return el('table', { class: 'admin-table' }, [thead, tbody]);
  }

  function renderPostRankings(container, rankings) {
    rankings.forEach(function (r) {
      var label = PERIOD_LABELS[r.period] || r.period;
      container.appendChild(el('h3', {}, '发帖排名 — ' + label));
      if (!r.items || r.items.length === 0) {
        container.appendChild(el('p', { class: 'admin-subtitle' }, '暂无数据'));
        return;
      }
      var rows = r.items.map(function (item, idx) {
        return [idx + 1, item.user_name || '-', item.user_email || '-', Number(item.post_count)];
      });
      container.appendChild(table(['排名', '用户', '邮箱', '发帖数'], rows));
    });
  }

  function renderActivityRankings(container, rankings) {
    rankings.forEach(function (r) {
      var label = PERIOD_LABELS[r.period] || r.period;
      container.appendChild(el('h3', {}, '活跃排名 — ' + label));
      if (!r.items || r.items.length === 0) {
        container.appendChild(el('p', { class: 'admin-subtitle' }, '暂无数据'));
        return;
      }
      var rows = r.items.map(function (item, idx) {
        return [
          idx + 1,
          item.user_name || '-',
          item.user_email || '-',
          Number(item.post_count),
          Number(item.reply_count),
          Number(item.like_count),
          Number(item.total),
        ];
      });
      container.appendChild(table(['排名', '用户', '邮箱', '发帖', '评论', '点赞', '总计'], rows));
    });
  }

  async function mount(container) {
    var errEl = el('p', { class: 'err', role: 'alert' });
    container.appendChild(errEl);

    var wrap = el('div', { class: 'admin-rankings' });
    container.appendChild(wrap);

    try {
      var data = await AdminAPI.rankings({ limit: 10 });
      var postSection = el('div', { class: 'admin-form-card' }, [
        el('h2', {}, '发帖排名'),
        el('p', { class: 'admin-subtitle' }, '按今日 / 本月 / 本年统计已发布帖子数量。'),
      ]);
      renderPostRankings(postSection, data.post_rankings || []);
      wrap.appendChild(postSection);

      var activitySection = el('div', { class: 'admin-form-card' }, [
        el('h2', {}, '用户活跃排名'),
        el('p', { class: 'admin-subtitle' }, '综合发帖、评论、点赞数据排名。'),
      ]);
      renderActivityRankings(activitySection, data.activity_rankings || []);
      wrap.appendChild(activitySection);
    } catch (err) {
      errEl.textContent = errorText(err);
    }
    return { unmount: function () {} };
  }

  window.BlinkAdminModules = window.BlinkAdminModules || {};
  window.BlinkAdminModules.rankings = { mount: mount };
})();
