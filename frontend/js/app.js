const API_BASE = '';

let token = localStorage.getItem('token');
let refreshTimer = null;
const TOKEN_REFRESH_INTERVAL = 10 * 60 * 1000; // 10分钟刷新一次

function startTokenRefresh() {
    if (refreshTimer) {
        clearInterval(refreshTimer);
    }
    refreshTimer = setInterval(async () => {
        try {
            await apiRequest('/api/admin/refresh-token', 'POST');
        } catch (error) {
            console.error('Failed to refresh token:', error);
            // 刷新失败，清除token并跳转到登录页
            token = null;
            localStorage.removeItem('token');
            if (refreshTimer) {
                clearInterval(refreshTimer);
                refreshTimer = null;
            }
            showLogin();
            showError('登录已过期，请重新登录');
        }
    }, TOKEN_REFRESH_INTERVAL);
}

function stopTokenRefresh() {
    if (refreshTimer) {
        clearInterval(refreshTimer);
        refreshTimer = null;
    }
}

function showApp() {
    document.getElementById('login-page').classList.add('hidden');
    document.getElementById('app-page').classList.remove('hidden');
    startTokenRefresh();
}

function showLogin() {
    document.getElementById('login-page').classList.remove('hidden');
    document.getElementById('app-page').classList.add('hidden');
    stopTokenRefresh();
}

function navigateTo(page) {
    document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));
    document.querySelectorAll('.sidebar-menu a').forEach(a => a.classList.remove('active'));
    
    document.getElementById(`${page}-page`).classList.remove('hidden');
    document.querySelector(`[data-page="${page}"]`)?.classList.add('active');
    
    loadPageData(page);
}

async function apiRequest(endpoint, method = 'GET', body = null) {
    const headers = {
        'Content-Type': 'application/json'
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    
    const options = { method, headers };
    if (body) {
        options.body = JSON.stringify(body);
    }
    
    const response = await fetch(`${API_BASE}${endpoint}`, options);
    const data = await response.json();
    
    if (response.status === 401) {
        token = null;
        localStorage.removeItem('token');
        showLogin();
        throw new Error('登录已过期');
    }
    
    if (!response.ok) {
        throw new Error(data.error || '请求失败');
    }
    
    return data;
}

document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const submitBtn = e.target.querySelector('button[type="submit"]');
    
    // 防重复提交
    if (submitBtn.disabled) return;
    
    // 简单的客户端验证
    if (!username.trim() || !password.trim()) {
        showError('请输入用户名和密码');
        return;
    }
    
    submitBtn.disabled = true;
    submitBtn.textContent = '登录中...';
    
    try {
        const data = await apiRequest('/api/login', 'POST', { username, password });
        token = data.token;
        localStorage.setItem('token', token);
        showSuccess('登录成功');
        showApp();
        navigateTo('dashboard');
    } catch (error) {
        showError(error.message);
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = '登录';
    }
});

document.getElementById('change-password-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;
    const confirmPassword = document.getElementById('confirm-password').value;
    const submitBtn = e.target.querySelector('button[type="submit"]');
    
    // 防重复提交
    if (submitBtn.disabled) return;
    
    if (newPassword !== confirmPassword) {
        showError('两次输入的密码不一致');
        return;
    }
    
    // 密码强度验证
    const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[^a-zA-Z0-9]).{8,}$/;
    if (!passwordRegex.test(newPassword)) {
        showError('密码必须至少8位，包含大小写字母、数字和特殊字符');
        return;
    }
    
    submitBtn.disabled = true;
    submitBtn.textContent = '修改中...';
    
    try {
        await apiRequest('/api/admin/change-password', 'POST', {
            old_password: oldPassword,
            new_password: newPassword,
            confirm_password: confirmPassword
        });
        showSuccess('密码修改成功，请重新登录');
        setTimeout(() => {
            document.getElementById('logout-btn').click();
        }, 1500);
    } catch (error) {
        showError(error.message);
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = '修改密码';
    }
});

