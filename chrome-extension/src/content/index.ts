// Content script for PasswordX Chrome Extension
// Handles form detection and auto-fill functionality

interface LoginForm {
  form: HTMLFormElement | null
  usernameField: HTMLInputElement | null
  passwordField: HTMLInputElement | null
}

// Detect login forms on the page
function detectLoginForms(): LoginForm[] {
  const forms: LoginForm[] = []

  // Find all password fields
  const passwordFields = document.querySelectorAll<HTMLInputElement>(
    'input[type="password"]'
  )

  passwordFields.forEach((passwordField) => {
    // Find the associated form
    const form = passwordField.closest('form')

    // Find username/email field (usually before password field)
    let usernameField: HTMLInputElement | null = null

    // Look for common username field patterns
    const usernameSelectors = [
      'input[type="email"]',
      'input[type="text"][name*="user"]',
      'input[type="text"][name*="email"]',
      'input[type="text"][name*="login"]',
      'input[type="text"][autocomplete="username"]',
      'input[autocomplete="email"]',
    ]

    const container = form || passwordField.parentElement?.parentElement

    if (container) {
      for (const selector of usernameSelectors) {
        const field = container.querySelector<HTMLInputElement>(selector)
        if (field && field !== passwordField) {
          usernameField = field
          break
        }
      }

      // If no username field found, look for any text input before password
      if (!usernameField && form) {
        const inputs = form.querySelectorAll<HTMLInputElement>('input')
        for (let i = 0; i < inputs.length; i++) {
          if (inputs[i] === passwordField) break
          if (
            inputs[i].type === 'text' ||
            inputs[i].type === 'email' ||
            !inputs[i].type
          ) {
            usernameField = inputs[i]
          }
        }
      }
    }

    forms.push({
      form,
      usernameField,
      passwordField,
    })
  })

  return forms
}

// Fill credential into form
function fillCredential(username: string, password: string) {
  const forms = detectLoginForms()

  if (forms.length === 0) {
    console.log('PasswordX: No login forms detected')
    return
  }

  // Fill the first form (or the one in focus)
  const targetForm = forms[0]

  if (targetForm.usernameField && username) {
    setInputValue(targetForm.usernameField, username)
  }

  if (targetForm.passwordField && password) {
    setInputValue(targetForm.passwordField, password)
  }

  console.log('PasswordX: Credential filled')
}

// Set input value and trigger events
function setInputValue(input: HTMLInputElement, value: string) {
  // Focus the input
  input.focus()

  // Set the value
  input.value = value

  // Dispatch events to trigger any listeners
  input.dispatchEvent(new Event('input', { bubbles: true }))
  input.dispatchEvent(new Event('change', { bubbles: true }))
  
  // For React controlled inputs
  const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
    window.HTMLInputElement.prototype,
    'value'
  )?.set
  
  if (nativeInputValueSetter) {
    nativeInputValueSetter.call(input, value)
    input.dispatchEvent(new Event('input', { bubbles: true }))
  }
}

// Fill a single field (for context menu)
function fillField(value: string) {
  const activeElement = document.activeElement as HTMLInputElement
  if (activeElement && activeElement.tagName === 'INPUT') {
    setInputValue(activeElement, value)
  }
}

// Listen for messages from popup/background
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'FILL_CREDENTIAL') {
    fillCredential(message.username, message.password)
    sendResponse({ success: true })
  } else if (message.type === 'FILL_FIELD') {
    fillField(message.value)
    sendResponse({ success: true })
  } else if (message.type === 'DETECT_FORMS') {
    const forms = detectLoginForms()
    sendResponse({
      success: true,
      hasLoginForm: forms.length > 0,
      formCount: forms.length,
    })
  }
  return true
})

// Add PasswordX icon to password fields
function addPasswordXIcons() {
  const passwordFields = document.querySelectorAll<HTMLInputElement>(
    'input[type="password"]:not([data-passwordx-icon])'
  )

  passwordFields.forEach((field) => {
    field.setAttribute('data-passwordx-icon', 'true')

    // Create icon container
    const iconContainer = document.createElement('div')
    iconContainer.style.cssText = `
      position: absolute;
      right: 8px;
      top: 50%;
      transform: translateY(-50%);
      width: 20px;
      height: 20px;
      cursor: pointer;
      z-index: 10000;
      opacity: 0.7;
      transition: opacity 0.2s;
    `

    iconContainer.innerHTML = `
      <svg viewBox="0 0 24 24" fill="none" stroke="#6366f1" stroke-width="2">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
      </svg>
    `

    iconContainer.addEventListener('mouseenter', () => {
      iconContainer.style.opacity = '1'
    })

    iconContainer.addEventListener('mouseleave', () => {
      iconContainer.style.opacity = '0.7'
    })

    iconContainer.addEventListener('click', (e) => {
      e.preventDefault()
      e.stopPropagation()
      // Open extension popup
      chrome.runtime.sendMessage({ type: 'OPEN_POPUP' })
    })

    // Position the icon
    const parent = field.parentElement
    if (parent) {
      parent.style.position = 'relative'
      parent.appendChild(iconContainer)
    }
  })
}

// Observe DOM for dynamically added forms
const observer = new MutationObserver(() => {
  addPasswordXIcons()
})

// Start observing
observer.observe(document.body, {
  childList: true,
  subtree: true,
})

// Initial scan
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', addPasswordXIcons)
} else {
  addPasswordXIcons()
}

// Notify that content script is loaded
console.log('PasswordX content script loaded')
