import { Routes, Route } from 'react-router-dom';
import Layout from '@/components/layout/Layout';
import Dashboard from '@/pages/Dashboard';
import RepoList from '@/pages/RepoList';
import RepoDetail from '@/pages/RepoDetail';
import Settings from '@/pages/Settings';

/**
 * Root application component.
 * Defines the top-level route structure for LocalRepo.
 */
function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/repos" element={<RepoList />} />
        <Route path="/repos/:name" element={<RepoDetail />} />
        <Route path="/settings" element={<Settings />} />
      </Routes>
    </Layout>
  );
}

export default App;