document.getElementById('change-username-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const password = document.getElementById('username-password').value;
    const newUsername = document.getElementById('new-username').value;
    const submitBtn = e.target.querySelector('button[type="submit"]');
    
    if (submitBtn.disabled) return;
    
    if (!newUsername.trim()) {
        showError('新用户名不能为空');
        return;
    }
    
    if (newUsername.trim() === password.trim()) {
        showError('新用户名不能与密码相同');
        return;
    }
    
    submitBtn.disabled = true;
    submitBtn.textContent = '修改中...';
    
    try {
        await apiRequest('/api/admin/change-username', 'POST', {
            new_username: newUsername,
            password: password
        });
        showSuccess('用户名修改成功，请重新登录');
        setTimeout(() => {
            document.getElementById('logout-btn').click();
        }, 1500);
    } catch (error) {
        showError(error.message);
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = '修改用户名';
    }
});

// 友好的错误提示函数
function showError(message) {
    // 移除已有的提示
    removeNotifications();
    
    const notification = document.createElement('div');
    notification.className = 'notification notification-error';
    notification.innerHTML = `
        <span>${escapeHtml(message)}</span>
        <button onclick="this.parentElement.remove()">&times;</button>
    `;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}

// 成功提示函数
function showSuccess(message) {
    removeNotifications();
    
    const notification = document.createElement('div');
    notification.className = 'notification notification-success';
    notification.innerHTML = `
        <span>${escapeHtml(message)}</span>
        <button onclick="this.parentElement.remove()">&times;</button>
    `;
    document.body.appendChild(notification);
    
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 3000);
}

