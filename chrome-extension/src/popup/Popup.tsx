import { useState, useEffect } from 'react'
import {
  Shield,
  Search,
  Key,
  Globe,
  Copy,
  ExternalLink,
  LogIn,
  Loader2,
  Eye,
  EyeOff,
  Settings,
  RefreshCw,
} from 'lucide-react'
import { useAuthStore, type Credential } from './store'
import { deriveKey, decrypt, getMasterKey, setMasterKey } from '../utils/crypto'

type View = 'login' | 'unlock' | 'main' | 'generator'

export default function Popup() {
  const { isAuthenticated, isUnlocked, credentials, user, login, unlock, fetchCredentials, logout } = useAuthStore()
  const [view, setView] = useState<View>('login')
  const [currentUrl, setCurrentUrl] = useState('')

  useEffect(() => {
    // Get current tab URL
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0]?.url) {
        setCurrentUrl(tabs[0].url)
      }
    })

    // Determine view based on auth state
    if (isAuthenticated && isUnlocked) {
      setView('main')
      fetchCredentials(currentUrl)
    } else if (isAuthenticated && !isUnlocked) {
      setView('unlock')
    } else {
      setView('login')
    }
  }, [isAuthenticated, isUnlocked, currentUrl])

  return (
    <div className="min-h-[400px] flex flex-col">
      {/* Header */}
      <header className="p-4 border-b border-dark-800 flex items-center gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center">
          <Shield className="w-4 h-4 text-white" />
        </div>
        <h1 className="font-semibold text-white">PasswordX</h1>
        {isAuthenticated && (
          <div className="ml-auto flex items-center gap-2">
            <button
              onClick={() => setView('generator')}
              className="p-2 hover:bg-dark-800 rounded-lg"
              title="Password Generator"
            >
              <RefreshCw className="w-4 h-4 text-dark-400" />
            </button>
            <button
              onClick={() => {
                // Open web app
                chrome.tabs.create({ url: 'http://localhost:5173' })
              }}
              className="p-2 hover:bg-dark-800 rounded-lg"
              title="Open Web App"
            >
              <ExternalLink className="w-4 h-4 text-dark-400" />
            </button>
          </div>
        )}
      </header>

      {/* Content */}
      <main className="flex-1 overflow-auto">
        {view === 'login' && <LoginView onLogin={login} onSuccess={() => setView('unlock')} />}
        {view === 'unlock' && <UnlockView onUnlock={unlock} onSuccess={() => setView('main')} />}
        {view === 'main' && <MainView credentials={credentials} currentUrl={currentUrl} />}
        {view === 'generator' && <GeneratorView onBack={() => setView('main')} />}
      </main>

      {/* Footer */}
      {isAuthenticated && (
        <footer className="p-3 border-t border-dark-800 text-xs text-dark-500 flex items-center justify-between">
          <span>{user?.email}</span>
          <button onClick={logout} className="text-dark-400 hover:text-red-400">
            Sign out
          </button>
        </footer>
      )}
    </div>
  )
}

interface LoginViewProps {
  onLogin: (email: string, password: string) => Promise<boolean>
  onSuccess: () => void
}

function LoginView({ onLogin, onSuccess }: LoginViewProps) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    const success = await onLogin(email, password)
    if (success) {
      onSuccess()
    } else {
      setError('Invalid credentials')
    }
    setLoading(false)
  }

  return (
    <div className="p-4">
      <div className="text-center mb-6">
        <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center mx-auto mb-3">
          <Shield className="w-8 h-8 text-white" />
        </div>
        <h2 className="text-lg font-semibold text-white">Sign In</h2>
        <p className="text-dark-400 text-sm mt-1">Access your passwords</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Email"
            className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-lg text-white placeholder-dark-500 text-sm"
            required
          />
        </div>
        <div>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Password"
            className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-lg text-white placeholder-dark-500 text-sm"
            required
          />
        </div>
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <button
          type="submit"
          disabled={loading}
          className="w-full py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-500 transition-colors font-medium disabled:opacity-50 flex items-center justify-center gap-2"
        >
          {loading ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              Signing in...
            </>
          ) : (
            <>
              <LogIn className="w-4 h-4" />
              Sign In
            </>
          )}
        </button>
      </form>
    </div>
  )
}

