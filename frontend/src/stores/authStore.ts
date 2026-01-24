import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: number
  email: string
  name: string
  avatar?: string
  tenant_id: number
  master_key_salt?: string
  role: string
  account_type: string
  status: string
}

interface Tenant {
  id: number
  name: string
  slug: string
}

interface AuthState {
  token: string | null
  user: User | null
  tenant: Tenant | null
  masterKey: string | null
  isAuthenticated: boolean
  setAuth: (token: string, user: User, tenant: Tenant) => void
  setMasterKey: (key: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      tenant: null,
      masterKey: null,
      isAuthenticated: false,

      setAuth: (token, user, tenant) =>
        set({
          token,
          user,
          tenant,
          isAuthenticated: true,
        }),

      setMasterKey: (key) =>
        set({
          masterKey: key,
        }),

      logout: () =>
        set({
          token: null,
          user: null,
          tenant: null,
          masterKey: null,
          isAuthenticated: false,
        }),
    }),
    {
      name: 'passwordx-auth',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        tenant: state.tenant,
        isAuthenticated: state.isAuthenticated,
        // Note: masterKey is NOT persisted for security
      }),
    }
  )
)
