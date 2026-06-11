import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export interface Industry {
  id: string;
  name: string;
  description: string | null;
}

export function useIndustries() {
  return useQuery({
    queryKey: ['industries'],
    queryFn: () => api<Industry[]>('/industries'),
    staleTime: 5 * 60_000, // reference data: rarely changes
  });
}