function removeNotifications() {
    document.querySelectorAll('.notification').forEach(n => n.remove());
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

document.querySelectorAll('.sidebar-menu a[data-page]').forEach(link => {
    link.addEventListener('click', (e) => {
        e.preventDefault();
        navigateTo(link.dataset.page);
    });
});

let models = [];
let sources = [];
let apiKeys = [];
let usageStats = null;
let serverInfo = null;

async function loadPageData(page) {
    switch(page) {
        case 'dashboard':
            await loadDashboard();
            break;
        case 'models':
            await loadModels();
            break;
        case 'sources':
            await loadSources();
            break;
        case 'keys':
            await loadKeys();
            break;
        case 'usage':
            await loadUsage();
            break;
    }
}

async function loadDashboard() {
    try {
        serverInfo = await apiRequest('/api/admin/server-info');
        document.getElementById('proxy-url').textContent = serverInfo.proxy_url;

        const data = await apiRequest('/api/admin/usage');
        usageStats = data;
        
        document.getElementById('stat-models').textContent = data.model_stats?.length || 0;
        document.getElementById('stat-sources').textContent = data.source_stats?.length || 0;
        document.getElementById('stat-keys').textContent = data.api_key_stats?.length || 0;
        
        const totalCalls = data.api_key_stats?.reduce((sum, k) => sum + k.used_count, 0) || 0;
        document.getElementById('stat-calls').textContent = totalCalls;
        
        const tbody = document.getElementById('model-stats-table');
        tbody.innerHTML = (data.model_stats || []).map(m => `
            <tr>
                <td>${m.name}</td>
                <td><code>${m.model}</code></td>
                <td>${m.strategy === 'round_robin' ? '轮询' : '用完切换'}</td>
                <td>${m.source_count}</td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load dashboard:', error);
    }
}

async function loadModels() {
    try {
        models = await apiRequest('/api/admin/models');
        const tbody = document.getElementById('models-table');
        tbody.innerHTML = models.map(m => `
            <tr>
                <td>
                    <div class="model-name-cell">
                        <span class="model-name">${m.name}</span>
                        <button class="btn btn-icon btn-copy" onclick="copyText('${escapeHtml(m.name).replace(/'/g, "\\'")}')" title="复制名称">
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="5" y="5" width="9" height="9" rx="1"></rect>
                                <path d="M2 11V3a1 1 0 0 1 1-1h8"></path>
                            </svg>
                        </button>
                    </div>
                </td>
                <td>
                    <div class="model-name-cell">
                        <code>${m.model}</code>
                        <button class="btn btn-icon btn-copy" onclick="copyText('${m.model}')" title="复制标识">
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="5" y="5" width="9" height="9" rx="1"></rect>
                                <path d="M2 11V3a1 1 0 0 1 1-1h8"></path>
                            </svg>
                        </button>
                    </div>
                </td>
                <td>${m.strategy === 'round_robin' ? '轮询' : '用完切换'}</td>
                <td>${m.sources?.length || 0}</td>
                <td>
                    <button class="btn btn-secondary btn-sm" onclick="editModel(${m.id})">编辑</button>
                    <button class="btn btn-danger btn-sm" onclick="deleteModel(${m.id})">删除</button>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load models:', error);
    }
}

async function loadSources() {
    try {
        sources = await apiRequest('/api/admin/sources');
        const tbody = document.getElementById('sources-table');
        tbody.innerHTML = sources.map(s => {
            const limit = s.billing_mode === 'count' ? s.limit_count : s.limit_tokens;
            const used = s.billing_mode === 'count' ? s.used_count : s.used_tokens;
            const limitText = limit === 0 ? '无限制' : `${used}/${limit}`;
            const status = !s.is_active ? '已禁用' : (limit === 0 || used < limit ? '正常' : '已用完');
            const badgeClass = !s.is_active ? 'badge-danger' : (limit === 0 || used < limit ? 'badge-success' : 'badge-warning');
            const lastReset = s.last_reset_at || '-';
            
            return `
            <tr>
                <td>
                    <span class="source-name" id="source-name-${s.id}">${s.name}</span>
                    <button class="btn btn-icon" onclick="renameSource(${s.id})" title="修改名称">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M11 4H4a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h7a2 2 0 0 0 2-2V9"></path>
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                        </svg>
                    </button>
                </td>
                <td><code>${s.api_url}</code></td>
                <td><code>${s.model_name}</code></td>
                <td>${s.billing_mode === 'count' ? '按次' : '按Token'}</td>
                <td>${limitText}</td>
                <td><small>${lastReset}</small></td>
                <td><span class="badge ${badgeClass}">${status}</span></td>
                <td>
                    <button class="btn btn-secondary btn-sm" onclick="editSource(${s.id})">编辑</button>
                    <button class="btn btn-danger btn-sm" onclick="deleteSource(${s.id})">删除</button>
                    <button class="btn btn-secondary btn-sm" onclick="resetSource(${s.id})">重置</button>
                </td>
            </tr>
        `}).join('');
    } catch (error) {
        console.error('Failed to load sources:', error);
    }
}

async function loadKeys() {
    try {
        apiKeys = await apiRequest('/api/admin/keys');
        const models = await apiRequest('/api/admin/models');
        const modelMap = {};
        models.forEach(m => modelMap[m.id] = m.name);
        
        const tbody = document.getElementById('keys-table');
        tbody.innerHTML = apiKeys.map(k => `
            <tr>
                <td>
                    <div class="api-key-cell">
                        <code class="api-key-text" data-key="${k.key}" data-shown="false">${maskKey(k.key)}</code>
                        <button class="btn btn-icon btn-copy" onclick="copyKey('${k.key}')" title="复制">
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="5" y="5" width="9" height="9" rx="1"></rect>
                                <path d="M2 11V3a1 1 0 0 1 1-1h8"></path>
                            </svg>
                        </button>
                        <button class="btn btn-icon btn-toggle" onclick="toggleKeyVisibility(this, '${k.key}')" title="显示/隐藏">
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" class="icon-eye">
                                <path d="M1 8s2.5-5 7-5 7 5 7 5-2.5 5-7 5-7-5-7-5z"></path>
                                <circle cx="8" cy="8" r="2"></circle>
                            </svg>
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" class="icon-eye-off hidden">
                                <path d="M1 1l14 14"></path>
                                <path d="M10.5 5.5c1.5-1 3.5-1 5 0"></path>
                                <path d="M1 8s2.5-5 7-5c1.5 0 2.5.5 3.5 1.5"></path>
                                <path d="M14 14L2 2"></path>
                                <path d="M8.5 10.5c-1.5 1-3.5 1-5 0"></path>
                                <path d="M14 3.5C12.5 2.5 10.5 2.5 9 3.5"></path>
                            </svg>
                        </button>
                    </div>
                </td>
                <td>${k.note || '-'}</td>
                <td>${k.external_model_id === 0 ? '全部模型' : (modelMap[k.external_model_id] || '-')}</td>
                <td>${k.used_count}</td>
                <td>${k.input_tokens}</td>
                <td>${k.output_tokens}</td>
                <td>
                    <button class="btn btn-danger btn-sm" onclick="deleteKey(${k.id})">删除</button>
                    <button class="btn btn-secondary btn-sm" onclick="resetKey(${k.id})">重置</button>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load keys:', error);
    }
}

async function loadUsage() {
    try {
        const data = await apiRequest('/api/admin/usage');
        
        const sourcesBody = document.getElementById('usage-sources-table');
        sourcesBody.innerHTML = (data.source_stats || []).map(s => {
            const limit = s.billing_mode === 'count' ? s.limit_count : s.limit_tokens;
            const used = s.billing_mode === 'count' ? s.used_count : s.used_tokens;
            const limitText = limit === 0 ? '无限制' : `${used}/${limit}`;
            const status = !s.is_active ? '已禁用' : (limit === 0 || used < limit ? '正常' : '已用完');
            const badgeClass = !s.is_active ? 'badge-danger' : (limit === 0 || used < limit ? 'badge-success' : 'badge-warning');
            const lastReset = s.last_reset_at || '-';
            
            return `
            <tr>
                <td>${s.name}</td>
                <td><code>${s.model_name}</code></td>
                <td>${s.billing_mode === 'count' ? '按次' : '按Token'}</td>
                <td>${s.used_count}</td>
                <td>${s.used_tokens}</td>
                <td>${limitText}</td>
                <td><small>${lastReset}</small></td>
                <td><span class="badge ${badgeClass}">${status}</span></td>
            </tr>
        `}).join('');
        
        const keysBody = document.getElementById('usage-keys-table');
        keysBody.innerHTML = (data.api_key_stats || []).map(k => `
            <tr>
                <td><code>${k.key}</code></td>
                <td>${k.note || '-'}</td>
                <td>${k.used_count}</td>
                <td>${k.input_tokens}</td>
                <td>${k.output_tokens}</td>
                <td>${k.used_tokens}</td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Failed to load usage:', error);
    }
}

document.getElementById('add-model-btn').addEventListener('click', () => {
    document.getElementById('model-id').value = '';
    document.getElementById('model-modal-title').textContent = '添加模型';
    document.getElementById('model-form').reset();
    openModal('model-modal');
});

async function editModel(id) {
    const model = models.find(m => m.id === id);
    if (!model) return;
    
    document.getElementById('model-id').value = model.id;
    document.getElementById('model-name').value = model.name;
    document.getElementById('model-model').value = model.model;
    document.getElementById('model-strategy').value = model.strategy;
    document.getElementById('model-modal-title').textContent = '编辑模型';
    openModal('model-modal');
}

async function deleteModel(id) {
    if (!confirm('确定要删除这个模型吗？')) return;
    try {
        await apiRequest(`/api/admin/models/${id}`, 'DELETE');
        await loadModels();
        showSuccess('模型删除成功');
    } catch (error) {
        showError(error.message);
    }
}

document.getElementById('model-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('model-id').value;
    const data = {
        name: document.getElementById('model-name').value,
        model: document.getElementById('model-model').value,
        strategy: document.getElementById('model-strategy').value
    };
    
    try {
        if (id) {
            await apiRequest(`/api/admin/models/${id}`, 'PUT', data);
            showSuccess('模型更新成功');
        } else {
            await apiRequest('/api/admin/models', 'POST', data);
            showSuccess('模型创建成功');
        }
        closeModal('model-modal');
        await loadModels();
    } catch (error) {
        showError(error.message);
    }
});

document.getElementById('add-source-btn').addEventListener('click', async () => {
    const models = await apiRequest('/api/admin/models');
    const select = document.getElementById('source-model-id');
    select.innerHTML = '<option value="">请选择模型</option>' + 
        models.map(m => `<option value="${m.id}">${m.name}</option>`).join('');
    
    document.getElementById('source-id').value = '';
    document.getElementById('source-modal-title').textContent = '添加请求源';
    document.getElementById('source-form').reset();
    openModal('source-modal');
});

async function editSource(id) {
    const source = sources.find(s => s.id === id);
    if (!source) return;
    
    const models = await apiRequest('/api/admin/models');
    const select = document.getElementById('source-model-id');
    select.innerHTML = '<option value="">请选择模型</option>' + 
        models.map(m => `<option value="${m.id}" ${m.id === source.external_model_id ? 'selected' : ''}>${m.name}</option>`).join('');
    
    document.getElementById('source-id').value = source.id;
    document.getElementById('source-name').value = source.name;
    document.getElementById('source-url').value = source.api_url;
    document.getElementById('source-key').value = source.api_key;
    document.getElementById('source-model-name').value = source.model_name;
    document.getElementById('source-billing').value = source.billing_mode;
document.getElementById('source-limit').value = source.billing_mode === 'count' ? source.limit_count : source.limit_tokens;
    document.getElementById('source-reset-interval').value = source.limit_reset_interval || '';
    document.getElementById('source-reset-time').value = source.limit_reset_time || '';
    document.getElementById('source-modal-title').textContent = '编辑请求源';
    openModal('source-modal');
}

async function deleteSource(id) {
    if (!confirm('确定要删除这个请求源吗？')) return;
    try {
        await apiRequest(`/api/admin/sources/${id}`, 'DELETE');
        await loadSources();
        showSuccess('请求源删除成功');
    } catch (error) {
        showError(error.message);
    }
}

async function resetSource(id) {
    if (!confirm('确定要重置这个请求源的用量吗？')) return;
    try {
        await apiRequest(`/api/admin/sources/${id}/reset`, 'POST');
        await loadSources();
        showSuccess('请求源用量已重置');
    } catch (error) {
        showError(error.message);
    }
}

document.getElementById('source-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('source-id').value;
    const billingMode = document.getElementById('source-billing').value;
    const limit = parseInt(document.getElementById('source-limit').value) || 0;
    const resetInterval = parseInt(document.getElementById('source-reset-interval').value) || 0;

const data = {
        external_model_id: parseInt(document.getElementById('source-model-id').value),
        name: document.getElementById('source-name').value,
        api_url: document.getElementById('source-url').value,
        api_key: document.getElementById('source-key').value,
        model_name: document.getElementById('source-model-name').value,
        billing_mode: billingMode,
        limit_count: billingMode === 'count' ? limit : 0,
        limit_tokens: billingMode === 'token' ? limit : 0,
        limit_reset_interval: resetInterval,
        limit_reset_time: document.getElementById('source-reset-time').value.trim()
    };

    try {
        if (id) {
            await apiRequest(`/api/admin/sources/${id}`, 'PUT', data);
            showSuccess('请求源更新成功');
        } else {
            await apiRequest('/api/admin/sources', 'POST', data);
            showSuccess('请求源创建成功');
        }
        closeModal('source-modal');
        await loadSources();
    } catch (error) {
        showError(error.message);
    }
});

