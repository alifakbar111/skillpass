import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Printer } from 'lucide-react';
import { useNavigate, useParams } from 'react-router-dom';
import { listPayslips, type Payslip } from '@/lib/hris/payroll';

export default function PayslipView() {
  const { runId } = useParams<{ runId: string }>();
  const navigate = useNavigate();

  const { data: payslips, isLoading } = useQuery<Payslip[]>({
    queryKey: ['hris', 'payslips', runId],
    queryFn: () => listPayslips(runId ?? ''),
    enabled: !!runId,
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
      <div className="flex items-center gap-3 mb-6">
        <button type="button" className="btn btn-ghost btn-sm" onClick={() => navigate('/hris/payroll-runs')}>
          <ArrowLeft className="h-4 w-4" />
        </button>
        <h1 className="text-2xl font-bold">
          Payslips {payslips?.[0] && `- ${payslips[0].periodStart} to ${payslips[0].periodEnd}`}
        </h1>
      </div>

      <div className="grid gap-4">
        {payslips?.map((slip) => (
          <div key={slip.id} className="card bg-base-100 border border-base-300 shadow-sm">
            <div className="card-body p-4">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <h3 className="font-bold text-lg">{slip.employeeName}</h3>
                  <span className="text-sm text-base-content/60">{slip.employeeCode}</span>
                </div>
                <button type="button" className="btn btn-ghost btn-sm print:hidden" onClick={() => window.print()}>
                  <Printer className="h-4 w-4" />
                </button>
              </div>

              <div className="overflow-x-auto">
                <table className="table table-sm">
                  <thead>
                    <tr>
                      <th>Component</th>
                      <th>Code</th>
                      <th>Type</th>
                      <th className="text-right">Amount</th>
                    </tr>
                  </thead>
                  <tbody>
                    {slip.breakdown.map((line) => (
                      <tr key={`${slip.id}-${line.componentCode}-${line.type}`}>
                        <td>{line.componentName}</td>
                        <td>
                          <span className="badge badge-ghost badge-sm">{line.componentCode}</span>
                        </td>
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
                      <td colSpan={3} className="text-right font-medium">
                        Gross Pay
                      </td>
                      <td className="text-right font-mono font-semibold">{fmt(slip.grossPay)}</td>
                    </tr>
                    <tr>
                      <td colSpan={3} className="text-right font-medium">
                        Total Deductions
                      </td>
                      <td className="text-right font-mono text-error">{fmt(slip.totalDeductions)}</td>
                    </tr>
                    <tr className="text-lg">
                      <td colSpan={3} className="text-right font-bold">
                        Net Pay
                      </td>
                      <td className="text-right font-mono font-bold text-primary">{fmt(slip.netPay)}</td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          </div>
        ))}
        {payslips?.length === 0 && (
          <div className="text-center py-12 text-base-content/50">No payslips generated for this run.</div>
        )}
      </div>
    </div>
  );
}
