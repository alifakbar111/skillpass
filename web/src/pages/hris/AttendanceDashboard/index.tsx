import { useQuery } from '@tanstack/react-query';
import { Clock, UserCheck, Users, UserX } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { type AttendanceLog, type DashboardResponse, getAttendanceDashboard } from '@/lib/hris/attendance';

export default function AttendanceDashboard() {
  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10));
  const [wsLogs, setWsLogs] = useState<AttendanceLog[]>([]);
  const wsRef = useRef<WebSocket | null>(null);

  const { data, isLoading } = useQuery<DashboardResponse>({
    queryKey: ['hris', 'attendance-dashboard', date],
    queryFn: () => getAttendanceDashboard(date),
  });

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const accessToken = localStorage.getItem('accessToken') ?? '';
    const ws = new WebSocket(
      `${protocol}://${window.location.hostname}:1234/api/v1/hris/attendance/ws?token=${accessToken}`,
    );
    wsRef.current = ws;

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === 'clock_in' || msg.type === 'clock_out') {
        setWsLogs((prev) => [msg.data, ...prev]);
      }
    };

    return () => ws.close();
  }, []);

  const stats = data?.stats;
  const logs = [...wsLogs, ...(data?.logs ?? [])];

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Attendance Dashboard</h1>
        <input
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="input input-bordered input-sm"
        />
      </div>

      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
            <div className="stat-figure text-primary">
              <Users className="h-6 w-6" />
            </div>
            <div className="stat-title">Total</div>
            <div className="stat-value text-2xl">{stats.totalEmployees}</div>
          </div>
          <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
            <div className="stat-figure text-success">
              <UserCheck className="h-6 w-6" />
            </div>
            <div className="stat-title">Present</div>
            <div className="stat-value text-2xl text-success">{stats.present}</div>
          </div>
          <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
            <div className="stat-figure text-warning">
              <Clock className="h-6 w-6" />
            </div>
            <div className="stat-title">Late</div>
            <div className="stat-value text-2xl text-warning">{stats.late}</div>
          </div>
          <div className="stat bg-base-100 rounded-box border border-base-300 shadow-sm">
            <div className="stat-figure text-error">
              <UserX className="h-6 w-6" />
            </div>
            <div className="stat-title">Absent</div>
            <div className="stat-value text-2xl text-error">{stats.absent}</div>
          </div>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Employee</th>
              <th>Clock In</th>
              <th>Clock Out</th>
              <th>Status</th>
              <th>Geofence</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((l) => (
              <tr key={l.id}>
                <td className="font-medium">{l.employeeName || l.employeeId.slice(0, 8)}</td>
                <td>{l.clockIn ? new Date(l.clockIn).toLocaleTimeString() : '—'}</td>
                <td>{l.clockOut ? new Date(l.clockOut).toLocaleTimeString() : '—'}</td>
                <td>
                  {l.isLate ? (
                    <span className="badge badge-warning badge-sm">Late ({l.lateMinutes}m)</span>
                  ) : (
                    <span className="badge badge-success badge-sm">On time</span>
                  )}
                </td>
                <td>
                  {l.isInGeofence === true && <span className="badge badge-success badge-sm">In range</span>}
                  {l.isInGeofence === false && <span className="badge badge-error badge-sm">Out of range</span>}
                  {l.isInGeofence == null && <span className="text-base-content/40">—</span>}
                </td>
              </tr>
            ))}
            {logs.length === 0 && (
              <tr>
                <td colSpan={5} className="text-center py-8 text-base-content/50">
                  No attendance records for this date.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
