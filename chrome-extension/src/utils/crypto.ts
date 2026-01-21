const PBKDF2_ITERATIONS = 100000
const KEY_LENGTH = 256

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

export async function deriveKey(password: string, saltBase64: string): Promise<CryptoKey> {
  const encoder = new TextEncoder()
  const passwordBuffer = encoder.encode(password)
  const salt = base64ToArrayBuffer(saltBase64)

  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    passwordBuffer,
    'PBKDF2',
    false,
    ['deriveKey']
  )

  return crypto.subtle.deriveKey(
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

export async function encrypt(plaintext: string, key?: CryptoKey): Promise<string> {
  const cryptoKey = key || masterKey
  if (!cryptoKey) throw new Error('No encryption key available')

  const encoder = new TextEncoder()
  const data = encoder.encode(plaintext)
  const iv = crypto.getRandomValues(new Uint8Array(12))

  const encrypted = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv: iv },
    cryptoKey,
    data
  )

  const combined = new Uint8Array(iv.length + encrypted.byteLength)
  combined.set(iv)
  combined.set(new Uint8Array(encrypted), iv.length)

  return arrayBufferToBase64(combined.buffer)
}

export async function decrypt(ciphertextBase64: string, key?: CryptoKey): Promise<string> {
  const cryptoKey = key || masterKey
  if (!cryptoKey) throw new Error('No decryption key available')

  const combined = base64ToArrayBuffer(ciphertextBase64)
  const combinedArray = new Uint8Array(combined)

  const iv = combinedArray.slice(0, 12)
  const encrypted = combinedArray.slice(12)

  const decrypted = await crypto.subtle.decrypt(
    { name: 'AES-GCM', iv: iv },
    cryptoKey,
    encrypted
  )

  const decoder = new TextDecoder()
  return decoder.decode(decrypted)
}

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
