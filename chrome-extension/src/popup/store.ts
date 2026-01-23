import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { deriveKey, decrypt, setMasterKey } from '../utils/crypto'

export interface Credential {
  id: number
  title: string
  url?: string
  username?: string
  password: string
  favicon?: string
}

interface User {
  id: number
  email: string
  name: string
  tenant_id: number
  master_key_salt: string
}

interface AuthState {
  isAuthenticated: boolean
  isUnlocked: boolean
  token: string | null
  user: User | null
  credentials: Credential[]
  login: (email: string, password: string) => Promise<boolean>
  unlock: (password: string) => Promise<boolean>
  logout: () => void
  fetchCredentials: () => Promise<void>
}

const API_BASE = 'http://localhost:8080/api'

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      isAuthenticated: false,
      isUnlocked: false,
      token: null,
      user: null,
      credentials: [],

      login: async (email: string, password: string) => {
        try {
          const res = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
          })

          if (!res.ok) return false

          const data = await res.json()
          
          // Store auth but don't unlock yet - need master password
          set({
            isAuthenticated: true,
            token: data.token,
            user: data.user,
          })

          // Derive and set master key
          if (data.user.master_key_salt) {
            const key = await deriveKey(password, data.user.master_key_salt)
            setMasterKey(key)
            set({ isUnlocked: true })
          }

          return true
        } catch {
          return false
        }
      },

      unlock: async (password: string) => {
        const { user } = get()
        if (!user?.master_key_salt) return false

        try {
          const key = await deriveKey(password, user.master_key_salt)
          setMasterKey(key)
          set({ isUnlocked: true })
          return true
        } catch {
          return false
        }
      },

      logout: () => {
        set({
          isAuthenticated: false,
          isUnlocked: false,
          token: null,
          user: null,
          credentials: [],
        })
      },

      fetchCredentials: async () => {
        const { token, isUnlocked } = get()
        if (!token || !isUnlocked) return

        try {
          // Fetch all credentials from user's vaults (no URL filter - filtering done client-side after decryption)
          const res = await fetch(`${API_BASE}/credentials/search`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })

          if (!res.ok) return

          const data = await res.json()
          const encryptedCreds = data.credentials || []

          // Decrypt credentials
          const decryptedCreds: Credential[] = []
          
          for (const cred of encryptedCreds) {
            try {
              const decrypted: Credential = {
                id: cred.id,
                title: await decrypt(cred.title_encrypted),
                url: cred.url_encrypted ? await decrypt(cred.url_encrypted) : undefined,
                username: cred.username_encrypted ? await decrypt(cred.username_encrypted) : undefined,
                password: await decrypt(cred.password_encrypted),
                favicon: cred.favicon,
              }
              decryptedCreds.push(decrypted)
            } catch {
              // Skip credentials that fail to decrypt
            }
          }

          set({ credentials: decryptedCreds })
        } catch {
          // Handle error silently
        }
      },
    }),
    {
      name: 'passwordx-extension-auth',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        token: state.token,
        user: state.user,
      }),
    }
  )
)
