import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Save } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import { getEmployeeSalary, listSalaryComponents, type SalaryComponent, setEmployeeSalary } from '@/lib/hris/payroll';

export default function EmployeeSalaryPage() {
  const { id: employeeId } = useParams<{ id: string }>();
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('payroll.manage');
  const [amounts, setAmounts] = useState<Record<string, number>>({});
  const [error, setError] = useState('');

  const { data: components } = useQuery({
    queryKey: ['hris', 'salary-components'],
    queryFn: listSalaryComponents,
  });

  const { data: salary, isLoading } = useQuery({
    queryKey: ['hris', 'employee-salary', employeeId],
    queryFn: () => getEmployeeSalary(employeeId ?? ''),
    enabled: !!employeeId,
  });

  useEffect(() => {
    if (salary) {
      const map: Record<string, number> = {};
      for (const s of salary) {
        map[s.componentId] = s.amount;
      }
      setAmounts(map);
    }
  }, [salary]);

  const activeComponents = components?.filter((c: SalaryComponent) => c.isActive) ?? [];

  const saveMut = useMutation({
    mutationFn: () => {
      const items = Object.entries(amounts)
        .filter(([, amt]) => amt > 0)
        .map(([componentId, amount]) => ({ componentId, amount }));
      return setEmployeeSalary(employeeId ?? '', items);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'employee-salary', employeeId] });
      setError('');
    },
    onError: (err: Error) => setError(err.message),
  });

  const fmt = (n: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(n);

  const earnings = activeComponents.filter((c: SalaryComponent) => c.type === 'earning');
  const deductions = activeComponents.filter((c: SalaryComponent) => c.type === 'deduction');
  const totalEarnings = earnings.reduce((sum: number, c: SalaryComponent) => sum + (amounts[c.id] ?? 0), 0);
  const totalDeductions = deductions.reduce((sum: number, c: SalaryComponent) => sum + (amounts[c.id] ?? 0), 0);

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Employee Salary Setup</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm" onClick={() => saveMut.mutate()}>
            <Save className="h-4 w-4" /> Save
          </button>
        )}
      </div>

      {error && <div className="alert alert-error mb-4">{error}</div>}

      <div className="grid gap-6 md:grid-cols-2">
        <div>
          <h2 className="text-lg font-semibold mb-3 text-success">Earnings</h2>
          <div className="space-y-2">
            {earnings.map((c: SalaryComponent) => (
              <div key={c.id} className="flex items-center gap-3">
                <span className="w-40 text-sm font-medium">{c.name}</span>
                <input
                  type="number"
                  step="0.01"
                  value={amounts[c.id] ?? c.defaultAmount}
                  onChange={(e) => setAmounts({ ...amounts, [c.id]: Number(e.target.value) })}
                  className="input input-bordered input-sm flex-1 font-mono"
                  disabled={!canManage}
                />
              </div>
            ))}
            {earnings.length === 0 && <p className="text-base-content/50 text-sm">No earning components configured.</p>}
          </div>
        </div>

        <div>
          <h2 className="text-lg font-semibold mb-3 text-error">Deductions</h2>
          <div className="space-y-2">
            {deductions.map((c: SalaryComponent) => (
              <div key={c.id} className="flex items-center gap-3">
                <span className="w-40 text-sm font-medium">{c.name}</span>
                <input
                  type="number"
                  step="0.01"
                  value={amounts[c.id] ?? c.defaultAmount}
                  onChange={(e) => setAmounts({ ...amounts, [c.id]: Number(e.target.value) })}
                  className="input input-bordered input-sm flex-1 font-mono"
                  disabled={!canManage}
                />
              </div>
            ))}
            {deductions.length === 0 && (
              <p className="text-base-content/50 text-sm">No deduction components configured.</p>
            )}
          </div>
        </div>
      </div>

      <div className="divider" />

      <div className="stats stats-vertical md:stats-horizontal shadow border border-base-300">
        <div className="stat">
          <div className="stat-title">Gross Pay</div>
          <div className="stat-value text-success text-2xl">{fmt(totalEarnings)}</div>
        </div>
        <div className="stat">
          <div className="stat-title">Deductions</div>
          <div className="stat-value text-error text-2xl">{fmt(totalDeductions)}</div>
        </div>
        <div className="stat">
          <div className="stat-title">Net Pay</div>
          <div className="stat-value text-primary text-2xl">{fmt(totalEarnings - totalDeductions)}</div>
        </div>
      </div>
    </div>
  );
}