interface UnlockViewProps {
  onUnlock: (password: string) => Promise<boolean>
  onSuccess: () => void
}

function UnlockView({ onUnlock, onSuccess }: UnlockViewProps) {
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    const success = await onUnlock(password)
    if (success) {
      onSuccess()
    } else {
      setError('Invalid master password')
    }
    setLoading(false)
  }

  return (
    <div className="p-4">
      <div className="text-center mb-6">
        <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center mx-auto mb-3">
          <Key className="w-8 h-8 text-white" />
        </div>
        <h2 className="text-lg font-semibold text-white">Unlock Vault</h2>
        <p className="text-dark-400 text-sm mt-1">Enter your master password</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Master Password"
            className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-lg text-white placeholder-dark-500 text-sm"
            autoFocus
            required
          />
        </div>
        {error && <p className="text-red-400 text-sm">{error}</p>}
        <button
          type="submit"
          disabled={loading}
          className="w-full py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-500 transition-colors font-medium disabled:opacity-50 flex items-center justify-center gap-2"
        >
          {loading ? (
            <>
              <Loader2 className="w-4 h-4 animate-spin" />
              Unlocking...
            </>
          ) : (
            'Unlock'
          )}
        </button>
      </form>
    </div>
  )
}

interface MainViewProps {
  credentials: Credential[]
  currentUrl: string
}

