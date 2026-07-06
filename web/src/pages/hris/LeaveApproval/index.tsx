import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, XCircle } from 'lucide-react';
import { useRef, useState } from 'react';
import Pagination from '@/components/ui/Pagination';
import { type LeaveRequest, listLeaveRequests, reviewLeaveRequest } from '@/lib/hris/leave';

const PAGE_SIZE = 15;

export default function LeaveApproval() {
  const qc = useQueryClient();
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [filter, setFilter] = useState('pending');
  const [reviewing, setReviewing] = useState<LeaveRequest | null>(null);
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);

  const { data: requests, isLoading } = useQuery({
    queryKey: ['hris', 'leave-requests', filter],
    queryFn: () => listLeaveRequests(filter || undefined),
  });

  const reviewMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: { status: string; comment: string } }) =>
      reviewLeaveRequest(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'leave-requests'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  function handleReview(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!reviewing) return;
    const fd = new FormData(e.currentTarget);
    reviewMut.mutate({
      id: reviewing.id,
      data: { status: fd.get('status') as string, comment: fd.get('comment') as string },
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

  // Client-side pagination
  const totalPages = requests ? Math.ceil(requests.length / PAGE_SIZE) : 0;
  const paginatedRequests = requests?.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);

  function handleFilterChange(newFilter: string) {
    setFilter(newFilter);
    setPage(1);
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
        <h1 className="text-2xl font-bold">Leave Approval</h1>
        <select
          value={filter}
          onChange={(e) => handleFilterChange(e.target.value)}
          className="select select-bordered select-sm"
        >
          <option value="">All</option>
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Employee</th>
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
            {paginatedRequests?.map((r) => (
              <tr key={r.id}>
                <td className="font-medium">{r.employeeName || r.employeeId.slice(0, 8)}</td>
                <td>
                  <span className="badge badge-ghost badge-sm">{r.leaveTypeName}</span>
                </td>
                <td>{r.startDate}</td>
                <td>{r.endDate}</td>
                <td>{r.totalDays}</td>
                <td className="max-w-xs truncate">{r.reason}</td>
                <td>{statusBadge(r.status)}</td>
                <td>
                  {r.status === 'pending' && (
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs"
                      onClick={() => {
                        setReviewing(r);
                        setError('');
                        dialogRef.current?.showModal();
                      }}
                    >
                      Review
                    </button>
                  )}
                </td>
              </tr>
            ))}
            {paginatedRequests?.length === 0 && (
              <tr>
                <td colSpan={8} className="text-center py-8 text-base-content/50">
                  No leave requests found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
        <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">Review Leave Request</h3>
          {reviewing && (
            <div className="bg-base-200 rounded-lg p-3 mb-4 text-sm space-y-1">
              <p>
                <span className="font-medium">Employee:</span> {reviewing.employeeName}
              </p>
              <p>
                <span className="font-medium">Type:</span> {reviewing.leaveTypeName}
              </p>
              <p>
                <span className="font-medium">Period:</span> {reviewing.startDate} — {reviewing.endDate} (
                {reviewing.totalDays} days)
              </p>
              <p>
                <span className="font-medium">Reason:</span> {reviewing.reason}
              </p>
            </div>
          )}
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleReview} className="space-y-3">
            <div className="flex gap-3">
              <label className="label cursor-pointer gap-2">
                <input type="radio" name="status" value="approved" className="radio radio-success radio-sm" required />
                <span className="flex items-center gap-1">
                  <CheckCircle className="h-4 w-4 text-success" /> Approve
                </span>
              </label>
              <label className="label cursor-pointer gap-2">
                <input type="radio" name="status" value="rejected" className="radio radio-error radio-sm" />
                <span className="flex items-center gap-1">
                  <XCircle className="h-4 w-4 text-error" /> Reject
                </span>
              </label>
            </div>
            <textarea
              name="comment"
              placeholder="Comment (optional)"
              className="textarea textarea-bordered w-full"
              rows={2}
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
