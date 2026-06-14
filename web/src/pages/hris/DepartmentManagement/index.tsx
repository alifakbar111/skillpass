import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { createDepartment, type Department, deleteDepartment, listDepartments, updateDepartment } from '@/lib/hris/org';

export default function DepartmentManagement() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<Department | null>(null);
  const [error, setError] = useState('');

  const { data: departments, isLoading } = useQuery({
    queryKey: ['hris', 'departments'],
    queryFn: listDepartments,
  });

  const createMut = useMutation({
    mutationFn: createDepartment,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'departments'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Department> }) => updateDepartment(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'departments'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteDepartment,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'departments'] }),
  });

  function openCreate() {
    setEditing(null);
    setError('');
    dialogRef.current?.showModal();
  }

  function openEdit(dept: Department) {
    setEditing(dept);
    setError('');
    dialogRef.current?.showModal();
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const name = fd.get('name') as string;
    const parentDepartmentId = (fd.get('parentDepartmentId') as string) || undefined;
    if (editing) {
      updateMut.mutate({ id: editing.id, data: { name, parentDepartmentId } });
    } else {
      createMut.mutate({ name, parentDepartmentId });
    }
  }

  const isPending = createMut.isPending || updateMut.isPending;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Departments</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm gap-2" onClick={openCreate}>
            <Plus className="h-4 w-4" />
            Add Department
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
                <th>Parent</th>
                <th>Created</th>
                {canManage && <th className="w-24">Actions</th>}
              </tr>
            </thead>
            <tbody>
              {departments?.length === 0 && (
                <tr>
                  <td colSpan={canManage ? 4 : 3} className="text-center text-base-content/50 py-8">
                    No departments yet
                  </td>
                </tr>
              )}
              {departments?.map((d) => {
                const parent = departments.find((p) => p.id === d.parentDepartmentId);
                return (
                  <tr key={d.id} className="hover">
                    <td className="font-medium">{d.name}</td>
                    <td className="text-sm text-base-content/60">{parent?.name ?? '-'}</td>
                    <td className="text-sm">{new Date(d.createdAt).toLocaleDateString()}</td>
                    {canManage && (
                      <td>
                        <div className="flex gap-1">
                          <button type="button" className="btn btn-ghost btn-xs" onClick={() => openEdit(d)}>
                            <Pencil className="h-3 w-3" />
                          </button>
                          <button
                            type="button"
                            className="btn btn-ghost btn-xs text-error"
                            onClick={() => deleteMut.mutate(d.id)}
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
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Department' : 'Add Department'}</h3>
          {error && <div className="alert alert-error mb-4 text-sm">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="form-control">
              <label className="label" htmlFor="deptName">
                <span className="label-text">Name *</span>
              </label>
              <input
                id="deptName"
                name="name"
                className="input input-bordered"
                required
                defaultValue={editing?.name ?? ''}
              />
            </div>
            <div className="form-control">
              <label className="label" htmlFor="parentDept">
                <span className="label-text">Parent Department</span>
              </label>
              <select
                id="parentDept"
                name="parentDepartmentId"
                className="select select-bordered"
                defaultValue={editing?.parentDepartmentId ?? ''}
              >
                <option value="">-- None (Top Level) --</option>
                {departments
                  ?.filter((d) => d.id !== editing?.id)
                  .map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
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
