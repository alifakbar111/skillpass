import { lazy, Suspense } from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { RootLayout } from './components/layout/RootLayout';
import { LoadingFallback } from './components/ui/LoadingFallback';
import { AuthProvider } from './hooks/useAuth';

// Eager — critical path (always needed)
import { Landing } from './pages/Landing';
import { Login } from './pages/Login';
import { Register } from './pages/Register';

// Lazy — loaded on navigation (named exports wrapped for lazy())
const JobseekerProfile = lazy(() => import('./pages/JobseekerProfile').then((m) => ({ default: m.JobseekerProfile })));
const JobseekerPassport = lazy(() =>
  import('./pages/JobseekerPassport').then((m) => ({ default: m.JobseekerPassport })),
);
const CompanyProfile = lazy(() => import('./pages/CompanyProfile').then((m) => ({ default: m.CompanyProfile })));
const CompanyVerification = lazy(() =>
  import('./pages/CompanyVerification').then((m) => ({ default: m.CompanyVerification })),
);
const CompanySearch = lazy(() => import('./pages/CompanySearch').then((m) => ({ default: m.CompanySearch })));
const CompanyJobs = lazy(() => import('./pages/CompanyJobs').then((m) => ({ default: m.CompanyJobs })));
const PublicJobs = lazy(() => import('./pages/PublicJobs').then((m) => ({ default: m.PublicJobs })));
const JobDetail = lazy(() => import('./pages/JobDetail').then((m) => ({ default: m.JobDetail })));
const PublicPassport = lazy(() => import('./pages/PublicPassport').then((m) => ({ default: m.PublicPassport })));
const AdminVerifications = lazy(() =>
  import('./pages/AdminVerifications').then((m) => ({ default: m.AdminVerifications })),
);

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: '/', element: <Landing /> },
      { path: '/auth/login', element: <Login /> },
      { path: '/auth/register', element: <Register /> },
      {
        path: '/jobseeker/profile',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <JobseekerProfile />
          </Suspense>
        ),
      },
      {
        path: '/jobseeker/passport',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <JobseekerPassport />
          </Suspense>
        ),
      },
      {
        path: '/company/profile',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <CompanyProfile />
          </Suspense>
        ),
      },
      {
        path: '/company/verification',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <CompanyVerification />
          </Suspense>
        ),
      },
      {
        path: '/company/search',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <CompanySearch />
          </Suspense>
        ),
      },
      {
        path: '/company/jobs',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <CompanyJobs />
          </Suspense>
        ),
      },
      {
        path: '/jobs',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <PublicJobs />
          </Suspense>
        ),
      },
      {
        path: '/jobs/:id',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <JobDetail />
          </Suspense>
        ),
      },
      {
        path: '/profiles/:username',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <PublicPassport />
          </Suspense>
        ),
      },
      {
        path: '/admin/verifications',
        element: (
          <Suspense fallback={<LoadingFallback />}>
            <AdminVerifications />
          </Suspense>
        ),
      },
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
