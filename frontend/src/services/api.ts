import axios, { AxiosError } from 'axios';
import { message } from 'antd';
import type { 
  Model, 
  Source, 
  ApiKey, 
  UsageStats, 
  ServerInfo, 
  LoginResponse,
  ImportResult 
} from '../types';

const api = axios.create({
  baseURL: '',
  timeout: 30000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string }>) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/';
    }
    return Promise.reject(error);
  }
);

let refreshTimer: ReturnType<typeof setInterval> | null = null;
const TOKEN_REFRESH_INTERVAL = 10 * 60 * 1000;

export function startTokenRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
  refreshTimer = setInterval(async () => {
    try {
      await api.post('/api/admin/refresh-token');
    } catch {
      stopTokenRefresh();
      localStorage.removeItem('token');
      window.location.href = '/';
      message.error('登录已过期，请重新登录');
    }
  }, TOKEN_REFRESH_INTERVAL);
}

export function stopTokenRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

export const authApi = {
  login: async (username: string, password: string): Promise<LoginResponse> => {
    const response = await api.post<LoginResponse>('/api/login', { username, password });
    return response.data;
  },
  
  changePassword: async (oldPassword: string, newPassword: string, confirmPassword: string) => {
    await api.post('/api/admin/change-password', {
      old_password: oldPassword,
      new_password: newPassword,
      confirm_password: confirmPassword,
    });
  },
  
  changeUsername: async (newUsername: string, password: string) => {
    await api.post('/api/admin/change-username', {
      new_username: newUsername,
      password,
    });
  },
};

export const modelApi = {
  getAll: async (): Promise<Model[]> => {
    const response = await api.get<Model[]>('/api/admin/models');
    return response.data;
  },
  
  create: async (data: { name: string; model: string; strategy: string }): Promise<Model> => {
    const response = await api.post<Model>('/api/admin/models', data);
    return response.data;
  },
  
  update: async (id: number, data: { name: string; model: string; strategy: string }): Promise<void> => {
    await api.put(`/api/admin/models/${id}`, data);
  },
  
  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/admin/models/${id}`);
  },
  
  importExcel: async (file: File): Promise<ImportResult> => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post<{ result: ImportResult }>('/api/admin/models/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
    return response.data.result;
  },
  
  downloadTemplate: async (): Promise<Blob> => {
    const response = await api.get('/api/admin/models/template', {
      responseType: 'blob',
    });
    return response.data;
  },
};

export const sourceApi = {
  getAll: async (): Promise<Source[]> => {
    const response = await api.get<Source[]>('/api/admin/sources');
    return response.data;
  },
  
  create: async (data: Partial<Source>): Promise<Source> => {
    const response = await api.post<Source>('/api/admin/sources', data);
    return response.data;
  },
  
  update: async (id: number, data: Partial<Source>): Promise<void> => {
    await api.put(`/api/admin/sources/${id}`, data);
  },
  
  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/admin/sources/${id}`);
  },
  
  reset: async (id: number): Promise<void> => {
    await api.post(`/api/admin/sources/${id}/reset`);
  },
  
  updateName: async (id: number, name: string): Promise<void> => {
    await api.patch(`/api/admin/sources/${id}/name`, { name });
  },
};

export const keyApi = {
  getAll: async (): Promise<ApiKey[]> => {
    const response = await api.get<ApiKey[]>('/api/admin/keys');
    return response.data;
  },
  
  create: async (data: { external_model_id: number | null; note: string }): Promise<{ key: string }> => {
    const response = await api.post<{ key: string }>('/api/admin/keys', data);
    return response.data;
  },
  
  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/admin/keys/${id}`);
  },
  
  reset: async (id: number): Promise<void> => {
    await api.post(`/api/admin/keys/${id}/reset`);
  },
};

export const usageApi = {
  get: async (): Promise<UsageStats> => {
    const response = await api.get<UsageStats>('/api/admin/usage');
    return response.data;
  },
};

export const serverApi = {
  getInfo: async (): Promise<ServerInfo> => {
    const response = await api.get<ServerInfo>('/api/admin/server-info');
    return response.data;
  },
};

export default api;