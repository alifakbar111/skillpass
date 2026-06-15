import { useQuery } from '@tanstack/react-query';
import { FileText } from 'lucide-react';
import { useState } from 'react';
import { getMyPayslips, type Payslip } from '@/lib/hris/payroll';

export default function MyPayslips() {
  const [selected, setSelected] = useState<Payslip | null>(null);

  const { data: payslips, isLoading } = useQuery({
    queryKey: ['hris', 'my-payslips'],
    queryFn: getMyPayslips,
  });

  const fmt = (n: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(n);

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">My Payslips</h1>
      </div>

      {!selected ? (
        <div className="overflow-x-auto">
          <table className="table table-sm">
            <thead>
              <tr>
                <th>Period</th>
                <th>Gross</th>
                <th>Deductions</th>
                <th>Net Pay</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {payslips?.map((p) => (
                <tr key={p.id}>
                  <td className="font-medium">
                    {p.periodStart} - {p.periodEnd}
                  </td>
                  <td className="font-mono text-sm">{fmt(p.grossPay)}</td>
                  <td className="font-mono text-sm text-error">{fmt(p.totalDeductions)}</td>
                  <td className="font-mono text-sm font-semibold">{fmt(p.netPay)}</td>
                  <td>
                    <button type="button" className="btn btn-ghost btn-xs" onClick={() => setSelected(p)}>
                      <FileText className="h-3 w-3" /> View
                    </button>
                  </td>
                </tr>
              ))}
              {payslips?.length === 0 && (
                <tr>
                  <td colSpan={5} className="text-center py-8 text-base-content/50">
                    No payslips available yet.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      ) : (
        <div>
          <button type="button" className="btn btn-ghost btn-sm mb-4" onClick={() => setSelected(null)}>
            Back to list
          </button>

          <div className="card bg-base-100 border border-base-300 shadow-sm">
            <div className="card-body p-4">
              <div className="flex items-center justify-between mb-3">
                <h3 className="font-bold text-lg">
                  Pay Period: {selected.periodStart} - {selected.periodEnd}
                </h3>
                <button type="button" className="btn btn-ghost btn-sm print:hidden" onClick={() => window.print()}>
                  Print
                </button>
              </div>

              <div className="overflow-x-auto">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>Component</th>
                      <th>Type</th>
                      <th className="text-right">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {selected.breakdown.map((line) => (
                      <tr key={`${line.componentCode}-${line.type}`}>
                        <td>{line.componentName}</td>
                        <td>
                          <span
                            className={`badge badge-sm ${line.type === 'earning' ? 'badge-success' : 'badge-error'}`}
                          >
                            {line.type}
                          </span>
                        </td>
                        <td className="text-right font-mono text-sm">
                          {line.type === 'deduction' ? '-' : ''}
                          {fmt(line.amount)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                  <tfoot>
                    <tr>
                      <td colSpan={2} className="text-right font-medium">
                        Gross Pay
                      </td>
                      <td className="text-right font-mono font-semibold">{fmt(selected.grossPay)}</td>
                    </tr>
                    <tr>
                      <td colSpan={2} className="text-right font-medium">
                        Total Deductions
                      </td>
                      <td className="text-right font-mono text-error">{fmt(selected.totalDeductions)}</td>
                    </tr>
                    <tr className="text-lg">
                      <td colSpan={2} className="text-right font-bold">
                        Net Pay
                      </td>
                      <td className="text-right font-mono font-bold text-primary">{fmt(selected.netPay)}</td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
