import { BrowserRouter } from 'react-router-dom';
import { UserProvider } from './hooks/UserContext';
import { AppRouter } from './router/AppRouter'

function App() {
  return (
    <BrowserRouter>
        <UserProvider >
          <AppRouter />
        </UserProvider>
    </BrowserRouter>
  )
}

export default App
