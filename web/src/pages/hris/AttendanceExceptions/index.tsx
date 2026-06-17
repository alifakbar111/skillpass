import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertTriangle, CheckCircle, Plus, XCircle } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { type AttendanceException, createException, listExceptions, reviewException } from '@/lib/hris/attendance';

export default function AttendanceExceptions() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('attendance.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const reviewDialogRef = useRef<HTMLDialogElement>(null);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('');
  const [reviewing, setReviewing] = useState<AttendanceException | null>(null);

  const { data: exceptions, isLoading } = useQuery({
    queryKey: ['hris', 'attendance-exceptions', filter],
    queryFn: () => listExceptions(filter || undefined),
  });

  const createMut = useMutation({
    mutationFn: createException,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'attendance-exceptions'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const reviewMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { status: string; comment: string } }) => reviewException(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'attendance-exceptions'] });
      reviewDialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    createMut.mutate({
      date: fd.get('date') as string,
      exceptionType: fd.get('exceptionType') as string,
      reason: fd.get('reason') as string,
    });
  }

  function handleReview(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!reviewing) return;
    const fd = new FormData(e.currentTarget);
    reviewMut.mutate({
      id: reviewing.id,
      data: {
        status: fd.get('status') as string,
        comment: fd.get('comment') as string,
      },
    });
  }

  function statusBadge(status: string) {
    if (status === 'approved')
      return (
        <span className="badge badge-success badge-sm gap-1">
          <CheckCircle className="h-3 w-3" /> Approved
        </span>
      );
    if (status === 'rejected')
      return (
        <span className="badge badge-error badge-sm gap-1">
          <XCircle className="h-3 w-3" /> Rejected
        </span>
      );
    return (
      <span className="badge badge-warning badge-sm gap-1">
        <AlertTriangle className="h-3 w-3" /> Pending
      </span>
    );
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
        <h1 className="text-2xl font-bold">Attendance Exceptions</h1>
        <div className="flex gap-2">
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="select select-bordered select-sm"
          >
            <option value="">All</option>
            <option value="pending">Pending</option>
            <option value="approved">Approved</option>
            <option value="rejected">Rejected</option>
          </select>
          <button
            type="button"
            className="btn btn-primary btn-sm"
            onClick={() => {
              setError('');
              dialogRef.current?.showModal();
            }}
          >
            <Plus className="h-4 w-4" /> New Exception
          </button>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Employee</th>
              <th>Date</th>
              <th>Type</th>
              <th>Reason</th>
              <th>Status</th>
              {canManage && <th>Action</th>}
            </tr>
          </thead>
          <tbody>
            {exceptions?.map((ex) => (
              <tr key={ex.id}>
                <td className="font-medium">{ex.employeeName || ex.employeeId.slice(0, 8)}</td>
                <td>{ex.date}</td>
                <td>
                  <span className="badge badge-ghost badge-sm">{ex.exceptionType}</span>
                </td>
                <td className="max-w-xs truncate">{ex.reason}</td>
                <td>{statusBadge(ex.status)}</td>
                {canManage && (
                  <td>
                    {ex.status === 'pending' && (
                      <button
                        type="button"
                        className="btn btn-ghost btn-xs"
                        onClick={() => {
                          setReviewing(ex);
                          setError('');
                          reviewDialogRef.current?.showModal();
                        }}
                      >
                        Review
                      </button>
                    )}
                  </td>
                )}
              </tr>
            ))}
            {exceptions?.length === 0 && (
              <tr>
                <td colSpan={canManage ? 6 : 5} className="text-center py-8 text-base-content/50">
                  No exceptions found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">New Exception Request</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleCreate} className="space-y-3">
            <input name="date" type="date" className="input input-bordered w-full" required />
            <select name="exceptionType" className="select select-bordered w-full" required>
              <option value="">Select type</option>
              <option value="late_excuse">Late Excuse</option>
              <option value="missed_clock">Missed Clock-in/out</option>
              <option value="work_from_home">Work From Home</option>
              <option value="overtime_claim">Overtime Claim</option>
              <option value="other">Other</option>
            </select>
            <textarea
              name="reason"
              placeholder="Reason"
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

      <dialog ref={reviewDialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">Review Exception</h3>
          {reviewing && (
            <div className="bg-base-200 rounded-lg p-3 mb-4 text-sm space-y-1">
              <p>
                <span className="font-medium">Employee:</span> {reviewing.employeeName}
              </p>
              <p>
                <span className="font-medium">Date:</span> {reviewing.date}
              </p>
              <p>
                <span className="font-medium">Type:</span> {reviewing.exceptionType}
              </p>
              <p>
                <span className="font-medium">Reason:</span> {reviewing.reason}
              </p>
            </div>
          )}
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleReview} className="space-y-3">
            <select name="status" className="select select-bordered w-full" required>
              <option value="">Select decision</option>
              <option value="approved">Approve</option>
              <option value="rejected">Reject</option>
            </select>
            <textarea
              name="comment"
              placeholder="Comment (optional)"
              className="textarea textarea-bordered w-full"
              rows={2}
            />
            <div className="modal-action">
              <button type="button" className="btn" onClick={() => reviewDialogRef.current?.close()}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary">
                Submit Review
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
