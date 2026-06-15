import { BrowserRouter } from 'react-router-dom'
import { Toaster } from 'sonner'

import { UserProvider } from './hooks/UserContext'
import { AppRouter } from './router/AppRouter'

function App() {
  return (
    <BrowserRouter>
      <UserProvider>
        <AppRouter />
      </UserProvider>
      {/* 通知系トースト。alert() を置き換える (ADR-0009 #3) */}
      <Toaster richColors position="top-center" />
    </BrowserRouter>
  )
}

export default App
