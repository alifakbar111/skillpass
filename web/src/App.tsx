import { Routes, Route } from 'react-router-dom';
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

export function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<RootLayout />}>
          <Route path="/" element={<Landing />} />
          <Route path="/auth/login" element={<Login />} />
          <Route path="/auth/register" element={<Register />} />
          <Route path="/jobseeker/profile" element={<JobseekerProfile />} />
          <Route path="/jobseeker/passport" element={<JobseekerPassport />} />
          <Route path="/company/profile" element={<CompanyProfile />} />
          <Route path="/company/verification" element={<CompanyVerification />} />
          <Route path="/company/search" element={<CompanySearch />} />
          <Route path="/company/jobs" element={<CompanyJobs />} />
          <Route path="/jobs" element={<PublicJobs />} />
          <Route path="/jobs/:id" element={<JobDetail />} />
          <Route path="/profiles/:username" element={<PublicPassport />} />
          <Route path="/admin/verifications" element={<AdminVerifications />} />
        </Route>
      </Routes>
    </AuthProvider>
  );
}
