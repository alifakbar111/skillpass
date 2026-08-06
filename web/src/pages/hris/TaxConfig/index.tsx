import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Landmark, Percent } from 'lucide-react';
import { useEffect, useState } from 'react';
import { PageHeader } from '@/components/hris/ui';
import { ApiError } from '@/lib/api';
import { type BpjsConfig, getBpjsConfig, updateBpjsConfig } from '@/lib/hris/payroll';
import { usePermissions } from '@/hooks/usePermissions';

// Fields shown as percentages (rate * 100) vs. absolute Rupiah caps.
const RATE_FIELDS: { key: keyof BpjsConfig; label: string }[] = [
  { key: 'kesehatanEmployee', label: 'Kesehatan — employee' },
  { key: 'kesehatanEmployer', label: 'Kesehatan — employer' },
  { key: 'jhtEmployee', label: 'JHT — employee' },
  { key: 'jhtEmployer', label: 'JHT — employer' },
  { key: 'jkkEmployer', label: 'JKK — employer' },
  { key: 'jkmEmployer', label: 'JKM — employer' },
  { key: 'jpEmployee', label: 'JP — employee' },
  { key: 'jpEmployer', label: 'JP — employer' },
];
const CAP_FIELDS: { key: keyof BpjsConfig; label: string }[] = [
  { key: 'kesehatanCap', label: 'Kesehatan salary cap' },
  { key: 'jpCap', label: 'JP (pension) salary cap' },
];

export default function TaxConfig() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('payroll.manage');

  const { data, isLoading } = useQuery({ queryKey: ['hris', 'bpjs-config'], queryFn: getBpjsConfig });
  const [form, setForm] = useState<BpjsConfig | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (data) setForm(data);
  }, [data]);

  const saveMutation = useMutation({
    mutationFn: () => updateBpjsConfig(form as BpjsConfig),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'bpjs-config'] });
      setError(null);
    },
    onError: (err) => setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save'),
  });

  if (isLoading || !form) {
    return (
      <div className="flex justify-center p-16">
        <span className="loading loading-spinner loading-lg text-primary" />
      </div>
    );
  }

  const setRate = (key: keyof BpjsConfig, pct: string) =>
    setForm((f) => (f ? { ...f, [key]: (Number(pct) || 0) / 100 } : f));
  const setCap = (key: keyof BpjsConfig, val: string) => setForm((f) => (f ? { ...f, [key]: Number(val) || 0 } : f));

  return (
    <div className="max-w-3xl">
      <PageHeader
        title="Tax & BPJS Configuration"
        subtitle="BPJS contribution rates and salary caps used by payroll."
        actions={
          canManage && (
            <button
              type="button"
              className="btn btn-primary btn-sm"
              disabled={saveMutation.isPending}
              onClick={() => saveMutation.mutate()}
            >
              {saveMutation.isPending ? 'Saving…' : 'Save changes'}
            </button>
          )
        }
      />

      {error && <div className="alert alert-error mb-4 text-sm">{error}</div>}

      <div className="mb-6 rounded-box border border-base-300 bg-base-100 p-5">
        <h3 className="mb-4 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-base-content/50">
          <Percent className="h-4 w-4" /> Contribution rates
        </h3>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {RATE_FIELDS.map((f) => (
            <label key={f.key} className="form-control">
              <span className="label-text mb-1 text-xs text-base-content/60">{f.label}</span>
              <div className="join">
                <input
                  type="number"
                  step="0.01"
                  disabled={!canManage}
                  className="input input-sm input-bordered join-item w-full"
                  value={((form[f.key] as number) * 100).toString()}
                  onChange={(e) => setRate(f.key, e.target.value)}
                />
                <span className="join-item btn btn-sm btn-disabled">%</span>
              </div>
            </label>
          ))}
        </div>
      </div>

      <div className="mb-6 rounded-box border border-base-300 bg-base-100 p-5">
        <h3 className="mb-4 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-base-content/50">
          <Landmark className="h-4 w-4" /> Salary caps (Rp)
        </h3>
        <div className="grid gap-4 sm:grid-cols-2">
          {CAP_FIELDS.map((f) => (
            <label key={f.key} className="form-control">
              <span className="label-text mb-1 text-xs text-base-content/60">{f.label}</span>
              <input
                type="number"
                disabled={!canManage}
                className="input input-sm input-bordered"
                value={(form[f.key] as number).toString()}
                onChange={(e) => setCap(f.key, e.target.value)}
              />
            </label>
          ))}
        </div>
      </div>

      <div className="rounded-box border border-base-300 bg-base-200/50 p-5 text-sm text-base-content/70">
        <p className="font-medium text-base-content">PPh21 — TER method (PMK 168/2023)</p>
        <p className="mt-1">
          Monthly income tax uses the average-effective-rate (TER) tables by PTKP category (A/B/C), applied
          automatically during each payroll run. Each employee's PTKP status is set on their record.
        </p>
      </div>
    </div>
  );
}
