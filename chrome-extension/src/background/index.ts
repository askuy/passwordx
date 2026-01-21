// Background service worker for PasswordX Chrome Extension

// Listen for messages from popup or content scripts
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'GET_CREDENTIALS') {
    // Handle credential retrieval
    handleGetCredentials(message.url).then(sendResponse)
    return true // Keep channel open for async response
  }

  if (message.type === 'SAVE_CREDENTIAL') {
    // Handle saving new credentials
    handleSaveCredential(message.credential).then(sendResponse)
    return true
  }
})

// Handle get credentials request
async function handleGetCredentials(url: string) {
  try {
    const storage = await chrome.storage.local.get(['token'])
    if (!storage.token) {
      return { success: false, error: 'Not authenticated' }
    }

    const response = await fetch(
      `http://localhost:8080/api/credentials/search?q=${encodeURIComponent(url)}`,
      {
        headers: {
          Authorization: `Bearer ${storage.token}`,
        },
      }
    )

    if (!response.ok) {
      return { success: false, error: 'Failed to fetch credentials' }
    }

    const data = await response.json()
    return { success: true, credentials: data.credentials }
  } catch (error) {
    return { success: false, error: 'Network error' }
  }
}

// Handle save credential request
async function handleSaveCredential(credential: {
  title: string
  url: string
  username: string
  password: string
  vaultId: number
}) {
  try {
    const storage = await chrome.storage.local.get(['token'])
    if (!storage.token) {
      return { success: false, error: 'Not authenticated' }
    }

    // Note: In production, encryption should happen in the popup where the master key is available
    const response = await fetch(
      `http://localhost:8080/api/vaults/${credential.vaultId}/credentials`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${storage.token}`,
        },
        body: JSON.stringify({
          title_encrypted: credential.title, // Would be encrypted
          url_encrypted: credential.url,
          username_encrypted: credential.username,
          password_encrypted: credential.password,
        }),
      }
    )

    if (!response.ok) {
      return { success: false, error: 'Failed to save credential' }
    }

    return { success: true }
  } catch (error) {
    return { success: false, error: 'Network error' }
  }
}

// Listen for installation
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === 'install') {
    console.log('PasswordX extension installed')
  } else if (details.reason === 'update') {
    console.log('PasswordX extension updated')
  }
})

// Context menu for quick actions
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: 'passwordx-generate',
    title: 'Generate Password',
    contexts: ['editable'],
  })

  chrome.contextMenus.create({
    id: 'passwordx-fill',
    title: 'Fill Login',
    contexts: ['editable'],
  })
})

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (!tab?.id) return

  if (info.menuItemId === 'passwordx-generate') {
    // Generate and fill password
    const password = generateRandomPassword(16)
    chrome.tabs.sendMessage(tab.id, {
      type: 'FILL_FIELD',
      value: password,
    })
  } else if (info.menuItemId === 'passwordx-fill') {
    // Open popup to select credential
    chrome.action.openPopup()
  }
})

function generateRandomPassword(length: number): string {
  const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*'
  const randomValues = new Uint8Array(length)
  crypto.getRandomValues(randomValues)
  let password = ''
  for (let i = 0; i < length; i++) {
    password += charset[randomValues[i] % charset.length]
  }
  return password
}
