import { useQuery } from '@tanstack/react-query';
import { getMyPermissions } from '@/lib/hris/rbac';

export function usePermissions() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['hris', 'permissions'],
    queryFn: getMyPermissions,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  function hasPermission(code: string): boolean {
    return data?.permissions.includes(code) ?? false;
  }

  function hasAnyPermission(...codes: string[]): boolean {
    if (!data) return false;
    return codes.some((c) => data.permissions.includes(c));
  }

  return {
    permissions: data?.permissions ?? [],
    roles: data?.roles ?? [],
    hasPermission,
    hasAnyPermission,
    isLoading,
    isHrisUser: !error && !!data,
    error,
  };
}