function MainView({ credentials, currentUrl }: MainViewProps) {
  const [search, setSearch] = useState('')
  const [visiblePasswords, setVisiblePasswords] = useState<Set<number>>(new Set())

  const filteredCredentials = credentials.filter((cred) => {
    if (!search) return true
    const lowerSearch = search.toLowerCase()
    return (
      cred.title.toLowerCase().includes(lowerSearch) ||
      cred.url?.toLowerCase().includes(lowerSearch) ||
      cred.username?.toLowerCase().includes(lowerSearch)
    )
  })

  // Sort by matching current URL first
  const sortedCredentials = [...filteredCredentials].sort((a, b) => {
    const aMatches = a.url && currentUrl.includes(new URL(a.url).hostname)
    const bMatches = b.url && currentUrl.includes(new URL(b.url).hostname)
    if (aMatches && !bMatches) return -1
    if (!aMatches && bMatches) return 1
    return 0
  })

  const togglePassword = (id: number) => {
    setVisiblePasswords((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const copyToClipboard = async (text: string) => {
    await navigator.clipboard.writeText(text)
  }

  const fillCredential = (cred: Credential) => {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0]?.id) {
        chrome.tabs.sendMessage(tabs[0].id, {
          type: 'FILL_CREDENTIAL',
          username: cred.username,
          password: cred.password,
        })
      }
    })
  }

  return (
    <div className="flex flex-col h-full">
      {/* Search */}
      <div className="p-3 border-b border-dark-800">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-dark-500" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search passwords..."
            className="w-full pl-9 pr-3 py-2 bg-dark-800 border border-dark-700 rounded-lg text-white placeholder-dark-500 text-sm"
          />
        </div>
      </div>

      {/* Credentials list */}
      <div className="flex-1 overflow-auto">
        {sortedCredentials.length === 0 ? (
          <div className="p-6 text-center">
            <Key className="w-10 h-10 text-dark-600 mx-auto mb-3" />
            <p className="text-dark-400">No passwords found</p>
          </div>
        ) : (
          <div className="divide-y divide-dark-800">
            {sortedCredentials.map((cred) => {
              const isMatch = cred.url && currentUrl.includes(new URL(cred.url).hostname)
              return (
                <div
                  key={cred.id}
                  className={`p-3 hover:bg-dark-800/50 ${isMatch ? 'bg-primary-500/10' : ''}`}
                >
                  <div className="flex items-start gap-3">
                    <div className="w-8 h-8 rounded-lg bg-dark-700 flex items-center justify-center flex-shrink-0">
                      {cred.favicon ? (
                        <img src={cred.favicon} alt="" className="w-4 h-4" />
                      ) : (
                        <Globe className="w-4 h-4 text-dark-400" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-white text-sm truncate">{cred.title}</p>
                      {cred.username && (
                        <p className="text-dark-400 text-xs truncate">{cred.username}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => togglePassword(cred.id)}
                        className="p-1.5 hover:bg-dark-700 rounded"
                        title={visiblePasswords.has(cred.id) ? 'Hide' : 'Show'}
                      >
                        {visiblePasswords.has(cred.id) ? (
                          <EyeOff className="w-3.5 h-3.5 text-dark-400" />
                        ) : (
                          <Eye className="w-3.5 h-3.5 text-dark-400" />
                        )}
                      </button>
                      <button
                        onClick={() => copyToClipboard(cred.password)}
                        className="p-1.5 hover:bg-dark-700 rounded"
                        title="Copy password"
                      >
                        <Copy className="w-3.5 h-3.5 text-dark-400" />
                      </button>
                      <button
                        onClick={() => fillCredential(cred)}
                        className="px-2 py-1 bg-primary-600 text-white text-xs rounded hover:bg-primary-500"
                        title="Fill"
                      >
                        Fill
                      </button>
                    </div>
                  </div>
                  {visiblePasswords.has(cred.id) && (
                    <div className="mt-2 ml-11">
                      <code className="text-xs text-dark-300 font-mono">{cred.password}</code>
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

interface GeneratorViewProps {
  onBack: () => void
}

function GeneratorView({ onBack }: GeneratorViewProps) {
  const [password, setPassword] = useState('')
  const [length, setLength] = useState(16)
  const [options, setOptions] = useState({
    uppercase: true,
    lowercase: true,
    numbers: true,
    symbols: true,
  })

  const generate = () => {
    let charset = ''
    if (options.uppercase) charset += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
    if (options.lowercase) charset += 'abcdefghijklmnopqrstuvwxyz'
    if (options.numbers) charset += '0123456789'
    if (options.symbols) charset += '!@#$%^&*()_+-='

    if (!charset) charset = 'abcdefghijklmnopqrstuvwxyz'

    const array = new Uint8Array(length)
    crypto.getRandomValues(array)
    let result = ''
    for (let i = 0; i < length; i++) {
      result += charset[array[i] % charset.length]
    }
    setPassword(result)
  }

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(password)
  }

  return (
    <div className="p-4">
      <button
        onClick={onBack}
        className="text-dark-400 hover:text-white text-sm mb-4"
      >
        ← Back
      </button>

      <h2 className="text-lg font-semibold text-white mb-4">Password Generator</h2>

      <div className="space-y-4">
        <div>
          <label className="text-sm text-dark-300 mb-2 block">Length: {length}</label>
          <input
            type="range"
            min="8"
            max="32"
            value={length}
            onChange={(e) => setLength(Number(e.target.value))}
            className="w-full"
          />
        </div>

        <div className="grid grid-cols-2 gap-2">
          {Object.entries(options).map(([key, value]) => (
            <label key={key} className="flex items-center gap-2 text-sm text-dark-300">
              <input
                type="checkbox"
                checked={value}
                onChange={(e) => setOptions((p) => ({ ...p, [key]: e.target.checked }))}
                className="rounded border-dark-600 bg-dark-800 text-primary-500"
              />
              <span className="capitalize">{key}</span>
            </label>
          ))}
        </div>

        <button
          onClick={generate}
          className="w-full py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-500 text-sm"
        >
          Generate
        </button>

        {password && (
          <div className="p-3 bg-dark-800 rounded-lg">
            <code className="text-white font-mono text-sm break-all">{password}</code>
            <button
              onClick={copyToClipboard}
              className="mt-2 w-full py-2 bg-dark-700 text-dark-300 rounded text-sm hover:bg-dark-600"
            >
              Copy
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
