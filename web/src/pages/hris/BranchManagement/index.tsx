import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { MapPin, Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { type Branch, createBranch, deleteBranch, listBranches, updateBranch } from '@/lib/hris/org';

export default function BranchManagement() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<Branch | null>(null);
  const [error, setError] = useState('');

  const { data: branches, isLoading } = useQuery({
    queryKey: ['hris', 'branches'],
    queryFn: listBranches,
  });

  const createMut = useMutation({
    mutationFn: createBranch,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'branches'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Branch> }) => updateBranch(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'branches'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteBranch,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'branches'] }),
  });

  function openCreate() {
    setEditing(null);
    setError('');
    dialogRef.current?.showModal();
  }

  function openEdit(branch: Branch) {
    setEditing(branch);
    setError('');
    dialogRef.current?.showModal();
  }

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const data: Partial<Branch> = {
      name: fd.get('name') as string,
      branchType: fd.get('branchType') as string,
      address: (fd.get('address') as string) || undefined,
      city: (fd.get('city') as string) || undefined,
      province: (fd.get('province') as string) || undefined,
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
        <h1 className="text-2xl font-bold">Branches</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm gap-2" onClick={openCreate}>
            <Plus className="h-4 w-4" />
            Add Branch
          </button>
        )}
      </div>

      {isLoading ? (
        <div className="flex justify-center p-12">
          <span className="loading loading-spinner loading-lg" />
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {branches?.length === 0 && (
            <p className="text-base-content/50 col-span-full text-center py-8">No branches yet</p>
          )}
          {branches?.map((b) => (
            <div key={b.id} className={`card bg-base-200 ${!b.isActive ? 'opacity-50' : ''}`}>
              <div className="card-body">
                <div className="flex items-start justify-between">
                  <div>
                    <h3 className="card-title text-base">{b.name}</h3>
                    <span className="badge badge-sm badge-outline">{b.branchType}</span>
                  </div>
                  {canManage && (
                    <div className="flex gap-1">
                      <button type="button" className="btn btn-ghost btn-xs" onClick={() => openEdit(b)}>
                        <Pencil className="h-3 w-3" />
                      </button>
                      <button
                        type="button"
                        className="btn btn-ghost btn-xs text-error"
                        onClick={() => deleteMut.mutate(b.id)}
                      >
                        <Trash2 className="h-3 w-3" />
                      </button>
                    </div>
                  )}
                </div>
                {(b.city || b.province) && (
                  <p className="text-sm text-base-content/60 flex items-center gap-1 mt-2">
                    <MapPin className="h-3 w-3" />
                    {[b.city, b.province].filter(Boolean).join(', ')}
                  </p>
                )}
                {!b.isActive && <span className="badge badge-sm badge-error mt-2">Inactive</span>}
              </div>
            </div>
          ))}
        </div>
      )}

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Branch' : 'Add Branch'}</h3>
          {error && <div className="alert alert-error mb-4 text-sm">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <div className="form-control">
              <label className="label" htmlFor="branchName">
                <span className="label-text">Name *</span>
              </label>
              <input
                id="branchName"
                name="name"
                className="input input-bordered"
                required
                defaultValue={editing?.name ?? ''}
              />
            </div>
            <div className="form-control">
              <label className="label" htmlFor="branchType">
                <span className="label-text">Type *</span>
              </label>
              <select
                id="branchType"
                name="branchType"
                className="select select-bordered"
                required
                defaultValue={editing?.branchType ?? 'headquarters'}
              >
                <option value="headquarters">Headquarters</option>
                <option value="branch">Branch</option>
                <option value="remote">Remote</option>
              </select>
            </div>
            <div className="form-control">
              <label className="label" htmlFor="branchAddr">
                <span className="label-text">Address</span>
              </label>
              <input
                id="branchAddr"
                name="address"
                className="input input-bordered"
                defaultValue={editing?.address ?? ''}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="form-control">
                <label className="label" htmlFor="branchCity">
                  <span className="label-text">City</span>
                </label>
                <input
                  id="branchCity"
                  name="city"
                  className="input input-bordered"
                  defaultValue={editing?.city ?? ''}
                />
              </div>
              <div className="form-control">
                <label className="label" htmlFor="branchProv">
                  <span className="label-text">Province</span>
                </label>
                <input
                  id="branchProv"
                  name="province"
                  className="input input-bordered"
                  defaultValue={editing?.province ?? ''}
                />
              </div>
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
