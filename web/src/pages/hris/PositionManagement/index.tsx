import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import {
  createPosition,
  deletePosition,
  listDepartments,
  listPositions,
  type Position,
  updatePosition,
} from '@/lib/hris/org';

export default function PositionManagement() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<Position | null>(null);
  const [error, setError] = useState('');

  const { data: positions, isLoading } = useQuery({ queryKey: ['hris', 'positions'], queryFn: listPositions });
  const { data: departments } = useQuery({ queryKey: ['hris', 'departments'], queryFn: listDepartments });

  const createMut = useMutation({
    mutationFn: createPosition,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'positions'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Position> }) => updatePosition(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'positions'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deletePosition,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'positions'] }),
  });

  function openCreate() {
    setEditing(null);
    setError('');
    dialogRef.current?.showModal();
  }
  function openEdit(pos: Position) {
    setEditing(pos);
    setError('');
    dialogRef.current?.showModal();
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const data = {
      name: fd.get('name') as string,
      departmentId: (fd.get('departmentId') as string) || undefined,
      level: fd.get('level') as string,
    };
    if (editing) {
      updateMut.mutate({ id: editing.id, data });
    } else {
      createMut.mutate(data);
    }
  }

  const isPending = createMut.isPending || updateMut.isPending;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Positions</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm gap-2" onClick={openCreate}>
            <Plus className="h-4 w-4" />
            Add Position
          </button>
        )}
      </div>

      {isLoading ? (
        <div className="flex justify-center p-12">
          <span className="loading loading-spinner loading-lg" />
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Department</th>
                <th>Level</th>
                <th>Created</th>
                {canManage && <th className="w-24">Actions</th>}
              </tr>
            </thead>
            <tbody>
              {positions?.length === 0 && (
                <tr>
                  <td colSpan={canManage ? 5 : 4} className="text-center text-base-content/50 py-8">
                    No positions yet
                  </td>
                </tr>
              )}
              {positions?.map((p) => {
                const dept = departments?.find((d) => d.id === p.departmentId);
                return (
                  <tr key={p.id} className="hover">
                    <td className="font-medium">{p.name}</td>
                    <td className="text-sm text-base-content/60">{dept?.name ?? '-'}</td>
                    <td>
                      <span className="badge badge-sm badge-outline">{p.level}</span>
                    </td>
                    <td className="text-sm">{new Date(p.createdAt).toLocaleDateString()}</td>
                    {canManage && (
                      <td>
                        <div className="flex gap-1">
                          <button type="button" className="btn btn-ghost btn-xs" onClick={() => openEdit(p)}>
                            <Pencil className="h-3 w-3" />
                          </button>
                          <button
                            type="button"
                            className="btn btn-ghost btn-xs text-error"
                            onClick={() => deleteMut.mutate(p.id)}
                          >
                            <Trash2 className="h-3 w-3" />
                          </button>
                        </div>
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Position' : 'Add Position'}</h3>
          {error && <div className="alert alert-error mb-4 text-sm">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="form-control">
              <label className="label" htmlFor="posName">
                <span className="label-text">Name *</span>
              </label>
              <input
                id="posName"
                name="name"
                className="input input-bordered"
                required
                defaultValue={editing?.name ?? ''}
              />
            </div>
            <div className="form-control">
              <label className="label" htmlFor="posDept">
                <span className="label-text">Department</span>
              </label>
              <select
                id="posDept"
                name="departmentId"
                className="select select-bordered"
                defaultValue={editing?.departmentId ?? ''}
              >
                <option value="">-- None --</option>
                {departments?.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="form-control">
              <label className="label" htmlFor="posLevel">
                <span className="label-text">Level *</span>
              </label>
              <select
                id="posLevel"
                name="level"
                className="select select-bordered"
                required
                defaultValue={editing?.level ?? 'staff'}
              >
                <option value="staff">Staff</option>
                <option value="supervisor">Supervisor</option>
                <option value="manager">Manager</option>
                <option value="director">Director</option>
              </select>
            </div>
            <div className="modal-action">
              <button type="button" className="btn btn-ghost" onClick={() => dialogRef.current?.close()}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary" disabled={isPending}>
                {isPending && <span className="loading loading-spinner loading-sm" />}
                {editing ? 'Save' : 'Create'}
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
