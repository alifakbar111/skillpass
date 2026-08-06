import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BadgeCheck, ExternalLink, ShieldCheck } from 'lucide-react';
import { useState } from 'react';
import { VerifiedBadge } from '@/components/VerifiedBadge';
import { usePermissions } from '@/hooks/usePermissions';
import { listEmployees } from '@/lib/hris/employees';
import {
  attestEmployeeSkills,
  getEmployeePassport,
  type IdentityProvider,
  listEmployeeAttestations,
  listVerifications,
  runVerification,
  setEmployeePassportVisibility,
} from '@/lib/hris/identity';

export default function IdentityVerification() {
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage') || hasPermission('employee.update');
  const canAttest = hasPermission('performance.manage');
  const [selected, setSelected] = useState<string>('');

  const { data: employeesResult, isLoading } = useQuery({
    queryKey: ['hris', 'employees', 'identity-list'],
    queryFn: () => listEmployees({ pageSize: 100 }),
  });
  const employees = employeesResult?.employees ?? [];
  const active = selected || employees[0]?.id || '';

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tight">Identity &amp; Credentials</h1>
        <p className="text-sm text-base-content/60">
          Verify identities against Dukcapil / PDDikti and issue signed, verifiable skill credentials — no blockchain.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-[260px_1fr]">
        <aside className="rounded-xl border border-base-300 bg-base-100 p-2">
          <ul className="max-h-[70vh] space-y-0.5 overflow-y-auto">
            {employees.map((e) => (
              <li key={e.id}>
                <button
                  type="button"
                  onClick={() => setSelected(e.id)}
                  className={[
                    'w-full rounded-lg px-3 py-2 text-left text-sm',
                    e.id === active ? 'bg-primary/10 text-primary' : 'hover:bg-base-200',
                  ].join(' ')}
                >
                  <span className="font-medium">
                    {e.firstName} {e.lastName}
                  </span>
                  <span className="block truncate text-xs text-base-content/50">{e.email}</span>
                </button>
              </li>
            ))}
            {employees.length === 0 && <p className="p-4 text-center text-sm text-base-content/50">No employees.</p>}
          </ul>
        </aside>

        {active ? (
          <EmployeeIdentityPanel employeeId={active} canManage={canManage} canAttest={canAttest} />
        ) : (
          <div className="rounded-xl border border-dashed border-base-300 p-8 text-center text-base-content/50">
            Select an employee.
          </div>
        )}
      </div>
    </div>
  );
}

function EmployeeIdentityPanel({
  employeeId,
  canManage,
  canAttest,
}: {
  employeeId: string;
  canManage: boolean;
  canAttest: boolean;
}) {
  const qc = useQueryClient();
  const invalidate = () => qc.invalidateQueries({ queryKey: ['hris', 'identity', employeeId] });

  const { data: verifications } = useQuery({
    queryKey: ['hris', 'identity', employeeId, 'verifications'],
    queryFn: () => listVerifications(employeeId),
  });
  const { data: attestations } = useQuery({
    queryKey: ['hris', 'identity', employeeId, 'attestations'],
    queryFn: () => listEmployeeAttestations(employeeId),
  });
  const { data: passport } = useQuery({
    queryKey: ['hris', 'identity', employeeId, 'passport'],
    queryFn: () => getEmployeePassport(employeeId),
  });

  const verifyMut = useMutation({
    mutationFn: (provider: IdentityProvider) => runVerification(employeeId, provider),
    onSuccess: invalidate,
  });
  const attestMut = useMutation({ mutationFn: () => attestEmployeeSkills(employeeId), onSuccess: invalidate });
  const passportMut = useMutation({
    mutationFn: (isPublic: boolean) => setEmployeePassportVisibility(employeeId, isPublic),
    onSuccess: invalidate,
  });

  return (
    <div className="space-y-4">
      <section className="rounded-xl border border-base-300 bg-base-100 p-5">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-lg font-semibold">
            <ShieldCheck className="h-5 w-5" /> Identity verification
          </h2>
          {canManage && (
            <div className="flex gap-1">
              {(['dukcapil', 'pddikti', 'manual'] as IdentityProvider[]).map((p) => (
                <button
                  key={p}
                  type="button"
                  className="btn btn-outline btn-xs"
                  disabled={verifyMut.isPending}
                  onClick={() => verifyMut.mutate(p)}
                >
                  {p}
                </button>
              ))}
            </div>
          )}
        </div>
        {verifications?.length ? (
          <ul className="space-y-2">
            {verifications.map((v) => (
              <li
                key={v.id}
                className="flex items-center justify-between rounded-lg border border-base-200 p-2 text-sm"
              >
                <span className="font-medium capitalize">{v.provider}</span>
                <div className="flex items-center gap-2">
                  {v.detail && <span className="text-xs text-base-content/50">{v.detail}</span>}
                  <span className={`badge badge-sm ${v.status === 'verified' ? 'badge-success' : 'badge-error'}`}>
                    {v.status}
                  </span>
                </div>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-base-content/50">No verification runs yet.</p>
        )}
      </section>

      <section className="rounded-xl border border-base-300 bg-base-100 p-5">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="flex items-center gap-2 text-lg font-semibold">
            <BadgeCheck className="h-5 w-5" /> Signed skill credentials
          </h2>
          {canAttest && (
            <button
              type="button"
              className="btn btn-primary btn-xs"
              disabled={attestMut.isPending}
              onClick={() => attestMut.mutate()}
            >
              Sign latest evaluation
            </button>
          )}
        </div>
        {attestMut.error && (
          <div className="alert alert-warning mb-2 text-sm">No current evaluation found for this employee.</div>
        )}
        {attestations?.length ? (
          <ul className="space-y-2">
            {attestations.map((a) => (
              <li
                key={a.id}
                className="flex items-center justify-between rounded-lg border border-base-200 p-2 text-sm"
              >
                <span className="font-medium">
                  {a.skillName} · {a.score}
                </span>
                <VerifiedBadge verified={a.verified && !a.revoked} credentialId={a.id} />
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-base-content/50">No signed credentials yet.</p>
        )}
      </section>

      {passport && (
        <section className="rounded-xl border border-base-300 bg-base-100 p-5">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg font-semibold">Public Skill Passport</h2>
              <p className="text-sm text-base-content/60">
                {passport.isPublic ? 'Public' : 'Private'} ·{' '}
                {passport.isPublic ? (
                  <a
                    href={`/passport/${passport.slug}`}
                    target="_blank"
                    rel="noreferrer"
                    className="link inline-flex items-center gap-1"
                  >
                    /passport/{passport.slug} <ExternalLink className="h-3 w-3" />
                  </a>
                ) : (
                  <span className="text-base-content/40">/passport/{passport.slug}</span>
                )}
              </p>
            </div>
            {canManage && (
              <input
                type="checkbox"
                className="toggle toggle-primary"
                checked={passport.isPublic}
                onChange={(e) => passportMut.mutate(e.target.checked)}
              />
            )}
          </div>
        </section>
      )}
    </div>
  );
}
