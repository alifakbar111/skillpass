import { useQuery } from '@tanstack/react-query';
import { useAuth } from '@/hooks/useAuth';
import { getLeaveBalances, type LeaveBalance } from '@/lib/hris/leave';

export default function LeaveBalancePage() {
  const { user } = useAuth();
  const year = new Date().getFullYear();

  const { data: balances, isLoading } = useQuery<LeaveBalance[]>({
    queryKey: ['hris', 'leave-balances', user?.id, year],
    queryFn: () => getLeaveBalances(user?.id ?? '', year),
    enabled: !!user?.id,
  });

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">My Leave Balance — {year}</h1>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {balances?.map((b) => (
          <div key={b.id} className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
            <div className="stat-title">{b.leaveTypeName}</div>
            <div className="stat-value text-2xl">{b.remaining}</div>
            <div className="stat-desc">
              {b.usedDays} used / {b.totalDays + b.carryOverDays} total
              {b.carryOverDays > 0 && ` (${b.carryOverDays} carried over)`}
            </div>
            <progress
              className="progress progress-primary w-full mt-2"
              value={b.usedDays}
              max={b.totalDays + b.carryOverDays}
            />
          </div>
        ))}
        {balances?.length === 0 && (
          <div className="col-span-full text-center py-12 text-base-content/50">
            No leave balances found. Ask your HR to initialize balances.
          </div>
        )}
      </div>
    </div>
  );
}
