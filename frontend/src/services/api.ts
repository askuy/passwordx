import axios from 'axios'
import { useAuthStore } from '../stores/authStore'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const { token, tenant } = useAuthStore.getState()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (tenant) {
    config.headers['X-Tenant-ID'] = tenant.id.toString()
  }
  return config
})

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Settings API
export const settingsAPI = {
  getPublic: () => api.get<{ disable_registration: boolean }>('/settings'),
}

// Auth API
export const authAPI = {
  register: (data: {
    email: string
    password: string
    name: string
    tenant_name: string
    tenant_slug: string
  }) => api.post('/auth/register', data),

  login: (data: { email: string; password: string }) =>
    api.post('/auth/login', data),

  getOAuthURL: (provider: string) => `/api/auth/oauth/${provider}`,
}

// Vault API
export const vaultAPI = {
  list: () => api.get('/vaults'),
  get: (id: number) => api.get(`/vaults/${id}`),
  create: (data: { name: string; description?: string; icon?: string }) =>
    api.post('/vaults', data),
  update: (id: number, data: { name?: string; description?: string; icon?: string }) =>
    api.put(`/vaults/${id}`, data),
  delete: (id: number) => api.delete(`/vaults/${id}`),
  addMember: (id: number, data: { user_id: number; role: string }) =>
    api.post(`/vaults/${id}/members`, data),
  removeMember: (id: number, userId: number) =>
    api.delete(`/vaults/${id}/members/${userId}`),
}

// Credential API
export const credentialAPI = {
  list: (vaultId: number) => api.get(`/vaults/${vaultId}/credentials`),
  get: (vaultId: number, credId: number) =>
    api.get(`/vaults/${vaultId}/credentials/${credId}`),
  create: (vaultId: number, data: {
    title_encrypted: string
    url_encrypted?: string
    username_encrypted?: string
    password_encrypted: string
    notes_encrypted?: string
    category?: string
    favicon?: string
  }) => api.post(`/vaults/${vaultId}/credentials`, data),
  update: (vaultId: number, credId: number, data: {
    title_encrypted?: string
    url_encrypted?: string
    username_encrypted?: string
    password_encrypted?: string
    notes_encrypted?: string
    category?: string
    favicon?: string
  }) => api.put(`/vaults/${vaultId}/credentials/${credId}`, data),
  delete: (vaultId: number, credId: number) =>
    api.delete(`/vaults/${vaultId}/credentials/${credId}`),
  search: (query: string) => api.get(`/credentials/search?q=${encodeURIComponent(query)}`),
}

// Tenant API
export const tenantAPI = {
  list: () => api.get('/tenants'),
  get: (id: number) => api.get(`/tenants/${id}`),
  create: (data: { name: string; slug: string }) => api.post('/tenants', data),
  update: (id: number, data: { name?: string; slug?: string }) =>
    api.put(`/tenants/${id}`, data),
  delete: (id: number) => api.delete(`/tenants/${id}`),
}

// User Management API (Admin only)
export const userAPI = {
  list: (tenantId?: number) =>
    api.get('/admin/users', { params: tenantId ? { tenant_id: tenantId } : {} }),
  get: (id: number) => api.get(`/admin/users/${id}`),
  create: (data: {
    email: string
    name: string
    password?: string
    account_type: string
    tenant_id?: number
    role?: string
  }) => api.post('/admin/users', data),
  update: (id: number, data: { name?: string; role?: string; status?: string }) =>
    api.put(`/admin/users/${id}`, data),
  delete: (id: number) => api.delete(`/admin/users/${id}`),
  resetPassword: (id: number, password: string) =>
    api.post(`/admin/users/${id}/reset-password`, { password }),
}

// Current user API
export const meAPI = {
  get: () => api.get('/me'),
}

export default api
