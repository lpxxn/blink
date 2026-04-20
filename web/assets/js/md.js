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

  /**
   * Return image URLs in `images` that are NOT already referenced by
   * a Markdown `![...](url)` in `body`. Caller renders the result as a
   * supplementary gallery so we don't duplicate inline images.
   */
  function orphanImages(body, images) {
    if (!images || !images.length) return [];
    var referenced = referencedImageUrls(body);
    var out = [];
    for (var i = 0; i < images.length; i++) {
      var u = String(images[i] || '');
      if (!u) continue;
      if (!referenced[u]) out.push(u);
    }
    return out;
  }

  /** Parse all Markdown image URLs out of a body. */
  function referencedImageUrls(body) {
    var set = {};
    if (!body) return set;
    // ![alt](url) or ![alt](<url with spaces>)
    var re = /!\[[^\]]*\]\(\s*(?:<([^>]+)>|([^\s)]+))(?:\s+"[^"]*")?\s*\)/g;
    var m;
    while ((m = re.exec(String(body)))) {
      var u = (m[1] != null ? m[1] : m[2]) || '';
      if (u) set[u] = true;
    }
    return set;
  }

  /** Author-facing moderation_note: legacy prefix "sensitive_hit:" → 有敏感词： */
  function formatModerationNote(note) {
    if (note == null) return '';
    var s = String(note).trim();
    if (!s) return '';
    var legacy = /^sensitive_hit:\s*/i;
    if (legacy.test(s)) {
      s = '有敏感词：' + s.replace(legacy, '').trim();
      s = s.replace(/, /g, '、');
    }
    return s;
  }

  global.BlinkMD = {
    parse: parse,
    plainSnippet: plainSnippet,
    insertImageMarkdown: insertImageMarkdown,
    mdImageUrl: mdImageUrl,
    orphanImages: orphanImages,
    referencedImageUrls: referencedImageUrls,
    formatModerationNote: formatModerationNote,
  };
})(typeof window !== 'undefined' ? window : this);
