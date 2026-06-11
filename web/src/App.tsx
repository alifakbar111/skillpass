import { QueryClientProvider } from '@tanstack/react-query';
import { lazy, Suspense } from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { RootLayout } from './components/layout/RootLayout';
import { ErrorBoundary } from './components/ui/ErrorBoundary';
import { LoadingFallback } from './components/ui/LoadingFallback';
import { ProtectedRoute } from './components/ui/ProtectedRoute';
import { AuthProvider } from './hooks/useAuth';
import { queryClient } from './lib/queryClient';

import { Landing } from './pages/Landing';
import { Login } from './pages/Login';
import { Register } from './pages/Register';

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
const CompanyApplications = lazy(() =>
  import('./pages/CompanyApplications').then((m) => ({ default: m.CompanyApplications })),
);
const CompanyAnalytics = lazy(() => import('./pages/CompanyAnalytics').then((m) => ({ default: m.CompanyAnalytics })));
const PublicJobs = lazy(() => import('./pages/PublicJobs').then((m) => ({ default: m.PublicJobs })));
const JobDetail = lazy(() => import('./pages/JobDetail').then((m) => ({ default: m.JobDetail })));
const PublicPassport = lazy(() => import('./pages/PublicPassport').then((m) => ({ default: m.PublicPassport })));
const AdminVerifications = lazy(() =>
  import('./pages/AdminVerifications').then((m) => ({ default: m.AdminVerifications })),
);
const EvaluationPage = lazy(() =>
  import('./pages/jobseeker/EvaluationPage').then((m) => ({ default: m.EvaluationPage })),
);
const ApplicationsPage = lazy(() =>
  import('./pages/jobseeker/ApplicationsPage').then((m) => ({ default: m.ApplicationsPage })),
);
const MatchesPage = lazy(() => import('./pages/jobseeker/MatchesPage').then((m) => ({ default: m.MatchesPage })));
const ForgotPassword = lazy(() => import('./pages/ForgotPassword').then((m) => ({ default: m.ForgotPassword })));
const ResetPassword = lazy(() => import('./pages/ResetPassword').then((m) => ({ default: m.ResetPassword })));
const VerifyEmail = lazy(() => import('./pages/VerifyEmail').then((m) => ({ default: m.VerifyEmail })));

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    children: [
      { path: '/', element: <Landing /> },
      { path: '/auth/login', element: <Login /> },
      { path: '/auth/register', element: <Register /> },
      { path: '/auth/forgot-password', element: <ForgotPassword /> },
      { path: '/auth/reset-password', element: <ResetPassword /> },
      { path: '/auth/verify-email', element: <VerifyEmail /> },
      {
        path: '/jobseeker/profile',
        element: (
          <ProtectedRoute requiredRole="jobseeker">
            <JobseekerProfile />
          </ProtectedRoute>
        ),
      },
      {
        path: '/jobseeker/passport',
        element: (
          <ProtectedRoute requiredRole="jobseeker">
            <JobseekerPassport />
          </ProtectedRoute>
        ),
      },
      {
        path: '/jobseeker/evaluation',
        element: (
          <ProtectedRoute requiredRole="jobseeker">
            <EvaluationPage />
          </ProtectedRoute>
        ),
      },
      {
        path: '/jobseeker/applications',
        element: (
          <ProtectedRoute requiredRole="jobseeker">
            <ApplicationsPage />
          </ProtectedRoute>
        ),
      },
      {
        path: '/jobseeker/matches',
        element: (
          <ProtectedRoute requiredRole="jobseeker">
            <MatchesPage />
          </ProtectedRoute>
        ),
      },
      {
        path: '/company/profile',
        element: (
          <ProtectedRoute requiredRole="company">
            <CompanyProfile />
          </ProtectedRoute>
        ),
      },
      {
        path: '/company/verification',
        element: (
          <ProtectedRoute requiredRole="company">
            <CompanyVerification />
          </ProtectedRoute>
        ),
      },
      {
        path: '/company/search',
        element: (
          <ProtectedRoute requiredRole="company">
            <CompanySearch />
          </ProtectedRoute>
        ),
      },
      {
        path: '/company/jobs',
        element: (
          <ProtectedRoute requiredRole="company">
            <CompanyJobs />
          </ProtectedRoute>
        ),
      },
      {
        path: '/company/applications',
        element: (
          <ProtectedRoute requiredRole="company">
            <CompanyApplications />
          </ProtectedRoute>
        ),
      },
      {
        path: '/company/analytics',
        element: (
          <ProtectedRoute requiredRole="company">
            <CompanyAnalytics />
          </ProtectedRoute>
        ),
      },
      {
        path: '/jobs',
        element: <PublicJobs />,
      },
      {
        path: '/jobs/:id',
        element: <JobDetail />,
      },
      {
        path: '/profiles/:username',
        element: <PublicPassport />,
      },
      {
        path: '/admin/verifications',
        element: (
          <ProtectedRoute requiredRole="admin">
            <AdminVerifications />
          </ProtectedRoute>
        ),
      },
    ],
  },
]);

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <ErrorBoundary>
          <Suspense fallback={<LoadingFallback />}>
            <RouterProvider router={router} />
          </Suspense>
        </ErrorBoundary>
      </AuthProvider>
    </QueryClientProvider>
  );
}
