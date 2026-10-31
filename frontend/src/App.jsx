import { Navigate, Route, Routes } from 'react-router-dom'
import Layout from './components/Layout.jsx'
import Inicio from './pages/Inicio.jsx'
import Postulacion from './pages/Postulacion.jsx'
import Pendiente from './pages/Pendiente.jsx'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Inicio />} />
        {/* Fase 2 (#9) reemplaza estas pantallas */}
        <Route path="login" element={<Pendiente titulo="Iniciar sesión" />} />
        <Route path="registro" element={<Pendiente titulo="Registrarse" />} />
        <Route path="recuperar" element={<Pendiente titulo="Recuperar contraseña" />} />
        <Route path="postulacion" element={<Postulacion />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
