import { Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout.jsx'
import Inicio from './pages/Inicio.jsx'
import Login from './pages/Login.jsx'
import Postulacion from './pages/Postulacion.jsx'
import Recuperar from './pages/Recuperar.jsx'
import Registro from './pages/Registro.jsx'
import Restablecer from './pages/Restablecer.jsx'
import RutaProtegida from './routes/RutaProtegida.jsx'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Inicio />} />
        <Route path="login" element={<Login />} />
        <Route path="registro" element={<Registro />} />
        <Route path="recuperar" element={<Recuperar />} />
        <Route path="restablecer" element={<Restablecer />} />
        <Route element={<RutaProtegida />}>
          <Route path="postulacion" element={<Postulacion />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
