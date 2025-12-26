import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { AppProvider } from './context/AppContext'
import { installWailsPolyfill, installGoAppPolyfill, setRestApiClient } from './services/wailsPolyfill'
import { initializeServices, isRestApiEnabled, createRestApiClient } from './services/adapter'

// Install Wails polyfills BEFORE any other code runs
// This ensures window.runtime and window.go.main.App exist even when running outside of Wails
installWailsPolyfill()
installGoAppPolyfill()

// Initialize services (REST API or Wails) before rendering
initializeServices().then(() => {
  // If REST API is enabled, wire up the REST API client to the Go App polyfill
  if (isRestApiEnabled()) {
    const client = createRestApiClient()
    setRestApiClient(client)
  }

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <AppProvider>
        <App />
      </AppProvider>
    </StrictMode>,
  )
}).catch((error) => {
  console.error('Failed to initialize services:', error)
  // Render anyway - the app should handle missing services gracefully
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <AppProvider>
        <App />
      </AppProvider>
    </StrictMode>,
  )
})
