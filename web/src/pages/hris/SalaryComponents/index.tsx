import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import {
  createSalaryComponent,
  deleteSalaryComponent,
  listSalaryComponents,
  type SalaryComponent,
  updateSalaryComponent,
} from '@/lib/hris/payroll';

export default function SalaryComponents() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('payroll.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<SalaryComponent | null>(null);
  const [error, setError] = useState('');

  const { data: components, isLoading } = useQuery({
    queryKey: ['hris', 'salary-components'],
    queryFn: listSalaryComponents,
  });

  const createMut = useMutation({
    mutationFn: createSalaryComponent,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'salary-components'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const updateMut = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<SalaryComponent> }) => updateSalaryComponent(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'salary-components'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMut = useMutation({
    mutationFn: deleteSalaryComponent,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'salary-components'] }),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const data: Partial<SalaryComponent> = {
      name: fd.get('name') as string,
      code: fd.get('code') as string,
      type: fd.get('type') as 'earning' | 'deduction',
      isTaxable: fd.get('isTaxable') === 'on',
      isFixed: fd.get('isFixed') === 'on',
      defaultAmount: Number(fd.get('defaultAmount')),
      isActive: editing ? fd.get('isActive') === 'on' : true,
    };
    if (editing) {
      updateMut.mutate({ id: editing.id, data });
    } else {
      createMut.mutate(data);
    }
  }

  const fmt = (n: number) =>
    new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(n);

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Salary Components</h1>
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
            <Plus className="h-4 w-4" /> Add Component
          </button>
        )}
      </div>

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Name</th>
              <th>Code</th>
              <th>Type</th>
              <th>Default Amount</th>
              <th>Taxable</th>
              <th>Fixed</th>
              <th>Status</th>
              {canManage && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {components?.map((c) => (
              <tr key={c.id}>
                <td className="font-medium">{c.name}</td>
                <td>
                  <span className="badge badge-ghost badge-sm">{c.code}</span>
                </td>
                <td>
                  <span className={`badge badge-sm ${c.type === 'earning' ? 'badge-success' : 'badge-error'}`}>
                    {c.type}
                  </span>
                </td>
                <td className="font-mono text-sm">{fmt(c.defaultAmount)}</td>
                <td>{c.isTaxable ? 'Yes' : 'No'}</td>
                <td>{c.isFixed ? 'Fixed' : 'Variable'}</td>
                <td>
                  <span className={`badge badge-sm ${c.isActive ? 'badge-success' : 'badge-error'}`}>
                    {c.isActive ? 'Active' : 'Inactive'}
                  </span>
                </td>
                {canManage && (
                  <td className="flex gap-1">
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs"
                      onClick={() => {
                        setEditing(c);
                        setError('');
                        dialogRef.current?.showModal();
                      }}
                    >
                      <Pencil className="h-3 w-3" />
                    </button>
                    <button
                      type="button"
                      className="btn btn-ghost btn-xs text-error"
                      onClick={() => deleteMut.mutate(c.id)}
                    >
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </td>
                )}
              </tr>
            ))}
            {components?.length === 0 && (
              <tr>
                <td colSpan={canManage ? 8 : 7} className="text-center py-8 text-base-content/50">
                  No salary components configured.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="font-bold text-lg mb-4">{editing ? 'Edit Component' : 'New Salary Component'}</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <input
              name="name"
              defaultValue={editing?.name}
              placeholder="Component name"
              className="input input-bordered w-full"
              required
            />
            <input
              name="code"
              defaultValue={editing?.code}
              placeholder="Code (e.g. BASIC, HRA)"
              className="input input-bordered w-full"
              required
            />
            <div>
              <label htmlFor="sc-type" className="label">
                <span className="label-text">Type</span>
              </label>
              <select
                id="sc-type"
                name="type"
                defaultValue={editing?.type ?? 'earning'}
                className="select select-bordered w-full"
                required
              >
                <option value="earning">Earning</option>
                <option value="deduction">Deduction</option>
              </select>
            </div>
            <div>
              <label htmlFor="sc-amount" className="label">
                <span className="label-text">Default amount</span>
              </label>
              <input
                id="sc-amount"
                name="defaultAmount"
                type="number"
                step="0.01"
                defaultValue={editing?.defaultAmount ?? 0}
                className="input input-bordered w-full"
              />
            </div>
            <label className="label cursor-pointer justify-start gap-2">
              <input
                type="checkbox"
                name="isTaxable"
                defaultChecked={editing?.isTaxable ?? true}
                className="checkbox checkbox-sm"
              />
              <span className="label-text">Taxable</span>
            </label>
            <label className="label cursor-pointer justify-start gap-2">
              <input
                type="checkbox"
                name="isFixed"
                defaultChecked={editing?.isFixed ?? true}
                className="checkbox checkbox-sm"
              />
              <span className="label-text">Fixed amount</span>
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
