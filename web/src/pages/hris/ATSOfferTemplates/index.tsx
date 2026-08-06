import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Pencil, Plus, Trash2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import {
  type AtsOfferTemplate,
  createOfferTemplate,
  deleteOfferTemplate,
  listOfferTemplates,
  updateOfferTemplate,
} from '@/lib/hris/ats';

const MERGE_HINT = 'Merge fields: {{candidateName}}, {{position}}, {{salary}}, {{startDate}}';

export default function ATSOfferTemplates() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('ats.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [editing, setEditing] = useState<AtsOfferTemplate | null>(null);
  const [error, setError] = useState('');

  const { data: templates, isLoading } = useQuery({
    queryKey: ['hris', 'ats', 'offer-templates'],
    queryFn: listOfferTemplates,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: ['hris', 'ats', 'offer-templates'] });
  const createMut = useMutation({
    mutationFn: createOfferTemplate,
    onSuccess: () => {
      invalidate();
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });
  const updateMut = useMutation({
    mutationFn: ({ id, body }: { id: string; body: { name: string; body: string } }) => updateOfferTemplate(id, body),
    onSuccess: () => {
      invalidate();
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });
  const deleteMut = useMutation({ mutationFn: deleteOfferTemplate, onSuccess: invalidate });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const body = { name: fd.get('name') as string, body: fd.get('body') as string };
    if (editing) updateMut.mutate({ id: editing.id, body });
    else createMut.mutate(body);
  }

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div className="max-w-3xl">
      <Link to="/hris/ats" className="btn btn-ghost btn-sm mb-4">
        <ArrowLeft className="h-4 w-4" /> Back to pipeline
      </Link>
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold tracking-tight">Offer Templates</h1>
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
            <Plus className="h-4 w-4" /> New Template
          </button>
        )}
      </div>

      <div className="space-y-3">
        {templates?.map((t) => (
          <div key={t.id} className="rounded-xl border border-base-300 bg-base-100 p-4">
            <div className="flex items-center justify-between">
              <span className="font-medium">{t.name}</span>
              {canManage && (
                <div className="flex gap-1">
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
                </div>
              )}
            </div>
            <p className="mt-2 line-clamp-3 whitespace-pre-wrap text-sm text-base-content/60">{t.body}</p>
          </div>
        ))}
        {templates?.length === 0 && (
          <p className="rounded-xl border border-dashed border-base-300 p-8 text-center text-base-content/50">
            No offer templates yet.
          </p>
        )}
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box max-w-2xl">
          <h3 className="mb-4 text-lg font-bold">{editing ? 'Edit Template' : 'New Template'}</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleSubmit} className="space-y-3">
            <input
              name="name"
              defaultValue={editing?.name}
              placeholder="Template name"
              className="input input-bordered w-full"
              required
            />
            <textarea
              name="body"
              defaultValue={editing?.body}
              placeholder="Dear {{candidateName}}, we are delighted to offer you the position of {{position}}..."
              className="textarea textarea-bordered w-full font-mono text-sm"
              rows={10}
              required
            />
            <p className="text-xs text-base-content/50">{MERGE_HINT}</p>
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
