import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { RootLayout } from './components/layout/RootLayout';
import { AuthProvider } from './hooks/useAuth';
import { Landing } from './pages/Landing';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { JobseekerProfile } from './pages/JobseekerProfile';
import { JobseekerPassport } from './pages/JobseekerPassport';
import { CompanyProfile } from './pages/CompanyProfile';
import { CompanyVerification } from './pages/CompanyVerification';
import { CompanySearch } from './pages/CompanySearch';
import { CompanyJobs } from './pages/CompanyJobs';
import { PublicJobs } from './pages/PublicJobs';
import { JobDetail } from './pages/JobDetail';
import { PublicPassport } from './pages/PublicPassport';
import { AdminVerifications } from './pages/AdminVerifications';

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: '/', element: <Landing /> },
      { path: '/auth/login', element: <Login /> },
      { path: '/auth/register', element: <Register /> },
      { path: '/jobseeker/profile', element: <JobseekerProfile /> },
      { path: '/jobseeker/passport', element: <JobseekerPassport /> },
      { path: '/company/profile', element: <CompanyProfile /> },
      { path: '/company/verification', element: <CompanyVerification /> },
      { path: '/company/search', element: <CompanySearch /> },
      { path: '/company/jobs', element: <CompanyJobs /> },
      { path: '/jobs', element: <PublicJobs /> },
      { path: '/jobs/:id', element: <JobDetail /> },
      { path: '/profiles/:username', element: <PublicPassport /> },
      { path: '/admin/verifications', element: <AdminVerifications /> },
    ],
  },
]);

export function App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  );
}
