/* ============================================
   MGHUB 导航页交互脚本
   ============================================ */

(function () {
  'use strict';

  // 布局切换
  const layoutToggle = document.querySelector('.layout-toggle');
  const navContainer = document.getElementById('nav-container');
  let currentLayout = localStorage.getItem('mghub_layout') || 'grid';

  function setLayout(layout) {
    currentLayout = layout;
    localStorage.setItem('mghub_layout', layout);

    // 更新按钮状态
    if (layoutToggle) {
      layoutToggle.querySelectorAll('button').forEach(function (btn) {
        btn.classList.toggle('active', btn.dataset.layout === layout);
      });
    }

    // 重新渲染
    renderNavItems(currentItems, layout);
  }

  if (layoutToggle) {
    layoutToggle.querySelectorAll('button').forEach(function (btn) {
      btn.addEventListener('click', function () {
        setLayout(btn.dataset.layout);
      });
    });
  }

  // 用户菜单
  const userMenuBtn = document.querySelector('.user-menu-btn');
  const userDropdown = document.querySelector('.user-dropdown');

  if (userMenuBtn && userDropdown) {
    userMenuBtn.addEventListener('click', function (e) {
      e.stopPropagation();
      userDropdown.classList.toggle('show');
    });

    document.addEventListener('click', function (e) {
      if (!userDropdown.contains(e.target) && !userMenuBtn.contains(e.target)) {
        userDropdown.classList.remove('show');
      }
    });
  }

  // 搜索功能
  const searchInput = document.getElementById('search-input');
  let currentItems = [];
  let searchTimeout = null;

  // 从页面数据初始化
  function initItems() {
    const itemsData = document.getElementById('nav-data');
    if (itemsData) {
      try {
        currentItems = JSON.parse(itemsData.textContent);
      } catch (e) {
        currentItems = [];
      }
    }
    renderNavItems(currentItems, currentLayout);
  }

  function renderNavItems(items, layout) {
    const container = document.getElementById('nav-container');
    if (!container) return;

    if (items.length === 0) {
      container.innerHTML =
        '<div class="empty-state"><div class="empty-icon">🔍</div><p>没有找到匹配的导航项</p></div>';
      return;
    }

    // 按分类分组
    const grouped = {};
    items.forEach(function (item) {
      const catName = item.category_name || '未分类';
      if (!grouped[catName]) {
        grouped[catName] = [];
      }
      grouped[catName].push(item);
    });

    let html = '';
    for (const catName in grouped) {
      const catItems = grouped[catName];
      html += '<div class="nav-category">';
      html +=
        '<div class="nav-category-header"><h2>' +
        escapeHtml(catName) +
        '</h2><span class="cat-count">' +
        catItems.length +
        ' 项</span></div>';

      if (layout === 'grid') {
        html += '<div class="nav-grid">';
        catItems.forEach(function (item) {
          html += renderGridCard(item);
        });
        html += '</div>';
      } else {
        html += '<div class="nav-list">';
        catItems.forEach(function (item) {
          html += renderListItem(item);
        });
        html += '</div>';
      }

      html += '</div>';
    }

    container.innerHTML = html;
  }

  function renderGridCard(item) {
    const iconHtml = renderIcon(item);
    const tagHtml = item.is_public
      ? '<span class="tag tag-public">公网</span>'
      : '<span class="tag tag-private">内网</span>';

    return (
      '<a href="' +
      escapeHtml(item.url) +
      '" target="_blank" rel="noopener noreferrer" class="nav-card">' +
      '<div class="card-icon">' +
      iconHtml +
      '</div>' +
      '<div class="card-body">' +
      '<div class="card-name">' +
      escapeHtml(item.name) +
      '</div>' +
      '<div class="card-desc">' +
      escapeHtml(item.description || '') +
      '</div>' +
      '</div>' +
      '<div class="card-tag">' +
      tagHtml +
      '</div>' +
      '</a>'
    );
  }

  function renderListItem(item) {
    const iconHtml = renderIcon(item);
    const tagHtml = item.is_public
      ? '<span class="tag tag-public">公网</span>'
      : '<span class="tag tag-private">需Tailscale</span>';

    return (
      '<a href="' +
      escapeHtml(item.url) +
      '" target="_blank" rel="noopener noreferrer" class="nav-list-item">' +
      '<div class="item-icon">' +
      iconHtml +
      '</div>' +
      '<div class="item-info">' +
      '<div class="item-name">' +
      escapeHtml(item.name) +
      '</div>' +
      '<div class="item-desc">' +
      escapeHtml(item.description || '') +
      '</div>' +
      '</div>' +
      '<div class="item-url">' +
      escapeHtml(item.url) +
      '</div>' +
      tagHtml +
      '</a>'
    );
  }

  function renderIcon(item) {
    if (item.icon_type === 'image' && item.icon) {
      return '<img src="' + escapeHtml(item.icon) + '" alt="' + escapeHtml(item.name) + '">';
    }
    return escapeHtml(item.icon || '🔗');
  }

  function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }

  // 搜索
  if (searchInput) {
    searchInput.addEventListener('input', function () {
      const keyword = searchInput.value.trim().toLowerCase();

      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }

      searchTimeout = setTimeout(function () {
        if (!keyword) {
          renderNavItems(currentItems, currentLayout);
          return;
        }

        const filtered = currentItems.filter(function (item) {
          return (
            item.name.toLowerCase().includes(keyword) ||
            (item.description && item.description.toLowerCase().includes(keyword)) ||
            item.url.toLowerCase().includes(keyword) ||
            (item.category_name && item.category_name.toLowerCase().includes(keyword))
          );
        });

        renderNavItems(filtered, currentLayout);
      }, 200);
    });

    // 快捷键 Ctrl+K / Cmd+K 聚焦搜索
    document.addEventListener('keydown', function (e) {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        searchInput.focus();
      }
      if (e.key === 'Escape' && document.activeElement === searchInput) {
        searchInput.value = '';
        renderNavItems(currentItems, currentLayout);
        searchInput.blur();
      }
    });
  }

  // 初始化
  initItems();
  setLayout(currentLayout);
})();
