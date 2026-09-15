/* ============================================
   导航页交互脚本
   ============================================ */

(function () {
  'use strict';

  // 布局切换
  const layoutToggle = document.querySelector('.layout-toggle');
  const navContainer = document.getElementById('nav-container');
  let currentLayout = localStorage.getItem('portal_layout') || 'grid';

  function setLayout(layout) {
    currentLayout = layout;
    localStorage.setItem('portal_layout', layout);

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
  const categoryFilter = document.getElementById('category-filter');
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
    buildCategoryFilter();
    renderNavItems(currentItems, currentLayout);
  }

  // 从导航项中动态提取分类列表，填充筛选下拉
  function buildCategoryFilter() {
    if (!categoryFilter) return;
    const cats = [];
    const seen = {};
    currentItems.forEach(function (item) {
      const cid = item.category_id;
      if (cid !== undefined && cid !== null && !seen[cid]) {
        seen[cid] = true;
        cats.push({ id: cid, name: item.category_name || '未分类' });
      }
    });
    categoryFilter.innerHTML =
      '<option value="">全部分类</option>' +
      cats
        .map(function (c) {
          return '<option value="' + c.id + '">' + escapeHtml(c.name) + '</option>';
        })
        .join('');
  }

  // 组合应用搜索关键词 + 分类筛选
  function applyFilter() {
    const keyword = searchInput ? searchInput.value.trim().toLowerCase() : '';
    const catId = categoryFilter ? categoryFilter.value : '';

    let list = currentItems;

    if (catId) {
      list = list.filter(function (item) {
        return String(item.category_id) === catId;
      });
    }

    if (keyword) {
      list = list.filter(function (item) {
        return (
          item.name.toLowerCase().includes(keyword) ||
          (item.description && item.description.toLowerCase().includes(keyword)) ||
          item.url.toLowerCase().includes(keyword) ||
          (item.category_name && item.category_name.toLowerCase().includes(keyword))
        );
      });
    }

    renderNavItems(list, currentLayout);
  }

  function renderNavItems(items, layout) {
    const container = document.getElementById('nav-container');
    if (!container) return;

    if (items.length === 0) {
      container.innerHTML =
        '<div class="empty-state"><div class="empty-icon">🔍</div><p>没有找到匹配的导航项</p></div>';
      return;
    }

     // 网格模式：所有卡片混合排列为一个完整网格，卡片自带分类标签
    if (layout === 'grid') {
      let html = '<div class="nav-grid">';
      items.forEach(function (item) {
        html += renderGridCard(item);
      });
      html += '</div>';
      container.innerHTML = html;
      return;
    }

    // 列表模式：按分类分组展示
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
      html += '<div class="nav-list">';
      catItems.forEach(function (item) {
        html += renderListItem(item);
      });
      html += '</div>';
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
      '<div class="card-cat">' +
      escapeHtml(item.category_name || '未分类') +
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
      : '<span class="tag tag-private">内网</span>';

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
      if (searchTimeout) {
        clearTimeout(searchTimeout);
      }
      searchTimeout = setTimeout(applyFilter, 200);
    });

    // 快捷键 Ctrl+K / Cmd+K 聚焦搜索
    document.addEventListener('keydown', function (e) {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        searchInput.focus();
      }
      if (e.key === 'Escape' && document.activeElement === searchInput) {
        searchInput.value = '';
        applyFilter();
        searchInput.blur();
      }
    });
  }

  // 分类筛选
  if (categoryFilter) {
    categoryFilter.addEventListener('change', applyFilter);
  }

  // 初始化
  initItems();
  setLayout(currentLayout);
})();
