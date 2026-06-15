import { Navigate } from 'react-router-dom';
import { LoadingFallback } from '@/components/ui/LoadingFallback';
import { useAuth } from '@/hooks/useAuth';

interface Props {
  children: React.ReactNode;
  requiredRole?: 'jobseeker' | 'company' | 'admin';
}

export function ProtectedRoute({ children, requiredRole }: Props) {
  const { user, loading } = useAuth();

  if (loading) return <LoadingFallback />;
  if (!user) return <Navigate to="/auth/login" replace />;
  if (requiredRole && user.role !== requiredRole) return <Navigate to="/" replace />;

  return <>{children}</>;
}
