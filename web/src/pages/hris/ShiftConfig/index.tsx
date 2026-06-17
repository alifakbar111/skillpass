import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Clock, Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import {
  createShiftTemplate,
  deleteShiftTemplate,
  listShiftTemplates,
  type ShiftTemplate,
  updateShiftTemplate,
} from '@/lib/hris/attendance';

const DAY_NAMES = ['', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

export default function ShiftConfig() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<ShiftTemplate | null>(null);
  const [error, setError] = useState('');

  const { data: shifts, isLoading } = useQuery({
    queryKey: ['hris', 'shifts'],
    queryFn: listShiftTemplates,
  });

  const createMut = useMutation({
    mutationFn: createShiftTemplate,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'shifts'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ShiftTemplate> }) => updateShiftTemplate(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'shifts'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteShiftTemplate,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'shifts'] }),
  });

  function openCreate() {
    setEditing(null);
    setError('');
    dialogRef.current?.showModal();
  }

  function openEdit(shift: ShiftTemplate) {
    setEditing(shift);
    setError('');
    dialogRef.current?.showModal();
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const days: number[] = [];
    for (let i = 1; i <= 7; i++) {
      if (fd.get(`day_${i}`)) days.push(i);
    }
    const data: Partial<ShiftTemplate> = {
      name: fd.get('name') as string,
      startTime: fd.get('startTime') as string,
      endTime: fd.get('endTime') as string,
      breakDurationMinutes: Number(fd.get('breakDurationMinutes')),
      lateToleranceMinutes: Number(fd.get('lateToleranceMinutes')),
      overtimeMultiplier: Number(fd.get('overtimeMultiplier')),
      applicableDays: days,
      isDefault: fd.get('isDefault') === 'on',
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
        <h1 className="text-2xl font-bold">Shift Configuration</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm" onClick={openCreate}>
            <Plus className="h-4 w-4" /> Add Shift
          </button>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {shifts?.map((s) => (
          <div key={s.id} className="card bg-base-100 shadow-sm border border-base-300">
            <div className="card-body">
              <div className="flex items-center justify-between">
                <h2 className="card-title text-base">
                  <Clock className="h-4 w-4" /> {s.name}
                </h2>
                {s.isDefault && <span className="badge badge-primary badge-sm">Default</span>}
              </div>
              <div className="text-sm space-y-1 mt-2">
                <p>
                  <span className="text-base-content/60">Time:</span> {s.startTime} — {s.endTime}
                </p>
                <p>
                  <span className="text-base-content/60">Break:</span> {s.breakDurationMinutes} min
                </p>
                <p>
                  <span className="text-base-content/60">Late tolerance:</span> {s.lateToleranceMinutes} min
                </p>
                <p>
                  <span className="text-base-content/60">OT multiplier:</span> {s.overtimeMultiplier}x
                </p>
                <p>
                  <span className="text-base-content/60">Days:</span>{' '}
                  {s.applicableDays?.map((d) => DAY_NAMES[d]).join(', ')}
                </p>
              </div>
              {canManage && (
                <div className="card-actions justify-end mt-2">
                  <button type="button" className="btn btn-ghost btn-xs" onClick={() => openEdit(s)}>
                    <Pencil className="h-3 w-3" />
                  </button>
                  <button
                    type="button"
                    className="btn btn-ghost btn-xs text-error"
                    onClick={() => deleteMut.mutate(s.id)}
                  >
                    <Trash2 className="h-3 w-3" />
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}
        {shifts?.length === 0 && (
          <div className="col-span-full text-center py-12 text-base-content/50">No shift templates configured yet.</div>
        )}
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Shift' : 'New Shift Template'}</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <input
              name="name"
              defaultValue={editing?.name}
              placeholder="Shift name"
              className="input input-bordered w-full"
              required
            />
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label htmlFor="shift-start" className="label">
                  <span className="label-text">Start time</span>
                </label>
                <input
                  id="shift-start"
                  name="startTime"
                  type="time"
                  defaultValue={editing?.startTime?.slice(0, 5)}
                  className="input input-bordered w-full"
                  required
                />
              </div>
              <div>
                <label htmlFor="shift-end" className="label">
                  <span className="label-text">End time</span>
                </label>
                <input
                  id="shift-end"
                  name="endTime"
                  type="time"
                  defaultValue={editing?.endTime?.slice(0, 5)}
                  className="input input-bordered w-full"
                  required
                />
              </div>
            </div>
            <div className="grid grid-cols-3 gap-3">
              <div>
                <label htmlFor="shift-break" className="label">
                  <span className="label-text">Break (min)</span>
                </label>
                <input
                  id="shift-break"
                  name="breakDurationMinutes"
                  type="number"
                  defaultValue={editing?.breakDurationMinutes ?? 60}
                  className="input input-bordered w-full"
                />
              </div>
              <div>
                <label htmlFor="shift-late" className="label">
                  <span className="label-text">Late tol. (min)</span>
                </label>
                <input
                  id="shift-late"
                  name="lateToleranceMinutes"
                  type="number"
                  defaultValue={editing?.lateToleranceMinutes ?? 15}
                  className="input input-bordered w-full"
                />
              </div>
              <div>
                <label htmlFor="shift-ot" className="label">
                  <span className="label-text">OT multiplier</span>
                </label>
                <input
                  id="shift-ot"
                  name="overtimeMultiplier"
                  type="number"
                  step="0.01"
                  defaultValue={editing?.overtimeMultiplier ?? 1.5}
                  className="input input-bordered w-full"
                />
              </div>
            </div>
            <div>
              <span id="shift-days-label" className="label">
                <span className="label-text">Working days</span>
              </span>
              <div className="flex gap-2 flex-wrap">
                {[1, 2, 3, 4, 5, 6, 7].map((d) => (
                  <label key={d} className="label cursor-pointer gap-1">
                    <input
                      type="checkbox"
                      name={`day_${d}`}
                      defaultChecked={editing?.applicableDays?.includes(d) ?? d <= 5}
                      className="checkbox checkbox-sm"
                    />
                    <span className="label-text text-sm">{DAY_NAMES[d]}</span>
                  </label>
                ))}
              </div>
            </div>
            <label className="label cursor-pointer justify-start gap-2">
              <input
                type="checkbox"
                name="isDefault"
                defaultChecked={editing?.isDefault}
                className="checkbox checkbox-sm"
              />
              <span className="label-text">Set as default shift</span>
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
