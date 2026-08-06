import { Navigate } from 'react-router-dom';
import { LoadingFallback } from '@/components/ui/LoadingFallback';
import { useAuth } from '@/hooks/useAuth';
import { usePermissions } from '@/hooks/usePermissions';

/**
 * HrisRoute gates the HRIS section by MEMBERSHIP, not role (Phase 2 · Sprint 5).
 *
 * A company account always has access. A hired jobseeker keeps their original
 * login and — now that they have an active employee record — also gains access
 * to their self-service HRIS pages (My Info / My Attendance / My Payslips).
 * Membership is proven by a successful /hris/me/permissions call.
 */
export function HrisRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth();
  const { isHrisUser, isLoading, error } = usePermissions();

  if (loading) return <LoadingFallback />;
  if (!user) return <Navigate to="/auth/login" replace />;

  // Company accounts are always HRIS-eligible.
  if (user.role === 'company') return <>{children}</>;

  // Otherwise wait for the membership check, then admit active employees only.
  if (isLoading) return <LoadingFallback />;
  if (error || !isHrisUser) return <Navigate to="/" replace />;

  return <>{children}</>;
}
