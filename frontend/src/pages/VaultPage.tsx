import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Key,
  Globe,
  User,
  Eye,
  EyeOff,
  Copy,
  Trash2,
  Edit,
  Loader2,
  RefreshCw,
} from 'lucide-react'
import { vaultAPI, credentialAPI } from '../services/api'
import { encrypt, decrypt, getMasterKey, generatePassword, calculatePasswordStrength } from '../utils/crypto'

interface Credential {
  id: number
  vault_id: number
  title_encrypted: string
  url_encrypted?: string
  username_encrypted?: string
  password_encrypted: string
  notes_encrypted?: string
  category?: string
  created_at: string
  updated_at: string
}

interface DecryptedCredential extends Credential {
  title: string
  url?: string
  username?: string
  password: string
  notes?: string
}

export default function VaultPage() {
  const { vaultId } = useParams<{ vaultId: string }>()
  const queryClient = useQueryClient()
  const [showAddModal, setShowAddModal] = useState(false)
  const [editingCred, setEditingCred] = useState<DecryptedCredential | null>(null)
  const [visiblePasswords, setVisiblePasswords] = useState<Set<number>>(new Set())

  const { data: vault } = useQuery({
    queryKey: ['vault', vaultId],
    queryFn: async () => {
      const res = await vaultAPI.get(Number(vaultId))
      return res.data
    },
  })

  const { data: credentials, isLoading } = useQuery({
    queryKey: ['credentials', vaultId],
    queryFn: async () => {
      const res = await credentialAPI.list(Number(vaultId))
      const creds = res.data.credentials as Credential[]

      // Decrypt credentials
      const key = getMasterKey()
      if (!key) return creds.map((c) => ({ ...c, title: c.title_encrypted, password: '••••••••' }))

      const decrypted = await Promise.all(
        creds.map(async (cred): Promise<DecryptedCredential> => {
          try {
            return {
              ...cred,
              title: await decrypt(cred.title_encrypted, key),
              url: cred.url_encrypted ? await decrypt(cred.url_encrypted, key) : undefined,
              username: cred.username_encrypted ? await decrypt(cred.username_encrypted, key) : undefined,
              password: await decrypt(cred.password_encrypted, key),
              notes: cred.notes_encrypted ? await decrypt(cred.notes_encrypted, key) : undefined,
            }
          } catch {
            return {
              ...cred,
              title: '[Decryption failed]',
              password: '[Decryption failed]',
            }
          }
        })
      )

      return decrypted
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (credId: number) => credentialAPI.delete(Number(vaultId), credId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials', vaultId] })
    },
  })

  const togglePasswordVisibility = (id: number) => {
    setVisiblePasswords((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const copyToClipboard = async (text: string) => {
    await navigator.clipboard.writeText(text)
  }

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8 animate-fade-in">
        <h1 className="text-3xl font-bold text-white mb-2">{vault?.name || 'Vault'}</h1>
        <p className="text-dark-400">{vault?.description || 'Manage your credentials'}</p>
      </div>

      {/* Actions */}
      <div className="flex justify-end mb-6 animate-fade-in" style={{ animationDelay: '0.05s' }}>
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium"
        >
          <Plus className="w-4 h-4" />
          Add Credential
        </button>
      </div>

      {/* Credentials list */}
      {isLoading ? (
        <div className="glass rounded-xl p-12 text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto text-primary-400" />
          <p className="text-dark-400 mt-4">Loading credentials...</p>
        </div>
      ) : credentials && credentials.length > 0 ? (
        <div className="space-y-4 stagger">
          {credentials.map((cred) => (
            <div key={cred.id} className="glass rounded-xl p-6">
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-10 h-10 rounded-lg bg-primary-500/20 flex items-center justify-center">
                      <Key className="w-5 h-5 text-primary-400" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-white">{cred.title}</h3>
                      {cred.url && (
                        <a
                          href={cred.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm text-primary-400 hover:underline flex items-center gap-1"
                        >
                          <Globe className="w-3 h-3" />
                          {new URL(cred.url).hostname}
                        </a>
                      )}
                    </div>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {cred.username && (
                      <div>
                        <p className="text-xs text-dark-500 mb-1">Username</p>
                        <div className="flex items-center gap-2">
                          <User className="w-4 h-4 text-dark-500" />
                          <span className="text-dark-300">{cred.username}</span>
                          <button
                            onClick={() => copyToClipboard(cred.username!)}
                            className="p-1 hover:bg-dark-700 rounded"
                            title="Copy username"
                          >
                            <Copy className="w-4 h-4 text-dark-500" />
                          </button>
                        </div>
                      </div>
                    )}

                    <div>
                      <p className="text-xs text-dark-500 mb-1">Password</p>
                      <div className="flex items-center gap-2">
                        <Key className="w-4 h-4 text-dark-500" />
                        <span className="text-dark-300 font-mono">
                          {visiblePasswords.has(cred.id) ? cred.password : '••••••••'}
                        </span>
                        <button
                          onClick={() => togglePasswordVisibility(cred.id)}
                          className="p-1 hover:bg-dark-700 rounded"
                          title={visiblePasswords.has(cred.id) ? 'Hide password' : 'Show password'}
                        >
                          {visiblePasswords.has(cred.id) ? (
                            <EyeOff className="w-4 h-4 text-dark-500" />
                          ) : (
                            <Eye className="w-4 h-4 text-dark-500" />
                          )}
                        </button>
                        <button
                          onClick={() => copyToClipboard(cred.password)}
                          className="p-1 hover:bg-dark-700 rounded"
                          title="Copy password"
                        >
                          <Copy className="w-4 h-4 text-dark-500" />
                        </button>
                      </div>
                    </div>
                  </div>

                  {cred.notes && (
                    <div className="mt-4">
                      <p className="text-xs text-dark-500 mb-1">Notes</p>
                      <p className="text-dark-400 text-sm">{cred.notes}</p>
                    </div>
                  )}
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setEditingCred(cred as DecryptedCredential)}
                    className="p-2 hover:bg-dark-700 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <Edit className="w-4 h-4 text-dark-400" />
                  </button>
                  <button
                    onClick={() => {
                      if (confirm('Are you sure you want to delete this credential?')) {
                        deleteMutation.mutate(cred.id)
                      }
                    }}
                    className="p-2 hover:bg-dark-700 rounded-lg transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="w-4 h-4 text-red-400" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="glass rounded-xl p-12 text-center">
          <Key className="w-12 h-12 text-dark-600 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-white mb-2">No credentials yet</h3>
          <p className="text-dark-400 mb-6">
            Add your first credential to this vault.
          </p>
          <button
            onClick={() => setShowAddModal(true)}
            className="inline-flex items-center gap-2 px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium"
          >
            <Plus className="w-5 h-5" />
            Add First Credential
          </button>
        </div>
      )}

      {/* Add/Edit Modal */}
      {(showAddModal || editingCred) && (
        <CredentialModal
          vaultId={Number(vaultId)}
          credential={editingCred}
          onClose={() => {
            setShowAddModal(false)
            setEditingCred(null)
          }}
        />
      )}
    </div>
  )
}

interface CredentialModalProps {
  vaultId: number
  credential?: DecryptedCredential | null
  onClose: () => void
}

function CredentialModal({ vaultId, credential, onClose }: CredentialModalProps) {
  const queryClient = useQueryClient()
  const [formData, setFormData] = useState({
    title: credential?.title || '',
    url: credential?.url || '',
    username: credential?.username || '',
    password: credential?.password || '',
    notes: credential?.notes || '',
  })

  const passwordStrength = calculatePasswordStrength(formData.password)

  const createMutation = useMutation({
    mutationFn: async () => {
      const key = getMasterKey()
      if (!key) throw new Error('Master key not available')

      const encrypted = {
        title_encrypted: await encrypt(formData.title, key),
        url_encrypted: formData.url ? await encrypt(formData.url, key) : undefined,
        username_encrypted: formData.username ? await encrypt(formData.username, key) : undefined,
        password_encrypted: await encrypt(formData.password, key),
        notes_encrypted: formData.notes ? await encrypt(formData.notes, key) : undefined,
      }

      if (credential) {
        return credentialAPI.update(vaultId, credential.id, encrypted)
      }
      return credentialAPI.create(vaultId, encrypted)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials', String(vaultId)] })
      onClose()
    },
  })

  const handleGeneratePassword = () => {
    const newPassword = generatePassword(20)
    setFormData((prev) => ({ ...prev, password: newPassword }))
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createMutation.mutate()
  }

  const getStrengthColor = () => {
    if (passwordStrength < 40) return 'bg-red-500'
    if (passwordStrength < 70) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <div className="glass rounded-2xl w-full max-w-md animate-fade-in glow max-h-[90vh] overflow-y-auto">
        <div className="p-6 border-b border-dark-700/50">
          <h2 className="text-xl font-semibold text-white">
            {credential ? 'Edit Credential' : 'Add Credential'}
          </h2>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Title *
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData((prev) => ({ ...prev, title: e.target.value }))}
              placeholder="Google Account"
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-white placeholder-dark-500"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              URL
            </label>
            <input
              type="url"
              value={formData.url}
              onChange={(e) => setFormData((prev) => ({ ...prev, url: e.target.value }))}
              placeholder="https://google.com"
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-white placeholder-dark-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Username
            </label>
            <input
              type="text"
              value={formData.username}
              onChange={(e) => setFormData((prev) => ({ ...prev, username: e.target.value }))}
              placeholder="user@example.com"
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-white placeholder-dark-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Password *
            </label>
            <div className="relative">
              <input
                type="text"
                value={formData.password}
                onChange={(e) => setFormData((prev) => ({ ...prev, password: e.target.value }))}
                placeholder="••••••••"
                className="w-full px-4 py-3 pr-12 bg-dark-800 border border-dark-700 rounded-xl text-white font-mono placeholder-dark-500"
                required
              />
              <button
                type="button"
                onClick={handleGeneratePassword}
                className="absolute right-3 top-1/2 -translate-y-1/2 p-1 hover:bg-dark-700 rounded"
                title="Generate password"
              >
                <RefreshCw className="w-4 h-4 text-dark-400" />
              </button>
            </div>
            {formData.password && (
              <div className="mt-2">
                <div className="h-1 bg-dark-700 rounded-full overflow-hidden">
                  <div
                    className={`h-full transition-all ${getStrengthColor()}`}
                    style={{ width: `${passwordStrength}%` }}
                  />
                </div>
              </div>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Notes
            </label>
            <textarea
              value={formData.notes}
              onChange={(e) => setFormData((prev) => ({ ...prev, notes: e.target.value }))}
              placeholder="Additional notes..."
              rows={3}
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-white placeholder-dark-500 resize-none"
            />
          </div>

          {createMutation.error && (
            <p className="text-red-400 text-sm">
              Failed to save credential. Please try again.
            </p>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-3 bg-dark-800 text-dark-300 rounded-xl hover:bg-dark-700 transition-colors font-medium"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={!formData.title || !formData.password || createMutation.isPending}
              className="flex-1 px-4 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium disabled:opacity-50 flex items-center justify-center gap-2"
            >
              {createMutation.isPending ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Saving...
                </>
              ) : credential ? (
                'Update'
              ) : (
                'Add'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
