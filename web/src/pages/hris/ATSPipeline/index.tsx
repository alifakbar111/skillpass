import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FileText, Plus, UserPlus } from 'lucide-react';
import { useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import { type AtsCandidate, addCandidate, listCandidates, listPipelines, moveCandidate } from '@/lib/hris/ats';

export default function ATSPipeline() {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('ats.manage');
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [error, setError] = useState('');
  const [pipelineId, setPipelineId] = useState<string>('');
  const [dragId, setDragId] = useState<string | null>(null);

  const { data: pipelines, isLoading: pipesLoading } = useQuery({
    queryKey: ['hris', 'ats', 'pipelines'],
    queryFn: listPipelines,
  });

  const activePipeline = pipelines?.find((p) => p.id === pipelineId) ?? pipelines?.[0];

  const { data: candidates, isLoading: candsLoading } = useQuery({
    queryKey: ['hris', 'ats', 'candidates', activePipeline?.id],
    queryFn: () => listCandidates(activePipeline?.id),
    enabled: !!activePipeline,
  });

  const addMut = useMutation({
    mutationFn: (body: { candidateName: string; candidateEmail: string; pipelineId?: string }) => addCandidate(body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'ats', 'candidates'] });
      dialogRef.current?.close();
    },
    onError: (err: Error) => setError(err.message),
  });

  const moveMut = useMutation({
    mutationFn: ({ id, stageId }: { id: string; stageId: string }) => moveCandidate(id, stageId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'ats', 'candidates'] }),
  });

  function handleAdd(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    addMut.mutate({
      candidateName: fd.get('candidateName') as string,
      candidateEmail: fd.get('candidateEmail') as string,
      pipelineId: activePipeline?.id,
    });
  }

  function onDrop(stageId: string) {
    if (dragId) moveMut.mutate({ id: dragId, stageId });
    setDragId(null);
  }

  if (pipesLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  const stages = activePipeline?.stages ?? [];
  const byStage = (stageId: string) =>
    (candidates ?? []).filter((c) => c.currentStageId === stageId && c.status === 'active');

  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3 mb-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Hiring Pipeline</h1>
          <p className="text-sm text-base-content/60">Drag candidates across stages to move them through hiring.</p>
        </div>
        <div className="flex items-center gap-2">
          {(pipelines?.length ?? 0) > 1 && (
            <select
              className="select select-bordered select-sm"
              value={activePipeline?.id}
              onChange={(e) => setPipelineId(e.target.value)}
            >
              {pipelines?.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          )}
          <Link to="/hris/ats/offer-templates" className="btn btn-ghost btn-sm">
            <FileText className="h-4 w-4" /> Offer Templates
          </Link>
          {canManage && (
            <button
              type="button"
              className="btn btn-primary btn-sm"
              onClick={() => {
                setError('');
                dialogRef.current?.showModal();
              }}
            >
              <UserPlus className="h-4 w-4" /> Add Candidate
            </button>
          )}
        </div>
      </div>

      {candsLoading ? (
        <div className="flex justify-center p-8">
          <span className="loading loading-spinner loading-lg" />
        </div>
      ) : (
        <div className="flex gap-4 overflow-x-auto pb-4">
          {stages.map((stage) => {
            const cards = byStage(stage.id);
            return (
              // biome-ignore lint/a11y/noStaticElementInteractions: drop zone for the drag-and-drop kanban; keyboard users move candidates via the detail page
              <div
                key={stage.id}
                aria-label={`${stage.name} stage`}
                className="flex w-72 shrink-0 flex-col rounded-xl bg-base-200/50 p-3"
                onDragOver={(e) => e.preventDefault()}
                onDrop={() => onDrop(stage.id)}
              >
                <div className="mb-2 flex items-center justify-between px-1">
                  <span className="text-sm font-semibold">{stage.name}</span>
                  <span className="badge badge-ghost badge-sm">{cards.length}</span>
                </div>
                <div className="flex flex-col gap-2">
                  {cards.map((c) => (
                    <CandidateCard
                      key={c.id}
                      candidate={c}
                      draggable={canManage}
                      onDragStart={() => setDragId(c.id)}
                      onClick={() => navigate(`/hris/ats/candidates/${c.id}`)}
                    />
                  ))}
                  {cards.length === 0 && (
                    <p className="px-1 py-4 text-center text-xs text-base-content/40">No candidates</p>
                  )}
                </div>
              </div>
            );
          })}
          {stages.length === 0 && (
            <div className="rounded-xl border border-dashed border-base-300 p-8 text-center text-base-content/50">
              <Plus className="mx-auto mb-2 h-6 w-6" />
              This pipeline has no stages yet.
            </div>
          )}
        </div>
      )}

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box">
          <h3 className="mb-4 text-lg font-bold">Add Candidate</h3>
          {error && <div className="alert alert-error mb-4">{error}</div>}
          <form onSubmit={handleAdd} className="space-y-3">
            <input name="candidateName" placeholder="Full name" className="input input-bordered w-full" required />
            <input
              name="candidateEmail"
              type="email"
              placeholder="Email address"
              className="input input-bordered w-full"
              required
            />
            <p className="text-xs text-base-content/50">
              If this email matches an existing SkillPass account, hiring them later links their login automatically.
            </p>
            <div className="modal-action">
              <button type="button" className="btn" onClick={() => dialogRef.current?.close()}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary" disabled={addMut.isPending}>
                Add
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

function CandidateCard({
  candidate,
  draggable,
  onDragStart,
  onClick,
}: {
  candidate: AtsCandidate;
  draggable: boolean;
  onDragStart: () => void;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      draggable={draggable}
      onDragStart={onDragStart}
      onClick={onClick}
      className="cursor-pointer rounded-lg border border-base-300 bg-base-100 p-3 text-left shadow-sm transition-shadow hover:shadow-md"
    >
      <p className="text-sm font-medium">{candidate.candidateName}</p>
      <p className="truncate text-xs text-base-content/50">{candidate.candidateEmail}</p>
      {candidate.jobTitle && <span className="badge badge-ghost badge-sm mt-1">{candidate.jobTitle}</span>}
    </button>
  );
}
