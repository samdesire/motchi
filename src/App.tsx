import {Routes, Route} from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import Home from "./Pages/Home"
import { Login } from './Pages/Login';
import Signup from './Pages/Signup';

import './App.css'

function App() {
  const queryClient = new QueryClient();

  return (
    <>
      <QueryClientProvider client={queryClient}>
          <Routes>
          <Route path='/' element={<Home />}></Route>
          <Route path='/login' element={<Login />}></Route>
          <Route path='/sign-up' element={<Signup />}></Route>
        </Routes>
      </QueryClientProvider>
    </>
  )
}

export default App
