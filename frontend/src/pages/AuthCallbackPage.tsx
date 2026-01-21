import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Shield, Loader2 } from 'lucide-react'
import { useAuthStore } from '../stores/authStore'

export default function AuthCallbackPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setAuth } = useAuthStore()

  useEffect(() => {
    const token = searchParams.get('token')
    if (token) {
      // Decode JWT to get user info (simple base64 decode for payload)
      try {
        const payload = JSON.parse(atob(token.split('.')[1]))
        // For OAuth users, we might not have full user data in token
        // In production, you'd make an API call to get full user details
        const user = {
          id: payload.user_id,
          email: payload.email,
          name: payload.email.split('@')[0],
          tenant_id: payload.tenant_id,
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
