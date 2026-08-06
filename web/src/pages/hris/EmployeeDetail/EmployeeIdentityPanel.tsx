import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BadgeCheck, Fingerprint, ShieldX } from 'lucide-react';
import { ApiError } from '@/lib/api';
import { getEmployeeDid, issueEmployeeDid, listEmployeeCredentials } from '@/lib/hris/identity';
import { usePermissions } from '@/hooks/usePermissions';

export function EmployeeIdentityPanel({ employeeId }: { employeeId: string }) {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canIssue = hasPermission('org.manage') || hasPermission('employee.update');

  const didQuery = useQuery({
    queryKey: ['hris', 'did', employeeId],
    queryFn: () => getEmployeeDid(employeeId),
    retry: false, // 404 = no DID yet
  });
  const hasDid = didQuery.isSuccess;

  const { data: creds = [] } = useQuery({
    queryKey: ['hris', 'credentials', employeeId],
    queryFn: () => listEmployeeCredentials(employeeId),
    enabled: hasDid,
  });

  const issueMutation = useMutation({
    mutationFn: () => issueEmployeeDid(employeeId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'did', employeeId] });
      qc.invalidateQueries({ queryKey: ['hris', 'credentials', employeeId] });
    },
  });

  const did = didQuery.data;

  return (
    <div className="mt-6 rounded-box border border-base-300 bg-base-100 p-5">
      <h2 className="mb-4 flex items-center gap-2 text-base font-semibold">
        <Fingerprint className="h-5 w-5 text-primary" />
        Verifiable Identity
        <span className="badge badge-ghost badge-sm">signature-based · no blockchain</span>
      </h2>

      {!hasDid ? (
        <div className="flex items-center gap-3">
          <span className="text-sm text-base-content/60">No decentralized identifier issued yet.</span>
          {canIssue && (
            <button
              type="button"
              className="btn btn-primary btn-sm gap-2"
              disabled={issueMutation.isPending}
              onClick={() => issueMutation.mutate()}
            >
              <Fingerprint className="h-4 w-4" />
              {issueMutation.isPending ? 'Issuing…' : 'Issue DID'}
            </button>
          )}
          {issueMutation.isError && (
            <span className="text-sm text-error">
              {issueMutation.error instanceof ApiError
                ? (issueMutation.error.serverMessage ?? issueMutation.error.message)
                : 'Failed to issue DID'}
            </span>
          )}
        </div>
      ) : (
        <>
          <dl className="grid gap-4 sm:grid-cols-2">
            <div>
              <dt className="text-xs font-medium uppercase tracking-wide text-base-content/50">DID</dt>
              <dd className="mt-0.5 break-all font-mono text-sm">{did?.did}</dd>
            </div>
            <div>
              <dt className="text-xs font-medium uppercase tracking-wide text-base-content/50">Public key (Ed25519)</dt>
              <dd className="mt-0.5 truncate font-mono text-xs text-base-content/70">{did?.publicKey}</dd>
            </div>
          </dl>
          <a
            className="mt-2 inline-block text-xs text-primary hover:underline"
            href={`/api/v1/did/${did?.id}`}
            target="_blank"
            rel="noreferrer"
          >
            View public DID document ↗
          </a>

          <div className="mt-5">
            <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-base-content/50">
              Signed credentials
            </p>
            {creds.length === 0 ? (
              <p className="text-sm text-base-content/50">No credentials issued yet.</p>
            ) : (
              <ul className="space-y-2">
                {creds.map((c) => (
                  <li
                    key={c.id}
                    className="flex items-center justify-between rounded-lg border border-base-300 px-3 py-2"
                  >
                    <div className="min-w-0">
                      <p className="text-sm font-medium capitalize">{c.credentialType}</p>
                      <p className="truncate font-mono text-xs text-base-content/50">{c.issuerDid}</p>
                    </div>
                    {c.verified ? (
                      <span className="badge badge-success gap-1">
                        <BadgeCheck className="h-3.5 w-3.5" /> Verified
                      </span>
                    ) : (
                      <span className="badge badge-error gap-1">
                        <ShieldX className="h-3.5 w-3.5" /> Invalid
                      </span>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </>
      )}
    </div>
  );
}
