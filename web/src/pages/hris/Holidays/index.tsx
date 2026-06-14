import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CalendarDays, Pencil, Plus, RefreshCw, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { createHoliday, deleteHoliday, type Holiday, listHolidays, updateHoliday } from '@/lib/hris/leave';

export default function Holidays() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<Holiday | null>(null);
  const [error, setError] = useState('');
  const [year] = useState(() => new Date().getFullYear());

  const { data: holidays, isLoading } = useQuery({
    queryKey: ['hris', 'holidays', year],
    queryFn: () => listHolidays(year),
  });

  const createMut = useMutation({
    mutationFn: createHoliday,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'holidays'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Holiday> }) => updateHoliday(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'holidays'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteHoliday,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'holidays'] }),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const data = {
      name: fd.get('name') as string,
      date: fd.get('date') as string,
      isRecurring: fd.get('isRecurring') === 'on',
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
        <h1 className="text-2xl font-bold">Holidays — {year}</h1>
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
            <Plus className="h-4 w-4" /> Add Holiday
          </button>
        )}
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Date</th>
              <th>Name</th>
              <th>Recurring</th>
              {canManage && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {holidays?.map((h) => (
              <tr key={h.id}>
                <td className="flex items-center gap-1">
                  <CalendarDays className="h-3 w-3" /> {h.date}
                </td>
                <td className="font-medium">{h.name}</td>
                <td>
                  {h.isRecurring ? (
                    <span className="flex items-center gap-1 text-success">
                      <RefreshCw className="h-3 w-3" /> Yes
                    </span>
                  ) : (
                    '—'
                  )}
                </td>
                {canManage && (
                  <td className="flex gap-1">
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs"
                      onClick={() => {
                        setEditing(h);
                        setError('');
                        dialogRef.current?.showModal();
                      }}
                    >
                      <Pencil className="h-3 w-3" />
                    </button>
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs text-error"
                      onClick={() => deleteMut.mutate(h.id)}
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </td>
                )}
              </tr>
            ))}
            {holidays?.length === 0 && (
              <tr>
                <td colSpan={canManage ? 4 : 3} className="text-center py-8 text-base-content/50">
                  No holidays configured for {year}.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Holiday' : 'New Holiday'}</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <input
              name="name"
              defaultValue={editing?.name}
              placeholder="Holiday name"
              className="input input-bordered w-full"
              required
            />
            <input
              name="date"
              type="date"
              defaultValue={editing?.date}
              className="input input-bordered w-full"
              required
            />
            <label className="label cursor-pointer justify-start gap-2">
              <input
                type="checkbox"
                name="isRecurring"
                defaultChecked={editing?.isRecurring}
                className="checkbox checkbox-sm"
              />
              <span className="label-text">Recurring every year</span>
            </label>
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
