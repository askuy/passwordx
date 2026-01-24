import { useState } from 'react'
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  LayoutDashboard,
  FolderLock,
  Settings,
  LogOut,
  Plus,
  Shield,
  Menu,
  X,
  Users,
} from 'lucide-react'
import { useAuthStore } from '../stores/authStore'
import { vaultAPI } from '../services/api'
import { clearMasterKey } from '../utils/crypto'
import CreateVaultModal from './CreateVaultModal'

interface VaultItem {
  id: number
  name: string
  icon?: string
}

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, tenant, logout } = useAuthStore()
  const [showCreateVault, setShowCreateVault] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(false)

  const { data: vaultsData } = useQuery({
    queryKey: ['vaults'],
    queryFn: async () => {
      const res = await vaultAPI.list()
      return res.data.vaults as VaultItem[]
    },
  })

  const handleLogout = () => {
    clearMasterKey()
    logout()
    navigate('/login')
  }

  // Check if user is admin
  const isAdmin = user?.role === 'super_admin' || user?.role === 'admin'

  const navItems = [
    { icon: LayoutDashboard, label: 'Dashboard', path: '/dashboard' },
    ...(isAdmin ? [{ icon: Users, label: 'Members', path: '/members' }] : []),
    { icon: Settings, label: 'Settings', path: '/settings' },
  ]

  return (
    <div className="min-h-screen flex">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`
          fixed lg:static inset-y-0 left-0 z-50
          w-72 glass flex flex-col
          transform transition-transform duration-300 ease-in-out
          ${sidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        `}
      >
        {/* Logo */}
        <div className="p-6 flex items-center gap-3 border-b border-dark-700/50">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center glow-sm">
            <Shield className="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 className="font-bold text-lg text-white">PasswordX</h1>
            <p className="text-xs text-dark-400">{tenant?.name}</p>
          </div>
          <button
            onClick={() => setSidebarOpen(false)}
            className="ml-auto lg:hidden p-2 hover:bg-dark-800 rounded-lg"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
          {/* Main nav items */}
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              onClick={() => setSidebarOpen(false)}
              className={`
                flex items-center gap-3 px-4 py-3 rounded-xl transition-all
                ${location.pathname === item.path
                  ? 'bg-primary-500/20 text-primary-400'
                  : 'text-dark-300 hover:bg-dark-800 hover:text-white'
                }
              `}
            >
              <item.icon className="w-5 h-5" />
              <span className="font-medium">{item.label}</span>
            </Link>
          ))}

          {/* Vaults section */}
          <div className="pt-6">
            <div className="flex items-center justify-between px-4 mb-2">
              <span className="text-xs font-semibold text-dark-500 uppercase tracking-wider">
                Vaults
              </span>
              <button
                onClick={() => setShowCreateVault(true)}
                className="p-1 hover:bg-dark-800 rounded-lg transition-colors"
                title="Create vault"
              >
                <Plus className="w-4 h-4 text-dark-400" />
              </button>
            </div>

            <div className="space-y-1 stagger">
              {vaultsData?.map((vault) => (
                <Link
                  key={vault.id}
                  to={`/vault/${vault.id}`}
                  onClick={() => setSidebarOpen(false)}
                  className={`
                    flex items-center gap-3 px-4 py-3 rounded-xl transition-all
                    ${location.pathname === `/vault/${vault.id}`
                      ? 'bg-primary-500/20 text-primary-400'
                      : 'text-dark-300 hover:bg-dark-800 hover:text-white'
                    }
                  `}
                >
                  <FolderLock className="w-5 h-5" />
                  <span className="font-medium truncate">{vault.name}</span>
                </Link>
              ))}
            </div>
          </div>
        </nav>

        {/* User section */}
        <div className="p-4 border-t border-dark-700/50">
          <div className="flex items-center gap-3 px-4 py-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white font-semibold">
              {user?.name?.charAt(0).toUpperCase() || 'U'}
            </div>
            <div className="flex-1 min-w-0">
              <p className="font-medium text-white truncate">{user?.name}</p>
              <p className="text-xs text-dark-400 truncate">{user?.email}</p>
            </div>
            <button
              onClick={handleLogout}
              className="p-2 hover:bg-dark-800 rounded-lg transition-colors text-dark-400 hover:text-red-400"
              title="Logout"
            >
              <LogOut className="w-5 h-5" />
            </button>
          </div>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 flex flex-col min-h-screen">
        {/* Mobile header */}
        <header className="lg:hidden p-4 glass border-b border-dark-700/50">
          <button
            onClick={() => setSidebarOpen(true)}
            className="p-2 hover:bg-dark-800 rounded-lg"
          >
            <Menu className="w-6 h-6" />
          </button>
        </header>

        {/* Page content */}
        <div className="flex-1 p-6 lg:p-8 overflow-auto">
          <Outlet />
        </div>
      </main>

      {/* Create vault modal */}
      {showCreateVault && (
        <CreateVaultModal onClose={() => setShowCreateVault(false)} />
      )}
    </div>
  )
}
