/* ============================================
   MGHUB 主题切换（浅色/深色/跟随系统）
   ============================================ */
(function () {
  'use strict';

  const STORAGE_KEY = 'mghub_theme';

  function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  function getStoredTheme() {
    try {
      return localStorage.getItem(STORAGE_KEY) || 'auto';
    } catch (e) {
      return 'auto';
    }
  }

  function applyTheme(theme) {
    const effective = theme === 'auto' ? getSystemTheme() : theme;
    document.documentElement.setAttribute('data-theme', effective);
    return effective;
  }

  function initTheme() {
    const theme = getStoredTheme();
    applyTheme(theme);

    // 监听系统主题变化（仅当 auto 模式时）
    if (window.matchMedia) {
      window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function () {
        if (getStoredTheme() === 'auto') {
          applyTheme('auto');
        }
      });
    }
  }

  function cycleTheme() {
    const current = getStoredTheme();
    let next;
    if (current === 'light') next = 'dark';
    else if (current === 'dark') next = 'auto';
    else next = 'light';

    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch (e) {}
    applyTheme(next);
    return next;
  }

  function getThemeLabel(theme) {
    if (theme === 'light') return '☀️ 浅色';
    if (theme === 'dark') return '🌙 深色';
    return '🔄 自动';
  }

  // 页面加载前立即应用主题（避免闪烁）
  initTheme();

  // 暴露给全局
  window.MGHUBTheme = {
    cycle: cycleTheme,
    get: getStoredTheme,
    getLabel: getThemeLabel,
    apply: applyTheme,
  };

  // 绑定所有主题切换按钮
  document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('[data-theme-toggle]').forEach(function (btn) {
      function updateLabel() {
        btn.textContent = getThemeLabel(getStoredTheme());
        btn.title = '点击切换主题（当前：' + getThemeLabel(getStoredTheme()) + '）';
      }
      updateLabel();
      btn.addEventListener('click', function () {
        cycleTheme();
        updateLabel();
      });
    });
  });
})();
