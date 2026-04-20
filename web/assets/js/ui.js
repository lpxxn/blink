/**
 * BlinkUI — small DOM + list helpers.
 *
 * `el(tag, props, children)` builds a node without a VDOM:
 *   el('a', { href: '/x', class: 'btn', onClick: go }, '打开')
 *   el('div', { class: 'card' }, [title, body])
 *
 * Special prop keys:
 *   class / className → node.className
 *   style (object)    → Object.assign(node.style, ...)
 *   dataset (object)  → Object.assign(node.dataset, ...)
 *   text              → textContent
 *   html              → innerHTML (use sparingly; sanitize upstream)
 *   onX (function)    → addEventListener('x', ...)
 *   hidden (bool)     → sets/removes the hidden attribute
 */
(function () {
  'use strict';

  function append(parent, children) {
    if (children == null || children === false) return;
    if (Array.isArray(children)) {
      for (const c of children) append(parent, c);
      return;
    }
    if (typeof children === 'string' || typeof children === 'number') {
      parent.appendChild(document.createTextNode(String(children)));
      return;
    }
    if (children instanceof Node) {
      parent.appendChild(children);
    }
  }

  function applyProps(node, props) {
    if (!props) return;
    for (const k of Object.keys(props)) {
      const v = props[k];
      if (v == null || v === false) continue;
      if (k === 'class' || k === 'className') {
        node.className = v;
      } else if (k === 'style' && typeof v === 'object') {
        Object.assign(node.style, v);
      } else if (k === 'dataset' && typeof v === 'object') {
        Object.assign(node.dataset, v);
      } else if (k === 'text') {
        node.textContent = v;
      } else if (k === 'html') {
        node.innerHTML = v;
      } else if (k.length > 2 && k.startsWith('on') && typeof v === 'function') {
        node.addEventListener(k.slice(2).toLowerCase(), v);
      } else if (k === 'hidden') {
        if (v) node.setAttribute('hidden', '');
      } else if (typeof v === 'boolean') {
        if (v) node.setAttribute(k, '');
      } else {
        node.setAttribute(k, v);
      }
    }
  }

  function el(tag, props, children) {
    const node = document.createElement(tag);
    applyProps(node, props);
    append(node, children);
    return node;
  }

  function resolve(target) {
    return typeof target === 'string' ? document.getElementById(target) : target;
  }

  function clear(target) {
    const node = resolve(target);
    if (!node) return;
    while (node.firstChild) node.removeChild(node.firstChild);
  }

  /** Show a message in a status/error region. level: 'ok' | 'err' | '' */
  function flash(target, message, level) {
    const node = resolve(target);
    if (!node) return;
    node.textContent = message || '';
    if (level === undefined) level = message ? 'err' : '';
    node.className = level;
  }

  function errorText(err) {
    if (!err) return '';
    if (typeof err === 'string') return err;
    if (err.message) return String(err.message);
    return String(err);
  }

  /** Render a list into a container; shows `emptyText` if items is empty. */
  function mountList(container, items, renderItem, emptyText) {
    const node = resolve(container);
    if (!node) return;
    clear(node);
    if (!items || items.length === 0) {
      node.appendChild(el('p', { class: 'empty-hint' }, emptyText || '暂无数据'));
      return;
    }
    for (const item of items) {
      const child = renderItem(item);
      if (child) node.appendChild(child);
    }
  }

  /** Create a cursor-based pager. Caller provides a loader (cursor) => {items, next}. */
  function createCursorPager({ loader, onAppend, onError, moreButton }) {
    let cursor = null;
    let busy = false;

    const moreBtn = typeof moreButton === 'string' ? document.getElementById(moreButton) : moreButton;

    async function load(reset) {
      if (busy) return;
      if (!reset && cursor == null) return;
      busy = true;
      if (moreBtn) moreBtn.disabled = true;
      try {
        const res = await loader(reset ? null : cursor);
        if (reset) onAppend(res.items, true);
        else onAppend(res.items, false);
        cursor = res.next || null;
        if (moreBtn) moreBtn.hidden = !cursor || !res.items || res.items.length === 0;
      } catch (err) {
        if (onError) onError(err);
      } finally {
        busy = false;
        if (moreBtn) moreBtn.disabled = false;
      }
    }

    if (moreBtn) {
      moreBtn.addEventListener('click', () => load(false));
    }

    return {
      reset: () => load(true),
      more: () => load(false),
      get cursor() { return cursor; },
    };
  }

  window.BlinkUI = { el, append, clear, flash, errorText, mountList, createCursorPager };
})();
