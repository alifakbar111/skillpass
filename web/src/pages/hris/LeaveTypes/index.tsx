import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { createLeaveType, deleteLeaveType, type LeaveType, listLeaveTypes, updateLeaveType } from '@/lib/hris/leave';

export default function LeaveTypes() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<LeaveType | null>(null);
  const [error, setError] = useState('');

  const { data: types, isLoading } = useQuery({
    queryKey: ['hris', 'leave-types'],
    queryFn: listLeaveTypes,
  });

  const createMut = useMutation({
    mutationFn: createLeaveType,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'leave-types'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<LeaveType> }) => updateLeaveType(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'leave-types'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteLeaveType,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'leave-types'] }),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const data: Partial<LeaveType> = {
      name: fd.get('name') as string,
      code: fd.get('code') as string,
      defaultDaysPerYear: Number(fd.get('defaultDaysPerYear')),
      isPaid: fd.get('isPaid') === 'on',
      requiresAttachment: fd.get('requiresAttachment') === 'on',
      isActive: editing ? fd.get('isActive') === 'on' : true,
    };
    if (editing) {
      updateMut.mutate({ id: editing.id, data });
    } else {
      createMut.mutate(data);
    }
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
        <h1 className="text-2xl font-bold">Leave Types</h1>
        {canManage && (
          <button
            type="button"
            className="btn btn-primary btn-sm"
            onClick={() => {
              setEditing(null);
              setError('');
              dialogRef.current?.showModal();
            }}
          >
            <Plus className="h-4 w-4" /> Add Type
          </button>
        )}
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Name</th>
              <th>Code</th>
              <th>Days/Year</th>
              <th>Paid</th>
              <th>Attachment</th>
              <th>Status</th>
              {canManage && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {types?.map((t) => (
              <tr key={t.id}>
                <td className="font-medium">{t.name}</td>
                <td>
                  <span className="badge badge-ghost badge-sm">{t.code}</span>
                </td>
                <td>{t.defaultDaysPerYear}</td>
                <td>
                  {t.isPaid ? (
                    <span className="badge badge-success badge-sm">Yes</span>
                  ) : (
                    <span className="badge badge-ghost badge-sm">No</span>
                  )}
                </td>
                <td>{t.requiresAttachment ? 'Required' : '—'}</td>
                <td>
                  {t.isActive ? (
                    <span className="badge badge-success badge-sm">Active</span>
                  ) : (
                    <span className="badge badge-error badge-sm">Inactive</span>
                  )}
                </td>
                {canManage && (
                  <td className="flex gap-1">
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs"
                      onClick={() => {
                        setEditing(t);
                        setError('');
                        dialogRef.current?.showModal();
                      }}
                    >
                      <Pencil className="h-3 w-3" />
                    </button>
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs text-error"
                      onClick={() => deleteMut.mutate(t.id)}
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </td>
                )}
              </tr>
            ))}
            {types?.length === 0 && (
              <tr>
                <td colSpan={canManage ? 7 : 6} className="text-center py-8 text-base-content/50">
                  No leave types configured.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Leave Type' : 'New Leave Type'}</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <input
              name="name"
              defaultValue={editing?.name}
              placeholder="Leave type name"
              className="input input-bordered w-full"
              required
            />
            <input
              name="code"
              defaultValue={editing?.code}
              placeholder="Code (e.g. AL, SL)"
              className="input input-bordered w-full"
              required
            />
            <div>
              <label htmlFor="lt-days" className="label">
                <span className="label-text">Default days per year</span>
              </label>
              <input
                id="lt-days"
                name="defaultDaysPerYear"
                type="number"
                defaultValue={editing?.defaultDaysPerYear ?? 12}
                className="input input-bordered w-full"
              />
            </div>
            <label className="label cursor-pointer justify-start gap-2">
              <input
                type="checkbox"
                name="isPaid"
                defaultChecked={editing?.isPaid ?? true}
                className="checkbox checkbox-sm"
              />
              <span className="label-text">Paid leave</span>
            </label>
            <label className="label cursor-pointer justify-start gap-2">
              <input
                type="checkbox"
                name="requiresAttachment"
                defaultChecked={editing?.requiresAttachment}
                className="checkbox checkbox-sm"
              />
              <span className="label-text">Requires attachment</span>
            </label>
            {editing && (
              <label className="label cursor-pointer justify-start gap-2">
                <input
                  type="checkbox"
                  name="isActive"
                  defaultChecked={editing.isActive}
                  className="checkbox checkbox-sm"
                />
                <span className="label-text">Active</span>
              </label>
            )}
            <div className="modal-action">
              <button type="button" className="btn" onClick={() => dialogRef.current?.close()}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary">
                {editing ? 'Update' : 'Create'}
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
