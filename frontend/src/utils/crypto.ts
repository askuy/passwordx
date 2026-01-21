// Web Crypto API based encryption utilities

const PBKDF2_ITERATIONS = 100000
const SALT_LENGTH = 32
const KEY_LENGTH = 256

// Derive a key from password using PBKDF2
export async function deriveKey(password: string, saltBase64: string): Promise<CryptoKey> {
  const encoder = new TextEncoder()
  const passwordBuffer = encoder.encode(password)
  const salt = base64ToArrayBuffer(saltBase64)

  const keyMaterial = await window.crypto.subtle.importKey(
    'raw',
    passwordBuffer,
    'PBKDF2',
    false,
    ['deriveKey']
  )

  return window.crypto.subtle.deriveKey(
    {
      name: 'PBKDF2',
      salt: salt,
      iterations: PBKDF2_ITERATIONS,
      hash: 'SHA-256',
    },
    keyMaterial,
    { name: 'AES-GCM', length: KEY_LENGTH },
    false,
    ['encrypt', 'decrypt']
  )
}

// Encrypt data using AES-GCM
export async function encrypt(plaintext: string, key: CryptoKey): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(plaintext)

  // Generate random IV
  const iv = window.crypto.getRandomValues(new Uint8Array(12))

  const encrypted = await window.crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: iv },
    key,
    data
  )

  // Combine IV and encrypted data
  const combined = new Uint8Array(iv.length + encrypted.byteLength)
  combined.set(iv)
  combined.set(new Uint8Array(encrypted), iv.length)

  return arrayBufferToBase64(combined.buffer)
}

// Decrypt data using AES-GCM
export async function decrypt(ciphertextBase64: string, key: CryptoKey): Promise<string> {
  const combined = base64ToArrayBuffer(ciphertextBase64)
  const combinedArray = new Uint8Array(combined)

  // Extract IV and encrypted data
  const iv = combinedArray.slice(0, 12)
  const encrypted = combinedArray.slice(12)

  const decrypted = await window.crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: iv },
    key,
    encrypted
  )

  const decoder = new TextDecoder()
  return decoder.decode(decrypted)
}

// Generate a random salt
export function generateSalt(): string {
  const salt = window.crypto.getRandomValues(new Uint8Array(SALT_LENGTH))
  return arrayBufferToBase64(salt.buffer)
}

// Generate a random password
export function generatePassword(length: number = 16, options: {
  uppercase?: boolean
  lowercase?: boolean
  numbers?: boolean
  symbols?: boolean
} = {}): string {
  const {
    uppercase = true,
    lowercase = true,
    numbers = true,
    symbols = true,
  } = options

  let charset = ''
  if (uppercase) charset += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  if (lowercase) charset += 'abcdefghijklmnopqrstuvwxyz'
  if (numbers) charset += '0123456789'
  if (symbols) charset += '!@#$%^&*()_+-=[]{}|;:,.<>?'

  if (charset.length === 0) {
    charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  }

  const randomValues = window.crypto.getRandomValues(new Uint8Array(length))
  let password = ''
  for (let i = 0; i < length; i++) {
    password += charset[randomValues[i] % charset.length]
  }

  return password
}

// Calculate password strength (0-100)
export function calculatePasswordStrength(password: string): number {
  let score = 0

  // Length
  if (password.length >= 8) score += 20
  if (password.length >= 12) score += 10
  if (password.length >= 16) score += 10

  // Character variety
  if (/[a-z]/.test(password)) score += 15
  if (/[A-Z]/.test(password)) score += 15
  if (/[0-9]/.test(password)) score += 15
  if (/[^a-zA-Z0-9]/.test(password)) score += 15

  return Math.min(100, score)
}

// Helper functions
function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return btoa(binary)
}

function base64ToArrayBuffer(base64: string): ArrayBuffer {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

// Store for the master key (in memory only)
let masterKey: CryptoKey | null = null

export function setMasterKey(key: CryptoKey) {
  masterKey = key
}

export function getMasterKey(): CryptoKey | null {
  return masterKey
}

export function clearMasterKey() {
  masterKey = null
}