async function renameSource(id) {
    const currentName = document.getElementById(`source-name-${id}`).textContent;
    document.getElementById('source-name-id').value = id;
    document.getElementById('source-name-input').value = currentName;
    openModal('source-name-modal');
}

document.getElementById('source-name-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('source-name-id').value;
    const name = document.getElementById('source-name-input').value;

    try {
        await apiRequest(`/api/admin/sources/${id}/name`, 'PATCH', { name });
        closeModal('source-name-modal');
        await loadSources();
        showSuccess('请求源名称修改成功');
    } catch (error) {
        showError(error.message);
    }
});

document.getElementById('add-key-btn').addEventListener('click', async () => {
    const models = await apiRequest('/api/admin/models');
    const select = document.getElementById('key-model-id');
    select.innerHTML = '<option value="">全部模型</option>' + 
        models.map(m => `<option value="${m.id}">${m.name}</option>`).join('');
    
    document.getElementById('key-form').reset();
    openModal('key-modal');
});

async function deleteKey(id) {
    if (!confirm('确定要删除这个 API Key 吗？')) return;
    try {
        await apiRequest(`/api/admin/keys/${id}`, 'DELETE');
        await loadKeys();
        showSuccess('API Key 删除成功');
    } catch (error) {
        showError(error.message);
    }
}

