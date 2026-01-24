import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Shield, Loader2 } from 'lucide-react'
import { useAuthStore } from '../stores/authStore'
import api from '../services/api'

export default function AuthCallbackPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setAuth } = useAuthStore()

  useEffect(() => {
    const token = searchParams.get('token')
    const error = searchParams.get('error')

    if (error) {
      // Handle OAuth errors
      if (error === 'not_invited') {
        alert('You need to be invited by an administrator to access this application.')
      } else if (error === 'inactive') {
        alert('Your account is inactive. Please contact an administrator.')
      }
      navigate('/login')
      return
    }

    if (token) {
      // Set token first so the API can use it
      // Then fetch full user data from /me endpoint
      const fetchUserData = async () => {
        try {
          // Set authorization header for this request
          const res = await api.get('/me', {
            headers: { Authorization: `Bearer ${token}` },
          })
          const { user, tenant } = res.data
          setAuth(token, user, tenant || { id: user.tenant_id, name: 'Workspace', slug: 'workspace' })
          navigate('/dashboard')
        } catch {
          // Fallback: decode JWT to get basic user info
          try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            const user = {
              id: payload.user_id,
              email: payload.email,
              name: payload.email.split('@')[0],
              tenant_id: payload.tenant_id,
              role: 'user',
              account_type: 'team',
              status: 'active',
            }
            const tenant = {
              id: payload.tenant_id,
              name: 'Workspace',
              slug: 'workspace',
            }
            setAuth(token, user, tenant)
            navigate('/dashboard')
          } catch {
            navigate('/login')
          }
        }
      }
      fetchUserData()
    } else {
      navigate('/login')
    }
  }, [searchParams, setAuth, navigate])

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center animate-fade-in">
        <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center mx-auto mb-4 glow">
          <Shield className="w-8 h-8 text-white" />
        </div>
        <div className="flex items-center justify-center gap-2 text-dark-300">
          <Loader2 className="w-5 h-5 animate-spin" />
          <span>Completing sign in...</span>
        </div>
      </div>
    </div>
  )
}
