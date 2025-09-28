import {Routes, Route} from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import Home from "./Pages/Home"
import { Login } from './Pages/Login';
import Signup from './Pages/Signup';
import Profile from './Pages/Profile';
import Pets from './Pages/Pets'
import Shop from './Pages/Shop';
import Game from './Pages/Game';

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
          <Route path='/profile' element={<Profile />}></Route>
          <Route path='/pets' element={<Pets />}></Route>
          <Route path='/shop' element={<Shop />}></Route>
          <Route path='/mingames' element={<Game />}></Route>
        </Routes>
      </QueryClientProvider>
    </>
  )
}

export default App
