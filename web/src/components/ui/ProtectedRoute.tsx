import { Navigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

interface Props {
  children: React.ReactNode;
  role?: 'jobseeker' | 'company';
}

export function ProtectedRoute({ children, role }: Props) {
  const { user, loading } = useAuth();

  if (loading) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg"></span></div>;
  if (!user) return <Navigate to="/auth/login" replace />;
  if (role && user.role !== role) return <Navigate to="/" replace />;

  return <>{children}</>;
}
