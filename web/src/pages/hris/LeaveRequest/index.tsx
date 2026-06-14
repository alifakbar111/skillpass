import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { useRef, useState } from 'react';
import { cancelLeaveRequest, createLeaveRequest, listLeaveTypes, myLeaveRequests } from '@/lib/hris/leave';

export default function LeaveRequestPage() {
  const qc = useQueryClient();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [error, setError] = useState('');

  const { data: requests, isLoading } = useQuery({
    queryKey: ['hris', 'my-leave-requests'],
    queryFn: myLeaveRequests,
  });

  const { data: leaveTypes } = useQuery({
    queryKey: ['hris', 'leave-types'],
    queryFn: listLeaveTypes,
  });

  const createMut = useMutation({
    mutationFn: createLeaveRequest,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'my-leave-requests'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const cancelMut = useMutation({
    mutationFn: cancelLeaveRequest,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'my-leave-requests'] }),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const start = fd.get('startDate') as string;
    const end = fd.get('endDate') as string;
    const days = Math.ceil((new Date(end).getTime() - new Date(start).getTime()) / 86400000) + 1;
    createMut.mutate({
      leaveTypeId: fd.get('leaveTypeId') as string,
      startDate: start,
      endDate: end,
      totalDays: days,
      reason: fd.get('reason') as string,
    });
  }

  function statusBadge(status: string) {
    const map: Record<string, string> = {
      pending: 'badge-warning',
      approved: 'badge-success',
      rejected: 'badge-error',
      cancelled: 'badge-ghost',
    };
    return <span className={`badge badge-sm ${map[status] ?? 'badge-ghost'}`}>{status}</span>;
  }

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">My Leave Requests</h1>
        <button
          type="button"
          className="btn btn-primary btn-sm"
          onClick={() => {
            setError('');
            dialogRef.current?.showModal();
          }}
        >
          <Plus className="h-4 w-4" /> New Request
        </button>
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Type</th>
              <th>Start</th>
              <th>End</th>
              <th>Days</th>
              <th>Reason</th>
              <th>Status</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {requests?.map((r) => (
              <tr key={r.id}>
                <td className="font-medium">{r.leaveTypeName}</td>
                <td>{r.startDate}</td>
                <td>{r.endDate}</td>
                <td>{r.totalDays}</td>
                <td className="max-w-xs truncate">{r.reason}</td>
                <td>{statusBadge(r.status)}</td>
                <td>
                  {(r.status === 'pending' || r.status === 'approved') && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs text-error"
                      onClick={() => cancelMut.mutate(r.id)}
                    >
                      Cancel
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {requests?.length === 0 && (
              <tr>
                <td colSpan={7} className="text-center py-8 text-base-content/50">
                  No leave requests yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">New Leave Request</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <select name="leaveTypeId" className="select select-bordered w-full" required>
              <option value="">Select leave type</option>
              {leaveTypes
                ?.filter((t) => t.isActive)
                .map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name} ({t.code})
                  </option>
                ))}
            </select>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label htmlFor="lr-start" className="label">
                  <span className="label-text">Start date</span>
                </label>
                <input id="lr-start" name="startDate" type="date" className="input input-bordered w-full" required />
              </div>
              <div>
                <label htmlFor="lr-end" className="label">
                  <span className="label-text">End date</span>
                </label>
                <input id="lr-end" name="endDate" type="date" className="input input-bordered w-full" required />
              </div>
            </div>
            <textarea
              name="reason"
              placeholder="Reason for leave"
              className="textarea textarea-bordered w-full"
              rows={3}
              required
            />
            <div className="modal-action">
              <button type="button" className="btn" onClick={() => dialogRef.current?.close()}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary">
                Submit
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
