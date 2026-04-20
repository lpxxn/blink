# Blink 前端共享模块

所有新页面应使用 `/web/assets/` 下的模块；旧页面会在后续阶段（P3–P5）逐个迁移过来。

## 加载顺序

HTML `<head>` 里只引 CSS；`<script>` 建议放在 `</body>` 前，顺序如下：

```html
<!-- 1) 基础工具：API 封装 -->
<script src="/web/assets/js/api.js"></script>
<!-- 2) DOM / 列表工具 -->
<script src="/web/assets/js/ui.js"></script>
<!-- 3) 导航自定义元素（依赖 api.js） -->
<script src="/web/assets/js/nav.js"></script>
<!-- 4)（管理页）权限门禁（依赖 api.js） -->
<script src="/web/assets/js/admin-gate.js"></script>
<!-- 5) 页面自己的脚本 -->
<script src="/web/assets/js/pages/feed.js"></script>
```

导航只需一行：

```html
<blink-nav active="feed"></blink-nav>
```

## 全局命名空间

| 对象 | 说明 |
|---|---|
| `window.BlinkAPI` | fetch 封装：`.get/.post/.patch/.put/.del`、`.me()`、`.logout()`；抛 `BlinkAPI.Error`。 |
| `window.BlinkUI` | `el/append/clear/flash/errorText/mountList/createCursorPager`。 |
| `window.BlinkAdmin` | `.gate({ onReady })` 超级管理员门禁。 |
| `window.BlinkMD` | Markdown 解析（P4 时搬到 `assets/js/md.js`，当前仍在 `/web/md.js`）。 |

## 覆盖 API 基址

在 `<script>` 里设 `window.BLINK_API = 'https://host'` 可切换；默认同源。

## 错误处理约定

所有 `BlinkAPI.*` 失败会抛 `BlinkAPI.Error { status, message, body }`，message 已经是中文。页面代码只需要：

```js
try {
  await BlinkAPI.post('/api/posts', payload);
} catch (err) {
  BlinkUI.flash('msg', err.message);
}
```
