import React from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Federations from './pages/Federations'
import FederationDetail from './pages/FederationDetail'
import Collaborators from './pages/Collaborators'
import Rounds from './pages/Rounds'
import Events from './pages/Events'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/federations" element={<Federations />} />
        <Route path="/federations/:id" element={<FederationDetail />} />
        <Route path="/collaborators" element={<Collaborators />} />
        <Route path="/rounds" element={<Rounds />} />
        <Route path="/events" element={<Events />} />
      </Routes>
    </Layout>
  )
}

export default App
