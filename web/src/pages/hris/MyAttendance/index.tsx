import { useQuery } from '@tanstack/react-query';
import { Calendar, Clock } from 'lucide-react';
import { useState } from 'react';
import { type AttendanceLog, getMyAttendance } from '@/lib/hris/attendance';

export default function MyAttendance() {
  const [month, setMonth] = useState(() => {
    const d = new Date();
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
  });

  const { data: logs, isLoading } = useQuery<AttendanceLog[]>({
    queryKey: ['hris', 'my-attendance', month],
    queryFn: () => getMyAttendance(month),
  });

  const totalPresent = logs?.filter((l) => l.attendanceCode === 'P').length ?? 0;
  const totalLate = logs?.filter((l) => l.isLate).length ?? 0;
  const totalLateMinutes = logs?.reduce((sum, l) => sum + l.lateMinutes, 0) ?? 0;

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">My Attendance</h1>
        <input
          type="month"
          value={month}
          onChange={(e) => setMonth(e.target.value)}
          className="input input-bordered input-sm"
        />
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
          <div className="stat-title">Days Present</div>
          <div className="stat-value text-2xl text-success">{totalPresent}</div>
        </div>
        <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
          <div className="stat-title">Days Late</div>
          <div className="stat-value text-2xl text-warning">{totalLate}</div>
        </div>
        <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
          <div className="stat-title">Total Late</div>
          <div className="stat-value text-2xl">{totalLateMinutes} min</div>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Date</th>
              <th>Clock In</th>
              <th>Clock Out</th>
              <th>Status</th>
              <th>Late</th>
              <th>Geofence</th>
            </tr>
          </thead>
          <tbody>
            {logs?.map((l) => (
              <tr key={l.id}>
                <td className="flex items-center gap-1">
                  <Calendar className="h-3 w-3" /> {l.date}
                </td>
                <td>{l.clockIn ? new Date(l.clockIn).toLocaleTimeString() : '—'}</td>
                <td>{l.clockOut ? new Date(l.clockOut).toLocaleTimeString() : '—'}</td>
                <td>
                  {l.isLate ? (
                    <span className="badge badge-warning badge-sm">Late</span>
                  ) : (
                    <span className="badge badge-success badge-sm">On time</span>
                  )}
                </td>
                <td>
                  {l.lateMinutes > 0 ? (
                    <span className="flex items-center gap-1 text-warning">
                      <Clock className="h-3 w-3" /> {l.lateMinutes}m
                    </span>
                  ) : (
                    '—'
                  )}
                </td>
                <td>
                  {l.isInGeofence === true && <span className="badge badge-success badge-sm">OK</span>}
                  {l.isInGeofence === false && <span className="badge badge-error badge-sm">Out</span>}
                  {l.isInGeofence == null && '—'}
                </td>
              </tr>
            ))}
            {logs?.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center py-8 text-base-content/50">
                  No attendance records this month.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
