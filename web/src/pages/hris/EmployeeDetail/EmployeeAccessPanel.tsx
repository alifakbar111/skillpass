import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, KeyRound, ShieldCheck, UserPlus } from 'lucide-react';
import { useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { ApiError } from '@/lib/api';
import { inviteEmployeeLogin } from '@/lib/hris/employees';
import { assignRole, getEmployeeRoles, listRoles, removeRole } from '@/lib/hris/rbac';

interface Props {
  employeeId: string;
  hasLogin: boolean;
}

export function EmployeeAccessPanel({ employeeId, hasLogin }: Props) {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canInvite = hasPermission('employee.update');
  const canManageRoles = hasPermission('roles.manage');

  const { data: allRoles = [] } = useQuery({ queryKey: ['hris', 'roles'], queryFn: listRoles });
  const { data: empRoles = [] } = useQuery({
    queryKey: ['hris', 'employee-roles', employeeId],
    queryFn: () => getEmployeeRoles(employeeId),
  });
  const assignedIds = new Set(empRoles.map((r) => r.id));

  const [invite, setInvite] = useState<{ email: string; tempPassword: string } | null>(null);
  const [error, setError] = useState<string | null>(null);

  const inviteMutation = useMutation({
    mutationFn: () => inviteEmployeeLogin(employeeId),
    onSuccess: (res) => {
      setInvite(res);
      setError(null);
      qc.invalidateQueries({ queryKey: ['hris', 'employee', employeeId] });
    },
    onError: (err) => setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Invite failed'),
  });

  const toggleRole = useMutation({
    mutationFn: ({ roleId, on }: { roleId: string; on: boolean }) =>
      on ? assignRole(employeeId, roleId) : removeRole(employeeId, roleId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'employee-roles', employeeId] }),
    onError: (err) => setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to update role'),
  });

  return (
    <div className="mt-6 rounded-box border border-base-300 bg-base-100 p-5">
      <h2 className="mb-4 flex items-center gap-2 text-base font-semibold">
        <ShieldCheck className="h-5 w-5 text-primary" />
        Access &amp; Roles
      </h2>

      {error && <div className="alert alert-error mb-4 text-sm">{error}</div>}

      {/* Login access */}
      <div className="mb-5">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-base-content/50">Login access</p>
        {hasLogin || invite ? (
          <div className="flex items-center gap-2 text-sm text-success">
            <Check className="h-4 w-4" /> This employee has a login account.
          </div>
        ) : (
          <div className="flex items-center gap-3">
            <span className="text-sm text-base-content/60">No login yet.</span>
            {canInvite && (
              <button
                type="button"
                className="btn btn-primary btn-sm gap-2"
                disabled={inviteMutation.isPending}
                onClick={() => inviteMutation.mutate()}
              >
                <UserPlus className="h-4 w-4" />
                {inviteMutation.isPending ? 'Creating…' : 'Invite to HRIS'}
              </button>
            )}
          </div>
        )}

        {invite && (
          <div className="mt-3 rounded-lg border border-success/30 bg-success/5 p-3 text-sm">
            <p className="flex items-center gap-2 font-medium text-success">
              <KeyRound className="h-4 w-4" /> Login created — share these once:
            </p>
            <div className="mt-2 grid gap-1 font-mono text-xs">
              <span>Email: {invite.email}</span>
              <span>Temporary password: {invite.tempPassword}</span>
            </div>
            <p className="mt-2 text-xs text-base-content/50">
              Ask them to sign in and change their password. This password won't be shown again.
            </p>
          </div>
        )}
      </div>

      {/* Roles */}
      <div>
        <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-base-content/50">HRIS roles</p>
        <div className="grid gap-2 sm:grid-cols-2">
          {allRoles.map((role) => {
            const on = assignedIds.has(role.id);
            return (
              <button
                key={role.id}
                type="button"
                disabled={!canManageRoles || toggleRole.isPending}
                onClick={() => toggleRole.mutate({ roleId: role.id, on: !on })}
                className={`flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-sm transition-colors ${
                  on ? 'border-primary/40 bg-primary/5' : 'border-base-300 hover:bg-base-200'
                } ${canManageRoles ? 'cursor-pointer' : 'cursor-default'}`}
              >
                <span
                  className={`grid h-4 w-4 shrink-0 place-items-center rounded border ${
                    on ? 'border-primary bg-primary text-primary-content' : 'border-base-content/30'
                  }`}
                >
                  {on && <Check className="h-3 w-3" />}
                </span>
                <span className="flex-1 truncate">{role.name}</span>
                {role.isSystem && <span className="badge badge-ghost badge-xs">system</span>}
              </button>
            );
          })}
        </div>
        {!canManageRoles && (
          <p className="mt-2 text-xs text-base-content/40">You need “Manage roles” permission to change roles.</p>
        )}
      </div>
    </div>
  );
}
