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

const HRISLayout = lazy(() => import('./components/hris/HRISLayout'));
const EmployeeList = lazy(() => import('./pages/hris/EmployeeList'));
const EmployeeCreate = lazy(() => import('./pages/hris/EmployeeCreate'));
const EmployeeDetail = lazy(() => import('./pages/hris/EmployeeDetail'));
const BranchManagement = lazy(() => import('./pages/hris/BranchManagement'));
const DepartmentManagement = lazy(() => import('./pages/hris/DepartmentManagement'));
const PositionManagement = lazy(() => import('./pages/hris/PositionManagement'));
const OrgChart = lazy(() => import('./pages/hris/OrgChart'));
const RoleManagement = lazy(() => import('./pages/hris/RoleManagement'));
const ShiftConfig = lazy(() => import('./pages/hris/ShiftConfig'));
const ClockInPage = lazy(() => import('./pages/hris/ClockIn'));
const AttendanceDashboard = lazy(() => import('./pages/hris/AttendanceDashboard'));
const MyAttendance = lazy(() => import('./pages/hris/MyAttendance'));
const AttendanceExceptions = lazy(() => import('./pages/hris/AttendanceExceptions'));
const LeaveTypes = lazy(() => import('./pages/hris/LeaveTypes'));
const LeaveRequest = lazy(() => import('./pages/hris/LeaveRequest'));
const LeaveApproval = lazy(() => import('./pages/hris/LeaveApproval'));
const LeaveBalance = lazy(() => import('./pages/hris/LeaveBalance'));
const Holidays = lazy(() => import('./pages/hris/Holidays'));
const SalaryComponents = lazy(() => import('./pages/hris/SalaryComponents'));
const EmployeeSalary = lazy(() => import('./pages/hris/EmployeeSalary'));
const PayrollRuns = lazy(() => import('./pages/hris/PayrollRuns'));
const PayslipView = lazy(() => import('./pages/hris/PayslipView'));
const MyPayslips = lazy(() => import('./pages/hris/MyPayslips'));

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
      {
        path: '/hris',
        element: (
          <ProtectedRoute requiredRole="company">
            <HRISLayout />
          </ProtectedRoute>
        ),
        children: [
          { index: true, element: <EmployeeList /> },
          { path: 'employees', element: <EmployeeList /> },
          { path: 'employees/new', element: <EmployeeCreate /> },
          { path: 'employees/:id', element: <EmployeeDetail /> },
          { path: 'branches', element: <BranchManagement /> },
          { path: 'departments', element: <DepartmentManagement /> },
          { path: 'positions', element: <PositionManagement /> },
          { path: 'org-chart', element: <OrgChart /> },
          { path: 'roles', element: <RoleManagement /> },
          { path: 'shifts', element: <ShiftConfig /> },
          { path: 'clock-in', element: <ClockInPage /> },
          { path: 'attendance', element: <AttendanceDashboard /> },
          { path: 'my-attendance', element: <MyAttendance /> },
          { path: 'attendance-exceptions', element: <AttendanceExceptions /> },
          { path: 'leave-types', element: <LeaveTypes /> },
          { path: 'leave-request', element: <LeaveRequest /> },
          { path: 'leave-approval', element: <LeaveApproval /> },
          { path: 'leave-balance', element: <LeaveBalance /> },
          { path: 'holidays', element: <Holidays /> },
          { path: 'salary-components', element: <SalaryComponents /> },
          { path: 'employees/:id/salary', element: <EmployeeSalary /> },
          { path: 'payroll-runs', element: <PayrollRuns /> },
          { path: 'payroll-runs/:runId/payslips', element: <PayslipView /> },
          { path: 'my-payslips', element: <MyPayslips /> },
        ],
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