async function resetKey(id) {
    if (!confirm('确定要重置这个 API Key 的用量吗？')) return;
    try {
        await apiRequest(`/api/admin/keys/${id}/reset`, 'POST');
        await loadKeys();
        showSuccess('API Key 用量已重置');
    } catch (error) {
        showError(error.message);
    }
}

document.getElementById('key-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const modelId = document.getElementById('key-model-id').value;
    const data = {
        external_model_id: modelId ? parseInt(modelId) : null,
        note: document.getElementById('key-note').value
    };
    
    try {
        const result = await apiRequest('/api/admin/keys', 'POST', data);
        closeModal('key-modal');
        document.getElementById('new-key-display').textContent = result.key;
        openModal('key-display-modal');
        await loadKeys();
        showSuccess('API Key 生成成功');
    } catch (error) {
        showError(error.message);
    }
});

document.getElementById('logout-btn').addEventListener('click', () => {
    token = null;
    localStorage.removeItem('token');
    stopTokenRefresh();
    showLogin();
});

document.getElementById('import-models-btn').addEventListener('click', () => {
    document.getElementById('import-form').reset();
    document.getElementById('import-result').classList.add('hidden');
    openModal('import-modal');
});

document.getElementById('import-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const fileInput = document.getElementById('import-file');
    const file = fileInput.files[0];
    if (!file) {
        showError('请选择文件');
        return;
    }

    const submitBtn = e.target.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = '导入中...';

    try {
        const formData = new FormData();
        formData.append('file', file);

        const headers = {};
        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        const response = await fetch(`${API_BASE}/api/admin/models/import`, {
            method: 'POST',
            headers: headers,
            body: formData
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || '导入失败');
        }

        const result = data.result;
        let resultHtml = `
            <div class="result-title">导入结果</div>
            <div class="result-item"><span>新建</span><span class="value">${result.external_models_created}</span></div>
            <div class="result-item"><span>更新对外模型</span><span class="value">${result.external_models_updated}</span></div>
            <div class="result-item"><span>新建上游模型</span><span class="value">${result.sources_created}</span></div>
            <div class="result-item"><span>更新上游模型</span><span class="value">${result.sources_updated}</span></div>
        `;

        if (result.errors && result.errors.length > 0) {
            resultHtml += `<div class="errors"><div class="result-title" style="color: var(--danger);">错误</div>`;
            result.errors.forEach(err => {
                resultHtml += `<div class="error-item">${escapeHtml(err)}</div>`;
            });
            resultHtml += `</div>`;
        }

        if (result.skipped && result.skipped.length > 0) {
            resultHtml += `<div class="skipped"><div class="result-title" style="color: var(--warning);">跳过</div>`;
            result.skipped.forEach(s => {
                resultHtml += `<div class="skipped-item">${escapeHtml(s)}</div>`;
            });
            resultHtml += `</div>`;
        }

        document.getElementById('import-result').innerHTML = resultHtml;
        document.getElementById('import-result').classList.remove('hidden');

        if (result.errors && result.errors.length === 0) {
            showSuccess('导入成功');
            await loadModels();
            await loadSources();
            submitBtn.textContent = '导入成功';
        } else if (result.errors && result.errors.length > 0) {
            showError('导入完成，但有错误');
            submitBtn.disabled = false;
            submitBtn.textContent = '导入';
        }
    } catch (error) {
        showError(error.message);
        submitBtn.disabled = false;
        submitBtn.textContent = '导入';
    }
});

