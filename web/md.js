/* Depends on global `marked` and `DOMPurify` (load before this file). */
(function (global) {
  'use strict';

  function parse(md) {
    if (md == null) md = '';
    var mk = global.marked;
    var purify = global.DOMPurify;
    if (!mk || !purify) {
      var esc = document.createElement('div');
      esc.textContent = String(md);
      return esc.innerHTML;
    }
    var raw = mk.parse(String(md), {
      breaks: true,
      gfm: true,
    });
    return purify.sanitize(raw, {
      FORBID_TAGS: ['style', 'form', 'input', 'button', 'script', 'iframe', 'object', 'embed', 'base'],
    });
  }

  /** Wrap URL for Markdown image when it contains spaces or parentheses. */
  function mdImageUrl(url) {
    var s = String(url);
    if (/[\s()]/.test(s)) {
      return '<' + s.replace(/>/g, '%3E') + '>';
    }
    return s;
  }

  /**
   * Insert a Markdown image line at the textarea cursor (or append).
   * Syntax: ![alt](url) — rendered after upload alongside gallery `images` JSON.
   */
  function insertImageMarkdown(textarea, url, alt) {
    if (!textarea || url == null || url === '') return;
    alt = alt != null && String(alt).trim() !== '' ? String(alt).trim() : '图片';
    alt = alt.replace(/\]/g, '');
    var line = '\n\n![' + alt + '](' + mdImageUrl(url) + ')\n';
    var v = textarea.value;
    var start = textarea.selectionStart;
    var end = textarea.selectionEnd;
    if (typeof start !== 'number' || typeof end !== 'number') {
      textarea.value = v + line;
    } else {
      textarea.value = v.slice(0, start) + line + v.slice(end);
      var pos = start + line.length;
      textarea.focus();
      try {
        textarea.setSelectionRange(pos, pos);
      } catch (e) { /* ignore */ }
    }
  }

  /** Plain-text excerpt for list titles (strips Markdown). */
  function plainSnippet(md, maxLen) {
    maxLen = maxLen || 80;
    var html = parse(md);
    var d = document.createElement('div');
    d.innerHTML = html;
    var t = (d.textContent || '').replace(/\s+/g, ' ').trim();
    if (!t) return '（无正文）';
    if (t.length <= maxLen) return t;
    return t.slice(0, maxLen) + '…';
  }

  global.BlinkMD = {
    parse: parse,
    plainSnippet: plainSnippet,
    insertImageMarkdown: insertImageMarkdown,
    mdImageUrl: mdImageUrl,
  };
})(typeof window !== 'undefined' ? window : this);
