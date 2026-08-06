import { useQuery } from '@tanstack/react-query';
import { getMyPermissions } from '@/lib/hris/rbac';

export function usePermissions() {
  const { data, isLoading, error } = useQuery({
    // Distinct from RoleManagement's ['hris', 'all-permissions'] — sharing a
    // key made both queries evict each other's cache (F1/H-2).
    queryKey: ['hris', 'my-permissions'],
    queryFn: getMyPermissions,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  const permissions = data?.permissions ?? [];

  function hasPermission(code: string): boolean {
    return permissions.includes(code);
  }

  function hasAnyPermission(...codes: string[]): boolean {
    return codes.some((c) => permissions.includes(c));
  }

  return {
    permissions,
    roles: data?.roles ?? [],
    hasPermission,
    hasAnyPermission,
    isLoading,
    isHrisUser: !error && !!data,
    error,
  };
}
