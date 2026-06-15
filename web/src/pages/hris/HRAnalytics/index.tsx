import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { RefreshCw, TrendingDown, TrendingUp, Users } from 'lucide-react';
import { useState } from 'react';
import { generateSnapshot, getHeadcountStats, listSnapshots } from '@/lib/hris/report';

export default function HRAnalytics() {
  const qc = useQueryClient();
  const [snapMonth, setSnapMonth] = useState('');

  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['hris', 'headcount-stats'],
    queryFn: getHeadcountStats,
  });

  const { data: snapshots } = useQuery({
    queryKey: ['hris', 'snapshots'],
    queryFn: listSnapshots,
  });

  const genSnap = useMutation({
    mutationFn: () => generateSnapshot(snapMonth),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'snapshots'] });
      setSnapMonth('');
    },
  });

  if (statsLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">HR Analytics</h1>

      {stats && (
        <>
          <div className="stats stats-vertical md:stats-horizontal shadow border border-base-300 mb-6 w-full">
            <div className="stat">
              <div className="stat-figure text-primary">
                <Users className="h-6 w-6" />
              </div>
              <div className="stat-title">Total Active</div>
              <div className="stat-value text-primary">{stats.totalActive}</div>
            </div>
            <div className="stat">
              <div className="stat-title">Avg Tenure</div>
              <div className="stat-value text-secondary">{stats.avgTenureMonths.toFixed(1)}</div>
              <div className="stat-desc">months</div>
            </div>
            <div className="stat">
              <div className="stat-title">Departments</div>
              <div className="stat-value">{stats.byDepartment?.length ?? 0}</div>
            </div>
            <div className="stat">
              <div className="stat-title">Branches</div>
              <div className="stat-value">{stats.byBranch?.length ?? 0}</div>
            </div>
          </div>

          <div className="grid gap-6 md:grid-cols-2 mb-8">
            <div className="card bg-base-100 border border-base-300">
              <div className="card-body p-4">
                <h3 className="font-semibold mb-2">By Department</h3>
                {stats.byDepartment?.map((d) => (
                  <div key={d.department} className="flex justify-between items-center py-1">
                    <span className="text-sm">{d.department}</span>
                    <span className="badge badge-sm">{d.count}</span>
                  </div>
                ))}
                {(!stats.byDepartment || stats.byDepartment.length === 0) && (
                  <p className="text-sm text-base-content/50">No data</p>
                )}
              </div>
            </div>

            <div className="card bg-base-100 border border-base-300">
              <div className="card-body p-4">
                <h3 className="font-semibold mb-2">By Status</h3>
                {stats.byStatus?.map((s) => (
                  <div key={s.status} className="flex justify-between items-center py-1">
                    <span className="text-sm capitalize">{s.status}</span>
                    <span
                      className={`badge badge-sm ${s.status === 'active' ? 'badge-success' : s.status === 'terminated' ? 'badge-error' : 'badge-warning'}`}
                    >
                      {s.count}
                    </span>
                  </div>
                ))}
              </div>
            </div>

            <div className="card bg-base-100 border border-base-300">
              <div className="card-body p-4">
                <h3 className="font-semibold mb-2">By Branch</h3>
                {stats.byBranch?.map((b) => (
                  <div key={b.branch} className="flex justify-between items-center py-1">
                    <span className="text-sm">{b.branch}</span>
                    <span className="badge badge-sm">{b.count}</span>
                  </div>
                ))}
                {(!stats.byBranch || stats.byBranch.length === 0) && (
                  <p className="text-sm text-base-content/50">No data</p>
                )}
              </div>
            </div>

            <div className="card bg-base-100 border border-base-300">
              <div className="card-body p-4">
                <h3 className="font-semibold mb-2">Gender Distribution</h3>
                {stats.genderBreakdown?.map((g) => (
                  <div key={g.gender} className="flex justify-between items-center py-1">
                    <span className="text-sm">{g.gender}</span>
                    <span className="badge badge-sm">{g.count}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </>
      )}

      <div className="divider">Monthly Snapshots</div>

      <div className="flex gap-3 items-end mb-4">
        <label className="form-control">
          <span className="label label-text text-xs">Month (YYYY-MM)</span>
          <input
            type="month"
            className="input input-bordered input-sm"
            value={snapMonth}
            onChange={(e) => setSnapMonth(e.target.value)}
          />
        </label>
        <button
          type="button"
          className="btn btn-primary btn-sm"
          disabled={!snapMonth || genSnap.isPending}
          onClick={() => genSnap.mutate()}
        >
          <RefreshCw className="h-4 w-4" /> Generate
        </button>
      </div>

      {snapshots && snapshots.length > 0 && (
        <div className="overflow-x-auto">
          <table className="table table-sm">
            <thead>
              <tr>
                <th>Month</th>
                <th>Headcount</th>
                <th>New Hires</th>
                <th>Terminations</th>
                <th>Turnover %</th>
                <th>Avg Tenure</th>
              </tr>
            </thead>
            <tbody>
              {snapshots.map((s) => (
                <tr key={s.id}>
                  <td className="font-medium">{s.snapshotMonth.slice(0, 7)}</td>
                  <td>{s.totalHeadcount}</td>
                  <td className="text-success">
                    <TrendingUp className="inline h-3 w-3 mr-1" />
                    {s.newHires}
                  </td>
                  <td className="text-error">
                    <TrendingDown className="inline h-3 w-3 mr-1" />
                    {s.terminations}
                  </td>
                  <td>{s.turnoverRate.toFixed(1)}%</td>
                  <td>{s.avgTenureMonths.toFixed(1)} mo</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