async function downloadTemplate() {
    try {
        const response = await fetch(`${API_BASE}/api/admin/models/template`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        
        const contentType = response.headers.get('content-type') || '';
        
        if (!response.ok) {
            if (contentType.includes('application/json')) {
                const data = await response.json();
                throw new Error(data.error || '下载失败');
            } else {
                throw new Error('下载失败: ' + response.status);
            }
        }

        if (!contentType.includes('application/vnd.openxmlformats-officedocument') && 
            !contentType.includes('application/octet-stream')) {
            const text = await response.text();
            console.error('Unexpected response:', text);
            throw new Error('服务器返回了非文件响应');
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'models_template.xlsx';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    } catch (error) {
        showError(error.message);
    }
}

function openModal(id) {
    document.getElementById(id).classList.add('active');
}

function closeModal(id) {
    document.getElementById(id).classList.remove('active');
}

async function copyProxyURL() {
    const proxyURL = document.getElementById('proxy-url').textContent;
    if (!proxyURL || proxyURL === '加载中...') {
        showError('代理地址加载中，请稍候');
        return;
    }
    try {
        await navigator.clipboard.writeText(proxyURL);
        showSuccess('代理地址已复制到剪贴板');
    } catch (error) {
        showError('复制失败，请手动复制');
    }
}

function maskKey(key) {
    if (key.length <= 8) {
        return '*'.repeat(key.length);
    }
    return key.substring(0, 4) + '*'.repeat(key.length - 8) + key.substring(key.length - 4);
}

function toggleKeyVisibility(btn, key) {
    const iconEye = btn.querySelector('.icon-eye');
    const iconEyeOff = btn.querySelector('.icon-eye-off');
    const row = btn.closest('tr');
    const keyText = row.querySelector('.api-key-text');
    
    if (keyText.dataset.shown === 'true') {
        keyText.textContent = maskKey(key);
        keyText.dataset.shown = 'false';
        iconEye.classList.remove('hidden');
        iconEyeOff.classList.add('hidden');
    } else {
        keyText.textContent = key;
        keyText.dataset.shown = 'true';
        iconEye.classList.add('hidden');
        iconEyeOff.classList.remove('hidden');
    }
}

async function copyKey(key) {
    try {
        await navigator.clipboard.writeText(key);
        showSuccess('API Key 已复制到剪贴板');
    } catch (error) {
        showError('复制失败，请手动复制');
    }
}

async function copyToClipboard(text) {
    try {
        await navigator.clipboard.writeText(text);
        showSuccess('已复制到剪贴板');
    } catch (error) {
        showError('复制失败，请手动复制');
    }
}

let newGeneratedKey = '';

function copyNewKey() {
    const keyText = document.getElementById('new-key-display').textContent;
    if (keyText) {
        copyToClipboard(keyText);
    }
}

function copyText(text) {
    copyToClipboard(text);
}

if (token) {
    showApp();
    navigateTo('dashboard');
} else {
    showLogin();
}
