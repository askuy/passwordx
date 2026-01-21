import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { User, Building, Shield, Key, Loader2, Check } from 'lucide-react'
import { useAuthStore } from '../stores/authStore'
import { tenantAPI } from '../services/api'
import { generatePassword, calculatePasswordStrength } from '../utils/crypto'

export default function SettingsPage() {
  const { user, tenant } = useAuthStore()
  const [activeTab, setActiveTab] = useState<'profile' | 'organization' | 'security'>('profile')

  const tabs = [
    { id: 'profile', label: 'Profile', icon: User },
    { id: 'organization', label: 'Organization', icon: Building },
    { id: 'security', label: 'Security', icon: Shield },
  ] as const

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8 animate-fade-in">
        <h1 className="text-3xl font-bold text-white mb-2">Settings</h1>
        <p className="text-dark-400">Manage your account and organization settings.</p>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 mb-8 animate-fade-in" style={{ animationDelay: '0.05s' }}>
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-xl font-medium transition-all ${
              activeTab === tab.id
                ? 'bg-primary-600 text-white'
                : 'text-dark-400 hover:text-white hover:bg-dark-800'
            }`}
          >
            <tab.icon className="w-4 h-4" />
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="glass rounded-2xl p-6 animate-fade-in" style={{ animationDelay: '0.1s' }}>
        {activeTab === 'profile' && <ProfileSettings user={user} />}
        {activeTab === 'organization' && <OrganizationSettings tenant={tenant} />}
        {activeTab === 'security' && <SecuritySettings />}
      </div>
    </div>
  )
}

function ProfileSettings({ user }: { user: any }) {
  const [name, setName] = useState(user?.name || '')
  const [email] = useState(user?.email || '')

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-white mb-4">Profile Information</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Full Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-white"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Email
            </label>
            <input
              type="email"
              value={email}
              disabled
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-dark-400 cursor-not-allowed"
            />
            <p className="text-xs text-dark-500 mt-1">Email cannot be changed</p>
          </div>

          <button className="px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium">
            Save Changes
          </button>
        </div>
      </div>
    </div>
  )
}

function OrganizationSettings({ tenant }: { tenant: any }) {
  const [name, setName] = useState(tenant?.name || '')
  const [saved, setSaved] = useState(false)

  const updateMutation = useMutation({
    mutationFn: () => tenantAPI.update(tenant.id, { name }),
    onSuccess: () => {
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    },
  })

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-white mb-4">Organization Settings</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Organization Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-white"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Organization Slug
            </label>
            <input
              type="text"
              value={tenant?.slug || ''}
              disabled
              className="w-full px-4 py-3 bg-dark-800 border border-dark-700 rounded-xl text-dark-400 cursor-not-allowed"
            />
            <p className="text-xs text-dark-500 mt-1">Slug cannot be changed</p>
          </div>

          <button
            onClick={() => updateMutation.mutate()}
            disabled={updateMutation.isPending}
            className="px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium flex items-center gap-2"
          >
            {updateMutation.isPending ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Saving...
              </>
            ) : saved ? (
              <>
                <Check className="w-4 h-4" />
                Saved!
              </>
            ) : (
              'Save Changes'
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

function SecuritySettings() {
  const [generatedPassword, setGeneratedPassword] = useState('')
  const [passwordLength, setPasswordLength] = useState(16)
  const [options, setOptions] = useState({
    uppercase: true,
    lowercase: true,
    numbers: true,
    symbols: true,
  })

  const handleGenerate = () => {
    const password = generatePassword(passwordLength, options)
    setGeneratedPassword(password)
  }

  const strength = calculatePasswordStrength(generatedPassword)

  const getStrengthColor = () => {
    if (strength < 40) return 'bg-red-500'
    if (strength < 70) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  return (
    <div className="space-y-8">
      {/* Password Generator */}
      <div>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Key className="w-5 h-5" />
          Password Generator
        </h3>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">
              Length: {passwordLength}
            </label>
            <input
              type="range"
              min="8"
              max="32"
              value={passwordLength}
              onChange={(e) => setPasswordLength(Number(e.target.value))}
              className="w-full"
            />
          </div>

          <div className="flex flex-wrap gap-4">
            {Object.entries(options).map(([key, value]) => (
              <label key={key} className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={value}
                  onChange={(e) =>
                    setOptions((prev) => ({ ...prev, [key]: e.target.checked }))
                  }
                  className="w-4 h-4 rounded border-dark-600 bg-dark-800 text-primary-500 focus:ring-primary-500"
                />
                <span className="text-dark-300 capitalize">{key}</span>
              </label>
            ))}
          </div>

          <button
            onClick={handleGenerate}
            className="px-6 py-3 bg-primary-600 text-white rounded-xl hover:bg-primary-500 transition-colors font-medium"
          >
            Generate Password
          </button>

          {generatedPassword && (
            <div className="mt-4">
              <div className="flex items-center gap-3 p-4 bg-dark-800 rounded-xl">
                <code className="flex-1 font-mono text-white break-all">
                  {generatedPassword}
                </code>
                <button
                  onClick={() => navigator.clipboard.writeText(generatedPassword)}
                  className="px-3 py-2 bg-dark-700 text-dark-300 rounded-lg hover:bg-dark-600 transition-colors text-sm"
                >
                  Copy
                </button>
              </div>
              <div className="mt-2">
                <div className="h-1 bg-dark-700 rounded-full overflow-hidden">
                  <div
                    className={`h-full transition-all ${getStrengthColor()}`}
                    style={{ width: `${strength}%` }}
                  />
                </div>
                <p className="text-xs text-dark-500 mt-1">
                  Strength: {strength < 40 ? 'Weak' : strength < 70 ? 'Medium' : 'Strong'}
                </p>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Security Info */}
      <div>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <Shield className="w-5 h-5" />
          Encryption
        </h3>
        <div className="p-4 bg-dark-800 rounded-xl">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-green-500/20 flex items-center justify-center">
              <Check className="w-4 h-4 text-green-400" />
            </div>
            <span className="font-medium text-white">AES-256-GCM Encryption</span>
          </div>
          <p className="text-dark-400 text-sm ml-11">
            All your passwords are encrypted using military-grade AES-256-GCM encryption.
            Your master password is never stored on our servers.
          </p>
        </div>
      </div>
    </div>
  )
}
