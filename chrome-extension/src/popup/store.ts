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
          console.log('PasswordX: Attempting login for', email)
          const res = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password }),
          })

          if (!res.ok) {
            console.error('PasswordX: Login failed, status:', res.status)
            return false
          }

          const data = await res.json()
          console.log('PasswordX: Login successful, user:', data.user?.email)
          
          // Store auth but don't unlock yet - need master password
          set({
            isAuthenticated: true,
            token: data.token,
            user: data.user,
          })

          // Derive and set master key
          if (data.user.master_key_salt) {
            console.log('PasswordX: Deriving master key...')
            const key = await deriveKey(password, data.user.master_key_salt)
            setMasterKey(key)
            set({ isUnlocked: true })
            console.log('PasswordX: Master key derived and vault unlocked')
          } else {
            console.warn('PasswordX: No master_key_salt in user data')
          }

          return true
        } catch (err) {
          console.error('PasswordX: Login error', err)
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
        if (!token || !isUnlocked) {
          console.log('PasswordX: Cannot fetch credentials - not authenticated or not unlocked')
          return
        }

        try {
          // Fetch all credentials from user's vaults (no URL filter - filtering done client-side after decryption)
          const res = await fetch(`${API_BASE}/credentials/search`, {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          })

          if (!res.ok) {
            console.error('PasswordX: Failed to fetch credentials, status:', res.status)
            return
          }

          const data = await res.json()
          const encryptedCreds = data.credentials || []
          console.log('PasswordX: Fetched', encryptedCreds.length, 'encrypted credentials')

          // Decrypt credentials
          const decryptedCreds: Credential[] = []
          let decryptErrors = 0
          
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
            } catch (err) {
              decryptErrors++
              console.error('PasswordX: Failed to decrypt credential', cred.id, err)
            }
          }

          console.log('PasswordX: Decrypted', decryptedCreds.length, 'credentials,', decryptErrors, 'failed')
          set({ credentials: decryptedCreds })
        } catch (err) {
          console.error('PasswordX: Error fetching credentials', err)
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
