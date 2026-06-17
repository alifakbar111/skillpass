import { useMutation } from '@tanstack/react-query';
import { Download } from 'lucide-react';
import { useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { type AttendanceRow, exportAttendance } from '@/lib/hris/report';

export default function AttendanceExport() {
  const { hasPermission } = usePermissions();
  const canExport = hasPermission('analytics.export');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [rows, setRows] = useState<AttendanceRow[]>([]);

  const fetchMut = useMutation({
    mutationFn: () => exportAttendance(from, to),
    onSuccess: (data) => setRows(data ?? []),
  });

  const downloadCSV = useMutation({
    mutationFn: async () => {
      const res = await fetch(`/api/v1/hris/reports/attendance-export?from=${from}&to=${to}&format=csv`, {
        headers: { Authorization: `Bearer ${localStorage.getItem('accessToken')}` },
      });
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `attendance_${from}_${to}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    },
  });

  const fmt = (h: string) => (h ? `${h}h` : '-');

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Attendance Export</h1>

      <div className="flex flex-wrap gap-3 items-end mb-6">
        <label className="form-control">
          <span className="label label-text text-xs">From</span>
          <input
            type="date"
            className="input input-bordered input-sm"
            value={from}
            onChange={(e) => setFrom(e.target.value)}
          />
        </label>
        <label className="form-control">
          <span className="label label-text text-xs">To</span>
          <input
            type="date"
            className="input input-bordered input-sm"
            value={to}
            onChange={(e) => setTo(e.target.value)}
          />
        </label>
        <button
          type="button"
          className="btn btn-primary btn-sm"
          disabled={!from || !to || fetchMut.isPending}
          onClick={() => fetchMut.mutate()}
        >
          {fetchMut.isPending ? <span className="loading loading-spinner loading-xs" /> : 'Load'}
        </button>
        {canExport && rows.length > 0 && (
          <button type="button" className="btn btn-outline btn-sm" onClick={() => downloadCSV.mutate()}>
            <Download className="h-4 w-4" /> CSV
          </button>
        )}
      </div>

      {rows.length > 0 && (
        <div className="overflow-x-auto">
          <table className="table table-sm">
            <thead>
              <tr>
                <th>Employee</th>
                <th>Code</th>
                <th>Date</th>
                <th>Clock In</th>
                <th>Clock Out</th>
                <th>Hours</th>
                <th>Status</th>
                <th>Shift</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={`${r.employeeCode}-${r.date}`}>
                  <td>{r.employeeName}</td>
                  <td className="font-mono text-xs">{r.employeeCode}</td>
                  <td>{r.date}</td>
                  <td>{r.clockIn}</td>
                  <td>{r.clockOut || '-'}</td>
                  <td>{fmt(r.workHours)}</td>
                  <td>
                    <span
                      className={`badge badge-sm ${r.status === 'present' ? 'badge-success' : r.status === 'late' ? 'badge-warning' : 'badge-ghost'}`}
                    >
                      {r.status}
                    </span>
                  </td>
                  <td className="text-xs">{r.shiftName || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
          <p className="text-xs text-base-content/50 mt-2">{rows.length} records</p>
        </div>
      )}

      {fetchMut.isSuccess && rows.length === 0 && (
        <p className="text-center py-8 text-base-content/50">No attendance records found for this period.</p>
      )}
    </div>
  );
}
