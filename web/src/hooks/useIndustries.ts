import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { Industry } from '../pages/CompanyJobs/type';

export type { Industry };

export function useIndustries() {
  return useQuery({
    queryKey: ['industries'],
    queryFn: () => api<Industry[]>('/industries'),
    staleTime: 5 * 60_000, // reference data: rarely changes
  });
}
