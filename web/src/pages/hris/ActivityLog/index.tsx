import { useQuery } from '@tanstack/react-query';
import { ScrollText } from 'lucide-react';
import { useState } from 'react';
import { TableSkeleton } from '@/components/ui/Skeleton';
import { listActivityLogs } from '@/lib/hris/activity';

const ACTION_LABELS: Record<string, string> = {
  'role.created': 'Created role',
  'role.updated': 'Updated role',
  'role.deleted': 'Deleted role',
  'role.permissions_set': 'Updated role permissions',
  'role.assigned': 'Assigned role to employee',
  'role.removed': 'Removed role from employee',
  'balance.adjusted': 'Adjusted leave balance',
  'balance.initialized': 'Initialized leave balances',
  'employee.created': 'Created employee',
  'employee.updated': 'Updated employee',
  'employee.terminated': 'Terminated employee',
  'template.updated': 'Updated onboarding template',
};

function formatAction(action: string): string {
  return ACTION_LABELS[action] || action.replace(/\./g, ' ');
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return 'Just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  const diffDay = Math.floor(diffHr / 24);
  if (diffDay < 7) return `${diffDay}d ago`;
  return d.toLocaleDateString();
}

export default function ActivityLogPage() {
  const [limit] = useState(100);

  const { data, isLoading } = useQuery({
    queryKey: ['hris', 'activity-logs', limit],
    queryFn: () => listActivityLogs(limit),
  });

  if (isLoading) {
    return (
      <div className="space-y-4">
        <h1 className="text-2xl font-bold">Activity Log</h1>
        <TableSkeleton rows={8} cols={4} />
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <ScrollText className="h-6 w-6 text-base-content/60" />
        <h1 className="text-2xl font-bold">Activity Log</h1>
        {data && <span className="text-sm text-base-content/40">{data.total} entries</span>}
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Time</th>
              <th>Actor</th>
              <th>Action</th>
              <th>Entity</th>
            </tr>
          </thead>
          <tbody>
            {data?.logs.map((log) => (
              <tr key={log.id} className="hover">
                <td
                  className="text-xs text-base-content/50 whitespace-nowrap"
                  title={new Date(log.createdAt).toLocaleString()}
                >
                  {formatTime(log.createdAt)}
                </td>
                <td className="font-medium text-sm">{log.actorName || log.actorId.slice(0, 8)}</td>
                <td>
                  <span className="badge badge-ghost badge-sm font-normal">{formatAction(log.action)}</span>
                </td>
                <td className="text-sm text-base-content/60">
                  {log.entityType}
                  {log.entityId ? ` #${log.entityId.slice(0, 8)}` : ''}
                </td>
              </tr>
            ))}
            {data?.logs.length === 0 && (
              <tr>
                <td colSpan={4} className="text-center py-12 text-base-content/50">
                  No activity logged yet. Actions will appear here as you use the system.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
