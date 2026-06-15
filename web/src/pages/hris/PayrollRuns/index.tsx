import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Calculator, CheckCircle, CreditCard, Eye, Plus } from 'lucide-react';
import { useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import {
  approvePayrollRun,
  calculatePayrollRun,
  createPayrollRun,
  listPayrollRuns,
  markPayrollPaid,
} from '@/lib/hris/payroll';

export default function PayrollRuns() {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const { hasPermission } = usePermissions();
  const canRun = hasPermission('payroll.run');
  const canApprove = hasPermission('payroll.approve');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [error, setError] = useState('');

  const { data: runs, isLoading } = useQuery({
    queryKey: ['hris', 'payroll-runs'],
    queryFn: listPayrollRuns,
  });

  const createMut = useMutation({
    mutationFn: createPayrollRun,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'payroll-runs'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const calcMut = useMutation({
    mutationFn: calculatePayrollRun,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'payroll-runs'] }),
    onError: (err: Error) => setError(err.message),
  });

  const approveMut = useMutation({
    mutationFn: approvePayrollRun,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'payroll-runs'] }),
    onError: (err: Error) => setError(err.message),
  });

  const paidMut = useMutation({
    mutationFn: markPayrollPaid,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'payroll-runs'] }),
    onError: (err: Error) => setError(err.message),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    createMut.mutate({
      periodStart: fd.get('periodStart') as string,
      periodEnd: fd.get('periodEnd') as string,
      notes: (fd.get('notes') as string) || undefined,
    });
  }

  const fmt = (n: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(n);

  const statusBadge = (status: string) => {
    const map: Record<string, string> = {
      draft: 'badge-ghost',
      calculated: 'badge-warning',
      approved: 'badge-success',
      paid: 'badge-info',
    };
    return <span className={`badge badge-sm ${map[status] ?? 'badge-ghost'}`}>{status}</span>;
  };

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Payroll Runs</h1>
        {canRun && (
          <button
            type="button"
            className="btn btn-primary btn-sm"
            onClick={() => {
              setError('');
              dialogRef.current?.showModal();
            }}
          >
            <Plus className="h-4 w-4" /> New Run
          </button>
        )}
      </div>

      {error && <div className="alert alert-error mb-4">{error}</div>}

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Period</th>
              <th>Status</th>
              <th>Employees</th>
              <th>Gross</th>
              <th>Deductions</th>
              <th>Net</th>
              <th>Run By</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {runs?.map((r) => (
              <tr key={r.id}>
                <td className="font-medium">
                  {r.periodStart} - {r.periodEnd}
                </td>
                <td>{statusBadge(r.status)}</td>
                <td>{r.employeeCount}</td>
                <td className="font-mono text-sm">{fmt(r.totalGross)}</td>
                <td className="font-mono text-sm">{fmt(r.totalDeductions)}</td>
                <td className="font-mono text-sm font-semibold">{fmt(r.totalNet)}</td>
                <td>{r.runByName}</td>
                <td className="flex gap-1">
                  {r.status === 'draft' && canRun && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs"
                      title="Calculate"
                      onClick={() => calcMut.mutate(r.id)}
                    >
                      <Calculator className="h-3 w-3" />
                    </button>
                  )}
                  {r.status === 'calculated' && canApprove && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs text-success"
                      title="Approve"
                      onClick={() => approveMut.mutate(r.id)}
                    >
                      <CheckCircle className="h-3 w-3" />
                    </button>
                  )}
                  {r.status === 'approved' && canApprove && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs text-info"
                      title="Mark Paid"
                      onClick={() => paidMut.mutate(r.id)}
                    >
                      <CreditCard className="h-3 w-3" />
                    </button>
                  )}
                  {r.status !== 'draft' && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs"
                      title="View Payslips"
                      onClick={() => navigate(`/hris/payroll-runs/${r.id}/payslips`)}
                    >
                      <Eye className="h-3 w-3" />
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {runs?.length === 0 && (
              <tr>
                <td colSpan={8} className="text-center py-8 text-base-content/50">
                  No payroll runs yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">New Payroll Run</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label htmlFor="pr-start" className="label">
                  <span className="label-text">Period start</span>
                </label>
                <input id="pr-start" name="periodStart" type="date" className="input input-bordered w-full" required />
              </div>
              <div>
                <label htmlFor="pr-end" className="label">
                  <span className="label-text">Period end</span>
                </label>
                <input id="pr-end" name="periodEnd" type="date" className="input input-bordered w-full" required />
              </div>
            </div>
            <textarea
              name="notes"
              placeholder="Notes (optional)"
              className="textarea textarea-bordered w-full"
              rows={2}
            />
            <div className="modal-action">
              <button type="button" className="btn" onClick={() => dialogRef.current?.close()}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary">
                Create
              </button>
            </div>
          </form>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button type="submit">close</button>
        </form>
      </dialog>
    </div>
  );
}
