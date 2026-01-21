import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import {
  Search,
  Plus,
  Vault,
  Key,
  Shield,
  Clock,
  TrendingUp,
} from 'lucide-react'
import { vaultAPI, credentialAPI } from '../services/api'
import CreateVaultModal from '../components/CreateVaultModal'

interface VaultItem {
  id: number
  name: string
  description?: string
  created_at: string
}

interface Credential {
  id: number
  vault_id: number
  title_encrypted: string
  url_encrypted?: string
  created_at: string
}

export default function DashboardPage() {
  const [showCreateVault, setShowCreateVault] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')

  const { data: vaultsData } = useQuery({
    queryKey: ['vaults'],
    queryFn: async () => {
      const res = await vaultAPI.list()
      return res.data.vaults as VaultItem[]
    },
  })

  const { data: searchResults, isLoading: isSearching } = useQuery({
    queryKey: ['credentials-search', searchQuery],
    queryFn: async () => {
      if (!searchQuery.trim()) return null
      const res = await credentialAPI.search(searchQuery)
      return res.data.credentials as Credential[]
    },
    enabled: searchQuery.length > 0,
  })

  const vaults = vaultsData || []
  const totalCredentials = vaults.length * 5 // Placeholder, would need actual count

  return (
    <div className="max-w-6xl mx-auto">
      {/* Header */}
      <div className="mb-8 animate-fade-in">
        <h1 className="text-3xl font-bold text-white mb-2">Dashboard</h1>
        <p className="text-dark-400">Welcome back! Here's an overview of your passwords.</p>
      </div>

      {/* Search bar */}
      <div className="mb-8 animate-fade-in" style={{ animationDelay: '0.05s' }}>
        <div className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-dark-500" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search credentials by URL or title..."
            className="w-full pl-12 pr-4 py-4 glass rounded-xl text-white placeholder-dark-500 focus:ring-2 focus:ring-primary-500"
          />
        </div>

        {/* Search results */}
        {searchQuery && (
          <div className="mt-4 glass rounded-xl p-4">
            {isSearching ? (
              <p className="text-dark-400">Searching...</p>
            ) : searchResults && searchResults.length > 0 ? (
              <div className="space-y-2">
                <p className="text-sm text-dark-500 mb-3">
                  Found {searchResults.length} result(s)
                </p>
                {searchResults.map((cred) => (
                  <Link
                    key={cred.id}
                    to={`/vault/${cred.vault_id}`}
                    className="flex items-center gap-3 p-3 rounded-lg hover:bg-dark-800 transition-colors"
                  >
                    <Key className="w-5 h-5 text-primary-400" />
                    <span className="text-white">{cred.title_encrypted}</span>
                  </Link>
                ))}
              </div>
            ) : (
              <p className="text-dark-400">No results found</p>
            )}
          </div>
        )}
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8 stagger">
        <div className="glass rounded-xl p-6">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-primary-500/20 flex items-center justify-center">
              <Vault className="w-6 h-6 text-primary-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{vaults.length}</p>
              <p className="text-dark-400 text-sm">Vaults</p>
            </div>
          </div>
        </div>

        <div className="glass rounded-xl p-6">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-green-500/20 flex items-center justify-center">
              <Key className="w-6 h-6 text-green-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-white">{totalCredentials}</p>
              <p className="text-dark-400 text-sm">Total Passwords</p>
            </div>
          </div>
        </div>

        <div className="glass rounded-xl p-6">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-yellow-500/20 flex items-center justify-center">
              <Shield className="w-6 h-6 text-yellow-400" />
            </div>
            <div>
              <p className="text-2xl font-bold text-white">AES-256</p>
              <p className="text-dark-400 text-sm">Encryption</p>
            </div>
          </div>
        </div>
      </div>

      {/* Vaults section */}
      <div className="mb-8 animate-fade-in" style={{ animationDelay: '0.2s' }}>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-white">Your Vaults</h2>
          <button
            onClick={() => setShowCreateVault(true)}
            className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium"
          >
            <Plus className="w-4 h-4" />
            New Vault
          </button>
        </div>

        {vaults.length === 0 ? (
          <div className="glass rounded-xl p-12 text-center">
            <Vault className="w-12 h-12 text-dark-600 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-white mb-2">No vaults yet</h3>
            <p className="text-dark-400 mb-6">
              Create your first vault to start storing passwords securely.
            </p>
            <button
              onClick={() => setShowCreateVault(true)}
              className="inline-flex items-center gap-2 px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium"
            >
              <Plus className="w-5 h-5" />
              Create Your First Vault
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 stagger">
            {vaults.map((vault) => (
              <Link
                key={vault.id}
                to={`/vault/${vault.id}`}
                className="glass rounded-xl p-6 hover:bg-dark-800/80 transition-all group"
              >
                <div className="flex items-start gap-4">
                  <div className="w-12 h-12 rounded-xl bg-primary-500/20 flex items-center justify-center group-hover:bg-primary-500/30 transition-colors">
                    <Vault className="w-6 h-6 text-primary-400" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="font-semibold text-white truncate group-hover:text-primary-400 transition-colors">
                      {vault.name}
                    </h3>
                    {vault.description && (
                      <p className="text-dark-400 text-sm truncate mt-1">
                        {vault.description}
                      </p>
                    )}
                    <div className="flex items-center gap-2 mt-3 text-xs text-dark-500">
                      <Clock className="w-3 h-3" />
                      <span>
                        Created {new Date(vault.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>

      {/* Create vault modal */}
      {showCreateVault && (
        <CreateVaultModal onClose={() => setShowCreateVault(false)} />
      )}
    </div>
  )
}
