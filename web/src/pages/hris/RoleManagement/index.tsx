import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, Plus, Save, Shield, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { PageHeader } from '@/components/hris/ui';
import { usePermissions } from '@/hooks/usePermissions';
import {
  createRole,
  deleteRole,
  getRolePermissionIds,
  type HrisPermission,
  type HrisRole,
  listPermissions,
  listRoles,
  setRolePermissions,
} from '@/lib/hris/rbac';

const MODULE_LABELS: Record<string, string> = {
  employee: 'Employee',
  attendance: 'Attendance',
  leave: 'Leave',
  payroll: 'Payroll',
  performance: 'Performance',
  ats: 'Recruitment (ATS)',
  analytics: 'Analytics',
  documents: 'Documents',
  face: 'Face / Biometrics',
  org: 'Organisation & Settings',
};

export default function RoleManagement() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('roles.manage');

  const { data: roles = [], isLoading: rolesLoading } = useQuery({ queryKey: ['hris', 'roles'], queryFn: listRoles });
  const { data: permissions = [] } = useQuery({ queryKey: ['hris', 'permissions'], queryFn: listPermissions });

  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);
  const selectedRole = roles.find((r) => r.id === selectedRoleId) ?? null;

  // Default-select the first role once loaded.
  useEffect(() => {
    if (!selectedRoleId && roles.length > 0) setSelectedRoleId(roles[0].id);
  }, [roles, selectedRoleId]);

  const { data: rolePermIds = [] } = useQuery({
    queryKey: ['hris', 'role-permissions', selectedRoleId],
    queryFn: () => getRolePermissionIds(selectedRoleId ?? ''),
    enabled: !!selectedRoleId,
  });

  // Local editable checked-set, seeded from the server each time role changes.
  const [checked, setChecked] = useState<Set<string>>(new Set());
  useEffect(() => {
    setChecked(new Set(rolePermIds));
  }, [rolePermIds]);

  const dirty = useMemo(() => {
    const server = new Set(rolePermIds);
    if (server.size !== checked.size) return true;
    for (const id of checked) if (!server.has(id)) return true;
    return false;
  }, [checked, rolePermIds]);

  const grouped = useMemo(() => {
    const map = new Map<string, HrisPermission[]>();
    for (const p of permissions) {
      const list = map.get(p.module) ?? [];
      list.push(p);
      map.set(p.module, list);
    }
    return [...map.entries()].sort((a, b) => a[0].localeCompare(b[0]));
  }, [permissions]);

  const saveMutation = useMutation({
    mutationFn: () => setRolePermissions(selectedRoleId ?? '', [...checked]),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'role-permissions', selectedRoleId] }),
  });

  const createMutation = useMutation({
    mutationFn: (name: string) => createRole(name),
    onSuccess: (role) => {
      qc.invalidateQueries({ queryKey: ['hris', 'roles'] });
      setSelectedRoleId(role.id);
      setNewRoleName('');
      setCreating(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteRole(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'roles'] });
      setSelectedRoleId(null);
    },
  });

  const [creating, setCreating] = useState(false);
  const [newRoleName, setNewRoleName] = useState('');

  function toggle(id: string) {
    if (!canManage) return;
    setChecked((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function toggleModule(mod: string, on: boolean) {
    if (!canManage) return;
    const ids = (grouped.find(([m]) => m === mod)?.[1] ?? []).map((p) => p.id);
    setChecked((prev) => {
      const next = new Set(prev);
      for (const id of ids) {
        if (on) next.add(id);
        else next.delete(id);
      }
      return next;
    });
  }

  return (
    <div>
      <PageHeader
        title="Roles & Permissions"
        subtitle="Control what each role can see and do across HRIS modules."
        actions={
          canManage && (
            <button type="button" className="btn btn-primary btn-sm gap-2" onClick={() => setCreating((v) => !v)}>
              <Plus className="h-4 w-4" /> New role
            </button>
          )
        }
      />

      <div className="grid gap-6 lg:grid-cols-[16rem_1fr]">
        {/* Roles list */}
        <aside className="rounded-box border border-base-300 bg-base-100 p-2 h-fit">
          {creating && (
            <form
              className="p-2"
              onSubmit={(e) => {
                e.preventDefault();
                if (newRoleName.trim()) createMutation.mutate(newRoleName.trim());
              }}
            >
              <input
                autoFocus
                className="input input-sm input-bordered w-full mb-2"
                placeholder="Role name…"
                value={newRoleName}
                onChange={(e) => setNewRoleName(e.target.value)}
              />
              <div className="flex gap-2">
                <button type="submit" className="btn btn-primary btn-xs flex-1" disabled={createMutation.isPending}>
                  Create
                </button>
                <button type="button" className="btn btn-ghost btn-xs" onClick={() => setCreating(false)}>
                  Cancel
                </button>
              </div>
            </form>
          )}
          {rolesLoading ? (
            <div className="p-4 text-center">
              <span className="loading loading-spinner" />
            </div>
          ) : (
            <ul className="flex flex-col gap-0.5">
              {roles.map((role: HrisRole) => (
                <li key={role.id}>
                  <button
                    type="button"
                    onClick={() => setSelectedRoleId(role.id)}
                    className={`flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors ${
                      selectedRoleId === role.id
                        ? 'bg-primary/10 text-primary font-medium'
                        : 'text-base-content/70 hover:bg-base-200'
                    }`}
                  >
                    <Shield className="h-4 w-4 shrink-0" />
                    <span className="flex-1 truncate">{role.name}</span>
                    {role.isSystem && <span className="badge badge-ghost badge-xs">system</span>}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </aside>

        {/* Permission matrix */}
        <section className="rounded-box border border-base-300 bg-base-100">
          {!selectedRole ? (
            <div className="p-10 text-center text-base-content/50">Select a role to view its permissions.</div>
          ) : (
            <>
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-base-300 p-4">
                <div>
                  <h2 className="text-lg font-semibold flex items-center gap-2">
                    {selectedRole.name}
                    {selectedRole.isSystem && <span className="badge badge-ghost badge-sm">system role</span>}
                  </h2>
                  {selectedRole.description && (
                    <p className="text-sm text-base-content/60">{selectedRole.description}</p>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {canManage && !selectedRole.isSystem && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-sm text-error gap-1"
                      onClick={() => {
                        if (confirm(`Delete role "${selectedRole.name}"?`)) deleteMutation.mutate(selectedRole.id);
                      }}
                    >
                      <Trash2 className="h-4 w-4" /> Delete
                    </button>
                  )}
                  {canManage && (
                    <button
                      type="button"
                      className="btn btn-primary btn-sm gap-2"
                      disabled={!dirty || saveMutation.isPending}
                      onClick={() => saveMutation.mutate()}
                    >
                      <Save className="h-4 w-4" />
                      {saveMutation.isPending ? 'Saving…' : dirty ? 'Save changes' : 'Saved'}
                    </button>
                  )}
                </div>
              </div>

              <div className="divide-y divide-base-300">
                {grouped.map(([mod, perms]) => {
                  const allOn = perms.every((p) => checked.has(p.id));
                  return (
                    <div key={mod} className="p-4">
                      <div className="mb-2 flex items-center justify-between">
                        <h3 className="text-xs font-semibold uppercase tracking-wider text-base-content/50">
                          {MODULE_LABELS[mod] ?? mod}
                        </h3>
                        {canManage && (
                          <button
                            type="button"
                            className="text-xs text-primary hover:underline"
                            onClick={() => toggleModule(mod, !allOn)}
                          >
                            {allOn ? 'Clear all' : 'Select all'}
                          </button>
                        )}
                      </div>
                      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                        {perms.map((p) => {
                          const on = checked.has(p.id);
                          return (
                            <button
                              key={p.id}
                              type="button"
                              disabled={!canManage}
                              onClick={() => toggle(p.id)}
                              className={`flex items-start gap-2 rounded-lg border px-3 py-2 text-left text-sm transition-colors ${
                                on ? 'border-primary/40 bg-primary/5' : 'border-base-300 hover:bg-base-200'
                              } ${canManage ? 'cursor-pointer' : 'cursor-default'}`}
                            >
                              <span
                                className={`mt-0.5 grid h-4 w-4 shrink-0 place-items-center rounded border ${
                                  on ? 'border-primary bg-primary text-primary-content' : 'border-base-content/30'
                                }`}
                              >
                                {on && <Check className="h-3 w-3" />}
                              </span>
                              <span className="min-w-0">
                                <span className="block font-mono text-xs text-base-content/80">{p.code}</span>
                                <span className="block text-xs text-base-content/50">{p.description}</span>
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </section>
      </div>
    </div>
  );
}
