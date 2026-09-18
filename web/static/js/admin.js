/* ============================================
   管理后台交互脚本
   ============================================ */

(function () {
  'use strict';

  // ========== 工具函数 ==========

  function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }

  function showToast(message, type) {
    type = type || 'success';
    let toast = document.getElementById('toast');
    if (!toast) {
      toast = document.createElement('div');
      toast.id = 'toast';
      toast.className = 'toast';
      document.body.appendChild(toast);
    }
    toast.textContent = message;
    toast.className = 'toast ' + type;
    setTimeout(function () {
      toast.classList.add('show');
    }, 10);
    setTimeout(function () {
      toast.classList.remove('show');
    }, 3000);
  }

  async function api(url, method, data) {
    const opts = {
      method: method || 'GET',
      headers: { 'Content-Type': 'application/json' },
    };
    if (data) {
      opts.body = JSON.stringify(data);
    }
    const res = await fetch(url, opts);
    const json = await res.json();
    if (!res.ok) {
      throw new Error(json.error || '请求失败');
    }
    return json;
  }

  function openModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
      modal.classList.add('show');
    }
  }

  function closeModal(id) {
    const modal = document.getElementById(id);
    if (modal) {
      modal.classList.remove('show');
    }
  }

  // 点击遮罩关闭模态框
  document.querySelectorAll('.modal-overlay').forEach(function (overlay) {
    overlay.addEventListener('click', function (e) {
      if (e.target === overlay) {
        overlay.classList.remove('show');
      }
    });
  });

  // ========== 表格排序 / 筛选 / 搜索 ==========

  // 表格状态：category / nav 两张表
  const tableState = {
    category: { search: '', sortKey: null, sortDir: 1 },
    nav: { search: '', category: '', type: '', visibility: '', sortKey: null, sortDir: 1 },
  };

  function compareValues(a, b, key) {
    const va = a[key];
    const vb = b[key];
    if (va === vb) return 0;
    if (va === null || va === undefined || va === '') return -1;
    if (vb === null || vb === undefined || vb === '') return 1;
    if (typeof va === 'number' && typeof vb === 'number') return va - vb;
    if (typeof va === 'boolean') return (va ? 1 : 0) - (vb ? 1 : 0);
    return String(va).localeCompare(String(vb), 'zh-CN');
  }

  // 给表格可排序列绑定点击事件
  function setupTableSort(tableId, stateKey) {
    const table = document.getElementById(tableId);
    if (!table) return;
    table.querySelectorAll('th.sortable').forEach(function (th) {
      th.addEventListener('click', function () {
        const key = th.dataset.sort;
        const state = tableState[stateKey];
        if (state.sortKey === key) {
          state.sortDir = -state.sortDir;
        } else {
          state.sortKey = key;
          state.sortDir = 1;
        }
        // 更新排序指示
        table.querySelectorAll('th.sortable .sort-indicator').forEach(function (ind) {
          ind.textContent = '';
        });
        const ind = th.querySelector('.sort-indicator');
        if (ind) ind.textContent = state.sortDir === 1 ? '↑' : '↓';
        if (stateKey === 'category') {
          renderCategories();
        } else {
          renderNavTable();
        }
      });
    });
  }

  // ========== 侧边栏切换 ==========

  document.querySelectorAll('.admin-nav-item').forEach(function (item) {
    item.addEventListener('click', function () {
      document.querySelectorAll('.admin-nav-item').forEach(function (i) {
        i.classList.remove('active');
      });
      item.classList.add('active');

      const target = item.dataset.target;
      document.querySelectorAll('.admin-section').forEach(function (sec) {
        sec.style.display = sec.id === target ? 'block' : 'none';
      });
      if (target === 'section-announcements' && typeof loadAnnouncements === 'function') loadAnnouncements();
    });
  });

  // ========== 分类管理 ==========

  let editingCategoryId = null;

  document.getElementById('btn-add-category').addEventListener('click', function () {
    editingCategoryId = null;
    document.getElementById('category-form').reset();
    document.getElementById('category-modal-title').textContent = '添加分类';
    openModal('category-modal');
  });

  document.getElementById('category-form').addEventListener('submit', async function (e) {
    e.preventDefault();
    const data = {
      name: document.getElementById('cat-name').value.trim(),
      icon: document.getElementById('cat-icon').value.trim(),
      sort_order: parseInt(document.getElementById('cat-sort').value) || 0,
    };
    try {
      if (editingCategoryId) {
        await api('/api/categories/' + editingCategoryId, 'PUT', data);
        showToast('分类更新成功');
      } else {
        await api('/api/categories', 'POST', data);
        showToast('分类添加成功');
      }
      closeModal('category-modal');
      loadCategories();
      loadNavItems();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  window.editCategory = function (id, name, icon, sortOrder) {
    editingCategoryId = id;
    document.getElementById('cat-name').value = name;
    document.getElementById('cat-icon').value = icon || '';
    document.getElementById('cat-sort').value = sortOrder || 0;
    document.getElementById('category-modal-title').textContent = '编辑分类';
    openModal('category-modal');
  };

  window.deleteCategory = async function (id, name) {
    if (!confirm('确定删除分类 "' + name + '" 吗？该分类下的所有导航项也会被删除。')) {
      return;
    }
    try {
      await api('/api/categories/' + id, 'DELETE');
      showToast('分类删除成功');
      loadCategories();
      loadNavItems();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  async function loadCategories() {
    try {
      const res = await api('/api/categories');
      window.categoriesData = res.data || [];

      // 更新导航项表单中的分类下拉
      const select = document.getElementById('nav-category');
      select.innerHTML = window.categoriesData
        .map(function (cat) {
          return '<option value="' + cat.id + '">' + escapeHtml(cat.name) + '</option>';
        })
        .join('');

      // 更新导航管理表格的分类筛选下拉
      const filter = document.getElementById('nav-category-filter');
      if (filter) {
        const current = tableState.nav.category;
        filter.innerHTML =
          '<option value="">全部分类</option>' +
          window.categoriesData
            .map(function (cat) {
              return '<option value="' + cat.id + '">' + escapeHtml(cat.name) + '</option>';
            })
            .join('');
        filter.value = current;
      }

      renderCategories();
    } catch (err) {
      console.error('加载分类失败:', err);
    }
  }

  function renderCategories() {
    const tbody = document.querySelector('#category-table tbody');
    let list = (window.categoriesData || []).slice();

    // 搜索
    const kw = tableState.category.search.toLowerCase();
    if (kw) {
      list = list.filter(function (cat) {
        return (cat.name || '').toLowerCase().includes(kw);
      });
    }

    // 排序
    if (tableState.category.sortKey) {
      const key = tableState.category.sortKey;
      const dir = tableState.category.sortDir;
      list.sort(function (a, b) {
        return compareValues(a, b, key) * dir;
      });
    }

    tbody.innerHTML = list
      .map(function (cat) {
        return (
          '<tr>' +
          '<td><span class="table-icon">' +
          escapeHtml(cat.icon || '📁') +
          '</span></td>' +
          '<td>' +
          escapeHtml(cat.name) +
          '</td>' +
          '<td>' +
          cat.sort_order +
          '</td>' +
          '<td class="actions">' +
          '<button class="btn btn-sm btn-outline" onclick="editCategory(' +
          cat.id +
          ", '" +
          escapeHtml(cat.name) +
          "', '" +
          escapeHtml(cat.icon || '') +
          "', " +
          cat.sort_order +
          ')">编辑</button>' +
          '<button class="btn btn-sm btn-danger" onclick="deleteCategory(' +
          cat.id +
          ", '" +
          escapeHtml(cat.name) +
          "')\">删除</button>" +
          '</td></tr>'
        );
      })
      .join('');
  }

  // 分类搜索
  const categorySearch = document.getElementById('category-search');
  if (categorySearch) {
    categorySearch.addEventListener('input', function () {
      tableState.category.search = this.value.trim();
      renderCategories();
    });
  }

  // ========== 导航项管理 ==========

  let editingNavId = null;
  let currentIconType = 'emoji';

  document.getElementById('btn-add-nav').addEventListener('click', function () {
    editingNavId = null;
    document.getElementById('nav-form').reset();
    document.getElementById('nav-modal-title').textContent = '添加导航项';
    document.getElementById('icon-preview').textContent = '🔗';
    setIconType('emoji');
    loadExistingIcons();
    openModal('nav-modal');
  });

  function setIconType(type) {
    currentIconType = type;
    document.querySelectorAll('.icon-type-tabs button').forEach(function (btn) {
      btn.classList.toggle('active', btn.dataset.type === type);
    });
    document.getElementById('emoji-input-group').style.display = type === 'emoji' ? 'block' : 'none';
    document.getElementById('image-input-group').style.display = type === 'image' ? 'block' : 'none';
  }

  document.querySelectorAll('.icon-type-tabs button').forEach(function (btn) {
    btn.addEventListener('click', function () {
      setIconType(btn.dataset.type);
    });
  });

  // emoji 输入实时预览
  document.getElementById('nav-icon-emoji').addEventListener('input', function () {
    document.getElementById('icon-preview').textContent = this.value || '🔗';
  });

  // 图片上传
  document.getElementById('upload-area').addEventListener('click', function () {
    document.getElementById('icon-file-input').click();
  });

  document.getElementById('icon-file-input').addEventListener('change', async function (e) {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
      const res = await fetch('/api/upload', {
        method: 'POST',
        body: formData,
      });
      const json = await res.json();
      if (!res.ok) throw new Error(json.error || '上传失败');

      document.getElementById('nav-icon-image').value = json.url;
      document.getElementById('icon-preview').innerHTML =
        '<img src="' + json.url + '" alt="icon">';
      showToast(json.deduped ? '图片已存在，已复用' : '图片上传成功');
      loadExistingIcons();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  // 加载已有上传图片，填充下拉选择
  function loadExistingIcons() {
    fetch('/api/uploads')
      .then(function (res) { return res.json(); })
      .then(function (json) {
        var list = json.data || [];
        var sel = document.getElementById('nav-icon-existing');
        var current = document.getElementById('nav-icon-image').value;
        sel.innerHTML = '<option value="">— 选择已有图片 —</option>';
        list.forEach(function (u) {
          var opt = document.createElement('option');
          opt.value = '/uploads/' + u.name;
          opt.textContent = u.name + '（' + (u.ref_count || 0) + ' 次引用）';
          sel.appendChild(opt);
        });
        if (current && current.indexOf('/uploads/') === 0) {
          sel.value = current;
        }
      })
      .catch(function () {});
  }

  document.getElementById('nav-icon-existing').addEventListener('change', function () {
    if (!this.value) return;
    document.getElementById('nav-icon-url').value = '';
    document.getElementById('nav-icon-image').value = this.value;
    document.getElementById('icon-preview').innerHTML =
      '<img src="' + this.value + '" alt="icon">';
  });

  document.getElementById('nav-icon-url').addEventListener('input', function () {
    var v = this.value.trim();
    if (!v) return;
    document.getElementById('nav-icon-existing').value = '';
    document.getElementById('nav-icon-image').value = v;
    document.getElementById('icon-preview').innerHTML =
      '<img src="' + v + '" alt="icon">';
  });

  document.getElementById('nav-form').addEventListener('submit', async function (e) {
    e.preventDefault();
    const iconValue =
      currentIconType === 'emoji'
        ? document.getElementById('nav-icon-emoji').value.trim()
        : document.getElementById('nav-icon-image').value.trim();

    const data = {
      category_id: parseInt(document.getElementById('nav-category').value),
      name: document.getElementById('nav-name').value.trim(),
      url: document.getElementById('nav-url').value.trim(),
      icon: iconValue,
      icon_type: currentIconType,
      description: document.getElementById('nav-desc').value.trim(),
      is_public: document.getElementById('nav-public').checked,
      visible_roles: document.getElementById('nav-visible-roles').value,
      sort_order: parseInt(document.getElementById('nav-sort').value) || 0,
    };

    if (!data.name || !data.url || !data.category_id) {
      showToast('名称、URL和分类不能为空', 'error');
      return;
    }

    try {
      if (editingNavId) {
        await api('/api/nav-items/' + editingNavId, 'PUT', data);
        showToast('导航项更新成功');
      } else {
        await api('/api/nav-items', 'POST', data);
        showToast('导航项添加成功');
      }
      closeModal('nav-modal');
      loadNavItems();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  window.editNavItem = function (id) {
    const item = window.navItemsData ? window.navItemsData[id] : null;
    if (!item) { showToast('导航项数据不存在', 'error'); return; }
    editingNavId = item.id;
    document.getElementById('nav-category').value = item.category_id;
    document.getElementById('nav-name').value = item.name;
    document.getElementById('nav-url').value = item.url;
    document.getElementById('nav-desc').value = item.description || '';
    document.getElementById('nav-public').checked = item.is_public;
    document.getElementById('nav-visible-roles').value = item.visible_roles || 'all';
    document.getElementById('nav-sort').value = item.sort_order || 0;
    document.getElementById('nav-modal-title').textContent = '编辑导航项';

    setIconType(item.icon_type || 'emoji');
    if (item.icon_type === 'image') {
      document.getElementById('nav-icon-image').value = item.icon || '';
      document.getElementById('icon-preview').innerHTML = item.icon
        ? '<img src="' + item.icon + '" alt="icon">'
        : '🖼️';
    } else {
      document.getElementById('nav-icon-emoji').value = item.icon || '';
      document.getElementById('nav-icon-image').value = '';
      document.getElementById('icon-preview').textContent = item.icon || '🔗';
    }
    document.getElementById('nav-icon-url').value = '';
    document.getElementById('nav-icon-existing').value = '';
    loadExistingIcons();
    openModal('nav-modal');
  };

  window.deleteNavItem = async function (id) {
    const item = window.navItemsData ? window.navItemsData[id] : null;
    const name = item ? item.name : '';
    if (!confirm('确定删除导航项 "' + name + '" 吗？')) return;
    try {
      await api('/api/nav-items/' + id, 'DELETE');
      showToast('导航项删除成功');
      loadNavItems();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  // 导航项可见范围显示
  const visibleRoleLabels = { all: '所有用户', admin: '仅管理员' };

  async function loadNavItems() {
    try {
      const res = await api('/api/nav-items');
      // 存储数据供编辑与筛选使用
      window.navItemsData = {};
      res.data.forEach(function (item) { window.navItemsData[item.id] = item; });
      window.navItemsList = res.data;
      renderNavTable();
    } catch (err) {
      console.error('加载导航项失败:', err);
    }
  }

  function renderNavTable() {
    const tbody = document.querySelector('#nav-table tbody');
    let list = (window.navItemsList || []).slice();

    // 分类筛选
    const catId = tableState.nav.category;
    if (catId) {
      list = list.filter(function (item) {
        return String(item.category_id) === catId;
      });
    }

    // 类型筛选
    if (tableState.nav.type) {
      list = list.filter(function (item) {
        const t = item.is_public ? 'public' : 'private';
        return t === tableState.nav.type;
      });
    }

    // 可见范围筛选
    if (tableState.nav.visibility) {
      list = list.filter(function (item) {
        return (item.visible_roles || 'all') === tableState.nav.visibility;
      });
    }

    // 搜索
    const kw = tableState.nav.search.toLowerCase();
    if (kw) {
      list = list.filter(function (item) {
        return (
          (item.name || '').toLowerCase().includes(kw) ||
          (item.url || '').toLowerCase().includes(kw) ||
          (item.description || '').toLowerCase().includes(kw) ||
          (item.category_name || '').toLowerCase().includes(kw)
        );
      });
    }

    // 排序
    if (tableState.nav.sortKey) {
      const key = tableState.nav.sortKey;
      const dir = tableState.nav.sortDir;
      list.sort(function (a, b) {
        return compareValues(a, b, key) * dir;
      });
    }

    tbody.innerHTML = list
      .map(function (item) {
        const iconHtml =
          item.icon_type === 'image' && item.icon
            ? '<img src="' + item.icon + '" alt="">'
            : escapeHtml(item.icon || '🔗');
        const tag = item.is_public
          ? '<span class="tag tag-public">公网</span>'
          : '<span class="tag tag-private">内网</span>';
        const visTag =
          '<span class="tag tag-member">' +
          (visibleRoleLabels[item.visible_roles] || '所有用户') +
          '</span>';
        return (
          '<tr>' +
          '<td><span class="table-icon">' + iconHtml + '</span></td>' +
          '<td>' + escapeHtml(item.name) + '</td>' +
          '<td>' + escapeHtml(item.category_name || '-') + '</td>' +
          '<td><a href="' + escapeHtml(item.url) + '" target="_blank" rel="noopener">' + escapeHtml(item.url) + '</a></td>' +
          '<td>' + tag + '</td>' +
          '<td>' + visTag + '</td>' +
          '<td>' + item.sort_order + '</td>' +
          '<td class="actions">' +
          '<button class="btn btn-sm btn-outline" onclick="editNavItem(' + item.id + ')">编辑</button>' +
          '<button class="btn btn-sm btn-danger" onclick="deleteNavItem(' + item.id + ')">删除</button>' +
          '</td></tr>'
        );
      })
      .join('');
  }

  // 导航搜索
  const navSearch = document.getElementById('nav-search');
  if (navSearch) {
    navSearch.addEventListener('input', function () {
      tableState.nav.search = this.value.trim();
      renderNavTable();
    });
  }

  // 导航分类筛选
  const navCatFilter = document.getElementById('nav-category-filter');
  if (navCatFilter) {
    navCatFilter.addEventListener('change', function () {
      tableState.nav.category = this.value;
      renderNavTable();
    });
  }
  const navTypeFilter = document.getElementById('nav-type-filter');
  if (navTypeFilter) {
    navTypeFilter.addEventListener('change', function () {
      tableState.nav.type = this.value;
      renderNavTable();
    });
  }
  const navVisFilter = document.getElementById('nav-visibility-filter');
  if (navVisFilter) {
    navVisFilter.addEventListener('change', function () {
      tableState.nav.visibility = this.value;
      renderNavTable();
    });
  }

  // ========== 用户管理（超管） ==========

  let editingUserId = null;

  const btnAddUser = document.getElementById('btn-add-user');
  if (btnAddUser) {
    btnAddUser.addEventListener('click', function () {
      editingUserId = null;
      document.getElementById('user-form').reset();
      document.getElementById('user-modal-title').textContent = '添加用户';
      document.getElementById('user-password-group').style.display = 'block';
      document.getElementById('user-password').required = true;
      document.getElementById('user-categories-group').style.display = 'none';
      document.getElementById('user-role').disabled = false;
      document.getElementById('user-username').disabled = false;
      openModal('user-modal');
    });
  }

  const userForm = document.getElementById('user-form');
  if (userForm) {
    userForm.addEventListener('submit', async function (e) {
      e.preventDefault();
      const data = {
        username: document.getElementById('user-username').value.trim(),
        role: document.getElementById('user-role').value,
      };

      try {
        if (editingUserId) {
          await api('/api/users/' + editingUserId, 'PUT', data);
          // 保存分类权限
          const categoryIDs = [];
          document.querySelectorAll('#user-categories-list input[type="checkbox"]:checked').forEach(function (cb) {
            categoryIDs.push(parseInt(cb.value));
          });
          await api('/api/users/' + editingUserId + '/categories', 'PUT', { category_ids: categoryIDs });
          showToast('用户更新成功');
        } else {
          data.password = document.getElementById('user-password').value;
          if (!data.password || data.password.length < 6) {
            showToast('密码至少6位', 'error');
            return;
          }
          await api('/api/users', 'POST', data);
          showToast('用户添加成功');
        }
        closeModal('user-modal');
        loadUsers();
      } catch (err) {
        showToast(err.message, 'error');
      }
    });
  }

  window.editUser = async function (id) {
    const user = window.usersData ? window.usersData[id] : null;
    if (!user) { showToast('用户数据不存在', 'error'); return; }
    editingUserId = user.id;
    document.getElementById('user-username').value = user.username;
    document.getElementById('user-username').disabled = true; // 用户名不可修改
    document.getElementById('user-role').value = user.role;
    document.getElementById('user-modal-title').textContent = '编辑用户';
    document.getElementById('user-password-group').style.display = 'none';
    document.getElementById('user-password').required = false;

    // 超管不可修改角色，且不受分类权限限制
    const isSuper = user.role === 'super_admin';
    document.getElementById('user-role').disabled = isSuper;
    document.getElementById('user-categories-group').style.display = isSuper ? 'none' : 'block';

    // 加载所有分类并渲染复选框
    try {
      const catsRes = await api('/api/categories');
      const permRes = await api('/api/users/' + user.id + '/categories');
      const allowedIDs = new Set(permRes.data || []);
      const list = document.getElementById('user-categories-list');
      list.innerHTML = catsRes.data.map(function (cat) {
        const checked = allowedIDs.has(cat.id) ? 'checked' : '';
        return '<label style="display:inline-flex;align-items:center;gap:4px;padding:6px 10px;background:var(--border-light);border-radius:6px;cursor:pointer;font-size:0.85rem;">' +
          '<input type="checkbox" value="' + cat.id + '" ' + checked + '>' +
          escapeHtml(cat.icon || '📁') + ' ' + escapeHtml(cat.name) +
          '</label>';
      }).join('');
    } catch (err) {
      console.error('加载分类权限失败:', err);
    }

    openModal('user-modal');
  };

  window.resetUserPassword = async function (id) {
    const user = window.usersData ? window.usersData[id] : null;
    const username = user ? user.username : '';
    const newPassword = prompt('请输入用户 "' + username + '" 的新密码（至少6位）:');
    if (!newPassword) return;
    if (newPassword.length < 6) {
      showToast('密码至少6位', 'error');
      return;
    }
    try {
      await api('/api/users/' + id + '/reset-password', 'POST', { password: newPassword });
      showToast('密码重置成功');
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  window.toggleUserStatus = async function (id) {
    const user = window.usersData ? window.usersData[id] : null;
    if (!user) { showToast('用户数据不存在', 'error'); return; }
    const newStatus = user.status === 'active' ? 'disabled' : 'active';
    const action = newStatus === 'active' ? '启用' : '禁用';
    if (!confirm('确定' + action + '用户 "' + user.username + '" 吗？')) return;
    try {
      await api('/api/users/' + id, 'PUT', { status: newStatus });
      showToast('用户已' + action);
      loadUsers();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  window.deleteUser = async function (id) {
    const user = window.usersData ? window.usersData[id] : null;
    const username = user ? user.username : '';
    if (!confirm('确定删除用户 "' + username + '" 吗？此操作不可恢复。')) return;
    try {
      await api('/api/users/' + id, 'DELETE');
      showToast('用户删除成功');
      loadUsers();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  async function loadUsers() {
    try {
      const res = await api('/api/users');
      // 存储数据供编辑使用
      window.usersData = {};
      res.data.forEach(function (user) { window.usersData[user.id] = user; });
      const tbody = document.querySelector('#user-table tbody');
      tbody.innerHTML = res.data
        .map(function (user) {
          const roleTag =
            user.role === 'super_admin'
              ? '<span class="tag tag-super">超管</span>'
              : user.role === 'admin'
                ? '<span class="tag tag-admin">管理员</span>'
                : '<span class="tag tag-member">成员</span>';
          const statusTag =
            user.status === 'active'
              ? '<span class="tag tag-public">正常</span>'
              : '<span class="tag tag-private">已禁用</span>';
          const isSuper = user.role === 'super_admin';
          return (
            '<tr>' +
            '<td><div class="avatar" style="display:inline-flex;width:32px;height:32px;border-radius:50%;background:var(--primary);color:#fff;align-items:center;justify-content:center;font-size:0.8rem;font-weight:600;">' +
            escapeHtml(user.username.charAt(0).toUpperCase()) +
            '</div></td>' +
            '<td>' +
            escapeHtml(user.username) +
            '</td>' +
            '<td>' +
            roleTag +
            '</td>' +
            '<td>' +
            statusTag +
            '</td>' +
            '<td class="actions">' +
            (isSuper ? '' : '<button class="btn btn-sm btn-outline" onclick="editUser(' + user.id + ')">编辑</button>') +
            '<button class="btn btn-sm btn-ghost" onclick="resetUserPassword(' + user.id + ')">重置密码</button>' +
            (isSuper ? '' : '<button class="btn btn-sm btn-ghost" onclick="toggleUserStatus(' + user.id + ')">' + (user.status === 'active' ? '禁用' : '启用') + '</button>') +
            (isSuper ? '' : '<button class="btn btn-sm btn-danger" onclick="deleteUser(' + user.id + ')">删除</button>') +
            '</td></tr>'
          );
        })
        .join('');
    } catch (err) {
      console.error('加载用户失败:', err);
    }
  }

  // ========== 站点设置（基本信息） ==========

  async function saveSettings() {
    const data = {
      site_title: document.getElementById('set-site-title').value.trim(),
      site_subtitle: document.getElementById('set-site-subtitle').value.trim(),
      logo_text: document.getElementById('set-logo-text').value.trim(),
      footer_text: document.getElementById('set-footer-text').value.trim(),
      wiki_url: document.getElementById('set-wiki-url').value.trim(),
      home_url: document.getElementById('set-home-url').value.trim(),
      home_badge: document.getElementById('set-home-badge').value.trim(),
      home_badge_icon: (document.getElementById('set-home-badge-icon') || {value:''}).value.trim(),
      custom_primary: (document.getElementById('set-custom-primary') || {value:''}).value.trim(),
      custom_accent: (document.getElementById('set-custom-accent') || {value:''}).value.trim(),
      custom_bg: (document.getElementById('set-custom-bg') || {value:''}).value.trim(),
      custom_theme_enabled: (document.getElementById('set-custom-theme-enabled') && document.getElementById('set-custom-theme-enabled').checked) ? '1' : '',
      home_hero_title: document.getElementById('set-home-hero-title').value.trim(),
      home_hero_sub: document.getElementById('set-home-hero-sub').value.trim(),
      logo_light: document.getElementById('set-logo-light').value.trim(),
      logo_dark: document.getElementById('set-logo-dark').value.trim(),
      favicon: document.getElementById('set-favicon').value.trim(),
      default_theme: document.getElementById('set-default-theme').value,
    };
    try {
      await api('/api/settings', 'POST', data);
      showToast('基本设置已保存');
    } catch (err) {
      showToast(err.message, 'error');
    }
  }

  const btnSaveSettings = document.getElementById('btn-save-settings');
  if (btnSaveSettings) btnSaveSettings.addEventListener('click', saveSettings);

  // ========== Logo / Favicon 上传 ==========

  function setupLogoUpload(uploadAreaId, hiddenInputId, previewId) {
    const area = document.getElementById(uploadAreaId);
    const input = area.querySelector('input[type="file"]');
    area.addEventListener('click', function () { input.click(); });
    input.addEventListener('change', async function (e) {
      const file = e.target.files[0];
      if (!file) return;
      const formData = new FormData();
      formData.append('file', file);
      try {
        const res = await fetch('/api/upload', { method: 'POST', body: formData });
        const json = await res.json();
        if (!res.ok) throw new Error(json.error || '上传失败');
        document.getElementById(hiddenInputId).value = json.url;
        document.getElementById(previewId).innerHTML = '<img src="' + json.url + '">';
        showToast(json.deduped ? '图片已存在，已复用' : '上传成功');
      } catch (err) {
        showToast(err.message, 'error');
      }
    });
  }

  setupLogoUpload('upload-logo-light', 'set-logo-light', 'preview-logo-light');
  setupLogoUpload('upload-logo-dark', 'set-logo-dark', 'preview-logo-dark');
  setupLogoUpload('upload-favicon', 'set-favicon', 'preview-favicon');

  // ========== 愿景要点动态管理 ==========

  let editingVisionId = null;

  document.getElementById('btn-add-vision').addEventListener('click', function () {
    editingVisionId = null;
    document.getElementById('vision-form').reset();
    document.getElementById('vision-modal-title').textContent = '添加愿景要点';
    openModal('vision-modal');
  });

  document.getElementById('vision-form').addEventListener('submit', async function (e) {
    e.preventDefault();
    const data = {
      content: document.getElementById('vision-content').value.trim(),
      sort_order: parseInt(document.getElementById('vision-sort').value) || 0,
    };
    if (!data.content) { showToast('愿景内容不能为空', 'error'); return; }
    try {
      if (editingVisionId) {
        await api('/api/visions/' + editingVisionId, 'PUT', data);
        showToast('愿景更新成功');
      } else {
        await api('/api/visions', 'POST', data);
        showToast('愿景添加成功');
      }
      closeModal('vision-modal');
      loadVisions();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  window.editVision = function (id, content, sortOrder) {
    editingVisionId = id;
    document.getElementById('vision-content').value = content;
    document.getElementById('vision-sort').value = sortOrder || 0;
    document.getElementById('vision-modal-title').textContent = '编辑愿景要点';
    openModal('vision-modal');
  };

  window.deleteVision = async function (id, content) {
    if (!confirm('确定删除这条愿景要点吗？')) return;
    try {
      await api('/api/visions/' + id, 'DELETE');
      showToast('愿景删除成功');
      loadVisions();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  async function loadVisions() {
    try {
      const res = await api('/api/visions');
      const list = document.getElementById('vision-list');
      if (!res.data || res.data.length === 0) {
        list.innerHTML = '<div style="color:var(--text-muted);font-size:0.85rem;padding:12px;">暂无愿景要点，点击上方"添加愿景"按钮添加</div>';
        return;
      }
      list.innerHTML = res.data.map(function (v) {
        return (
          '<div class="dynamic-item">' +
          '<div class="item-icon">✦</div>' +
          '<div class="item-name" style="flex:2;">' + escapeHtml(v.content) + '</div>' +
          '<div style="font-size:0.75rem;color:var(--text-muted);flex-shrink:0;">排序:' + v.sort_order + '</div>' +
          '<div class="item-actions">' +
          '<button class="btn btn-sm btn-outline" onclick="editVision(' + v.id + ", '" + escapeHtml(v.content).replace(/'/g, "\\'") + "', " + v.sort_order + ')">编辑</button>' +
          '<button class="btn btn-sm btn-danger" onclick="deleteVision(' + v.id + ", '" + escapeHtml(v.content).replace(/'/g, "\\'") + "')\">删除</button>" +
          '</div></div>'
        );
      }).join('');
    } catch (err) {
      console.error('加载愿景失败:', err);
    }
  }

  // ========== 链接动态管理 ==========

  let editingLinkId = null;
  const linkTypeLabels = { header: '顶部导航', cta: '首页按钮', footer: '页脚' };

  document.getElementById('btn-add-link').addEventListener('click', function () {
    editingLinkId = null;
    document.getElementById('link-form').reset();
    document.getElementById('link-modal-title').textContent = '添加链接';
    openModal('link-modal');
  });

  document.getElementById('link-form').addEventListener('submit', async function (e) {
    e.preventDefault();
    const data = {
      name: document.getElementById('link-name').value.trim(),
      url: document.getElementById('link-url').value.trim(),
      icon: document.getElementById('link-icon').value.trim(),
      link_type: document.getElementById('link-type').value,
      sort_order: parseInt(document.getElementById('link-sort').value) || 0,
    };
    if (!data.name || !data.url) { showToast('名称和URL不能为空', 'error'); return; }
    try {
      if (editingLinkId) {
        await api('/api/links/' + editingLinkId, 'PUT', data);
        showToast('链接更新成功');
      } else {
        await api('/api/links', 'POST', data);
        showToast('链接添加成功');
      }
      closeModal('link-modal');
      loadLinks();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });

  // 编辑链接：通过 ID 从缓存取数据，避免 JSON 序列化放入 onclick 的转义问题
  window.editLink = function (id) {
    const link = window.linksData ? window.linksData[id] : null;
    if (!link) { showToast('链接数据不存在', 'error'); return; }
    editingLinkId = link.id;
    document.getElementById('link-name').value = link.name;
    document.getElementById('link-url').value = link.url;
    document.getElementById('link-icon').value = link.icon || '';
    document.getElementById('link-type').value = link.link_type || 'header';
    document.getElementById('link-sort').value = link.sort_order || 0;
    document.getElementById('link-modal-title').textContent = '编辑链接';
    openModal('link-modal');
  };

  window.deleteLink = async function (id) {
    const link = window.linksData ? window.linksData[id] : null;
    const name = link ? link.name : '';
    if (!confirm('确定删除链接 "' + name + '" 吗？')) return;
    try {
      await api('/api/links/' + id, 'DELETE');
      showToast('链接删除成功');
      loadLinks();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  async function loadLinks() {
    try {
      const res = await api('/api/links');
      // 存储数据供编辑使用
      window.linksData = {};
      res.data.forEach(function (l) { window.linksData[l.id] = l; });
      const list = document.getElementById('link-list');
      if (!res.data || res.data.length === 0) {
        list.innerHTML = '<div style="color:var(--text-muted);font-size:0.85rem;padding:12px;">暂无链接，点击上方"添加链接"按钮添加</div>';
        return;
      }
      list.innerHTML = res.data.map(function (l) {
        return (
          '<div class="dynamic-item">' +
          '<div class="item-icon">' + escapeHtml(l.icon || '🔗') + '</div>' +
          '<div class="item-name">' + escapeHtml(l.name) + '</div>' +
          '<div class="item-url">' + escapeHtml(l.url) + '</div>' +
          '<span class="tag tag-member" style="flex-shrink:0;">' + (linkTypeLabels[l.link_type] || l.link_type) + '</span>' +
          '<div class="item-actions">' +
          '<button class="btn btn-sm btn-outline" onclick="editLink(' + l.id + ')">编辑</button>' +
          '<button class="btn btn-sm btn-danger" onclick="deleteLink(' + l.id + ')">删除</button>' +
          '</div></div>'
        );
      }).join('');
    } catch (err) {
      console.error('加载链接失败:', err);
    }
  }

  // ========== 图片管理 ==========

  async function loadUploads() {
    const list = document.getElementById('upload-list');
    if (!list) return;
    try {
      const res = await api('/api/uploads');
      if (!res.data || res.data.length === 0) {
        list.innerHTML = '<div style="color:var(--text-muted);font-size:0.85rem;padding:12px;">暂无上传图片，可在导航项图标、Logo 或 Favicon 中上传</div>';
        return;
      }
      list.innerHTML = '<div class="gallery-grid">' + res.data.map(function (u) {
        const size = u.size >= 1048576
          ? (u.size / 1048576).toFixed(1) + ' MB'
          : (u.size >= 1024 ? Math.round(u.size / 1024) + ' KB' : u.size + ' B');
        const time = u.created_at ? new Date(u.created_at).toLocaleString() : '';
        const refTag = u.ref_count > 0
          ? '<span class="tag tag-public">被引用 ' + u.ref_count + ' 处</span>'
          : '<span class="tag tag-member">未引用</span>';
        const delBtn = u.ref_count > 0
          ? '<button class="btn btn-sm btn-ghost" disabled title="该图片正被引用，无法删除">删除</button>'
          : '<button class="btn btn-sm btn-danger" onclick="deleteUpload(\'' + u.name.replace(/'/g, "\\'") + '\')">删除</button>';
        const url = '/uploads/' + encodeURIComponent(u.name);
        return (
          '<div class="gallery-card">' +
          '<div class="gallery-thumb"><img src="' + url + '" alt=""></div>' +
          '<div class="gallery-body">' +
          '<div class="gallery-name" title="' + escapeHtml(u.name) + '">' + escapeHtml(u.name) + '</div>' +
          '<div class="gallery-meta">' + size + (time ? ' · ' + time : '') + '</div>' +
          '<div class="gallery-tags">' + refTag + '</div>' +
          '<div class="gallery-actions">' +
          '<button class="btn btn-sm btn-outline" onclick="copyImageUrl(\'' + url.replace(/'/g, "\\'") + '\')">复制链接</button>' +
          delBtn +
          '</div>' +
          '</div></div>'
        );
      }).join('') + '</div>';
    } catch (err) {
      list.innerHTML = '<div style="color:var(--text-muted);font-size:0.85rem;padding:12px;">加载失败: ' + escapeHtml(err.message) + '</div>';
    }
  }

  window.copyImageUrl = async function (url) {
    try {
      await navigator.clipboard.writeText(url);
      showToast('链接已复制: ' + url);
    } catch (err) {
      prompt('复制图片链接:', url);
    }
  };

  window.deleteUpload = async function (name) {
    if (!confirm('确定删除图片 "' + name + '" 吗？此操作不可恢复。')) return;
    try {
      await api('/api/uploads/' + encodeURIComponent(name), 'DELETE');
      showToast('图片删除成功');
      loadUploads();
    } catch (err) {
      showToast(err.message, 'error');
    }
  };

  const btnRefreshUploads = document.getElementById('btn-refresh-uploads');
  if (btnRefreshUploads) btnRefreshUploads.addEventListener('click', function () { showToast('已刷新'); loadUploads(); });

  // ========== 初始化 ==========

  setupTableSort('category-table', 'category');
  setupTableSort('nav-table', 'nav');

  loadCategories();
  loadNavItems();
  if (document.getElementById('user-table')) { loadUsers(); }
  if (document.getElementById('vision-list')) { loadVisions(); }
  if (document.getElementById('link-list')) { loadLinks(); }
  if (document.getElementById('upload-list')) { loadUploads(); }
})();


function esc(s){ var d=document.createElement('div'); d.textContent=s==null?'':String(s); return d.innerHTML; }

// ==================== 公告管理 ====================
let announcements = [];

async function loadAnnouncements() {
  try {
    const res = await fetch('/api/announcements?all=1');
    const json = await res.json();
    announcements = json.data || [];
    renderAnnouncements();
  } catch(e) { console.error('loadAnnouncements', e); }
}

function renderAnnouncements() {
  const el = document.getElementById('announcement-list');
  if (!el) return;
  if (announcements.length === 0) {
    el.innerHTML = '<p style="color:var(--text-muted);padding:12px;">暂无公告</p>';
    return;
  }
  el.innerHTML = '<table class="admin-table"><thead><tr><th>ID</th><th>标题</th><th>级别</th><th>启用</th><th>显示位置</th><th>时间窗</th><th>操作</th></tr></thead><tbody>' +
    announcements.map(function(a) {
      const levelLabel = {info:'普通', warn:'警告', error:'紧急'}[a.level] || a.level;
      const enabled = a.enabled ? '<span style="color:green;">是</span>' : '<span style="color:red;">否</span>';
      const where = a.show_on_home ? '首页+导航页' : '导航页';
      const win = (a.starts_at || '—') + ' ~ ' + (a.ends_at || '—');
      return '<tr>' +
        '<td>' + a.id + '</td>' +
        '<td>' + esc(a.title) + '</td>' +
        '<td>' + levelLabel + '</td>' +
        '<td>' + enabled + '</td>' +
        '<td>' + where + '</td>' +
        '<td style="font-size:12px;">' + win + '</td>' +
        '<td>' +
          '<button class="btn btn-sm" onclick="openAnnouncementModal(' + a.id + ')">编辑</button> ' +
          '<button class="btn btn-sm btn-danger" onclick="deleteAnnouncement(' + a.id + ')">删除</button>' +
        '</td>' +
      '</tr>';
    }).join('') + '</tbody></table>';
}

function openAnnouncementModal(id) {
  const a = id ? announcements.find(function(x){return x.id===id;}) : null;
  const html = '<div class="modal-overlay" onclick="if(event.target===this)closeModal()">' +
    '<div class="modal-box" style="max-width:560px;">' +
      '<div class="modal-header"><h3>' + (a ? '编辑公告' : '新建公告') + '</h3><button class="modal-close" onclick="closeModal()">✕</button></div>' +
      '<div class="modal-body">' +
        '<div class="form-group"><label class="form-label">标题</label><input id="ann-title" class="form-input" value="' + esc(a?a.title:'') + '" placeholder="公告标题"></div>' +
        '<div class="form-group"><label class="form-label">内容</label><textarea id="ann-content" class="form-input" rows="3" placeholder="公告内容（纯文本）">' + esc(a?a.content:'') + '</textarea></div>' +
        '<div class="form-group"><label class="form-label">级别</label>' +
          '<select id="ann-level" class="form-select">' +
            '<option value="info"' + (a&&a.level==='info'?' selected':'') + '>普通（蓝）</option>' +
            '<option value="warn"' + (a&&a.level==='warn'?' selected':'') + '>警告（橙）</option>' +
            '<option value="error"' + (a&&a.level==='error'?' selected':'') + '>紧急（红）</option>' +
          '</select>' +
        '</div>' +
        '<div class="form-row">' +
          '<label><input type="checkbox" id="ann-enabled"' + (a?a.enabled:true?' checked':'') + '> 启用</label>' +
          '<label style="margin-left:16px;"><input type="checkbox" id="ann-dismissible"' + (a?a.dismissible:true?' checked':'') + '> 允许成员关闭</label>' +
          '<label style="margin-left:16px;"><input type="checkbox" id="ann-showhome"' + (a&&a.show_on_home?' checked':'') + '> 首页也显示</label>' +
        '</div>' +
        '<div class="form-row">' +
          '<div class="form-group" style="flex:1;"><label class="form-label">上线时间（可空）</label><input id="ann-starts" class="form-input" value="' + esc(a?a.starts_at:'') + '" placeholder="YYYY-MM-DD HH:MM:SS"></div>' +
          '<div class="form-group" style="flex:1;"><label class="form-label">下线时间（可空）</label><input id="ann-ends" class="form-input" value="' + esc(a?a.ends_at:'') + '" placeholder="YYYY-MM-DD HH:MM:SS"></div>' +
        '</div>' +
      '</div>' +
      '<div class="modal-footer">' +
        '<button class="btn" onclick="closeModal()">取消</button>' +
        '<button class="btn btn-primary" onclick="saveAnnouncement(' + (id||0) + ')">保存</button>' +
      '</div>' +
    '</div></div>';
  const div = document.createElement('div');
  div.innerHTML = html;
  div.id = 'ann-modal-wrap';
  document.body.appendChild(div);
}

async function saveAnnouncement(id) {
  const data = {
    title: document.getElementById('ann-title').value.trim(),
    content: document.getElementById('ann-content').value.trim(),
    level: document.getElementById('ann-level').value,
    enabled: document.getElementById('ann-enabled').checked,
    dismissible: document.getElementById('ann-dismissible').checked,
    show_on_home: document.getElementById('ann-showhome').checked,
    starts_at: document.getElementById('ann-starts').value.trim(),
    ends_at: document.getElementById('ann-ends').value.trim()
  };
  if (!data.title) { alert('标题不能为空'); return; }
  const url = id ? '/api/announcements/' + id : '/api/announcements';
  const method = id ? 'PUT' : 'POST';
  const res = await fetch(url, {method: method, headers:{'Content-Type':'application/json'}, body: JSON.stringify(data)});
  const json = await res.json();
  if (json.error) { alert(json.error); return; }
  closeModal();
  loadAnnouncements();
}

async function deleteAnnouncement(id) {
  if (!confirm('确定删除这条公告？')) return;
  await fetch('/api/announcements/' + id, {method:'DELETE'});
  loadAnnouncements();
}
