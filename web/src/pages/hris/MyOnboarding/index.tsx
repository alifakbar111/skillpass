import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, Circle, PartyPopper } from 'lucide-react';
import { completeItem, getMyChecklist, uncompleteItem } from '@/lib/hris/onboarding';

export default function MyOnboarding() {
  const qc = useQueryClient();

  const { data: checklist, isLoading } = useQuery({
    queryKey: ['hris', 'my-onboarding'],
    queryFn: getMyChecklist,
  });

  const toggleItem = useMutation({
    mutationFn: ({ itemId, completed }: { itemId: string; completed: boolean }) =>
      completed ? uncompleteItem(itemId) : completeItem(itemId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'my-onboarding'] }),
  });

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  if (!checklist)
    return (
      <div className="text-center py-12">
        <p className="text-base-content/50 text-lg">No onboarding checklist assigned to you.</p>
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">My Onboarding</h1>
          <p className="text-sm text-base-content/60">Template: {checklist.templateName}</p>
        </div>
        {checklist.status === 'completed' && (
          <div className="badge badge-success gap-1">
            <PartyPopper className="h-3 w-3" /> Completed
          </div>
        )}
      </div>

      <div className="flex items-center gap-3 mb-6">
        <progress className="progress progress-primary flex-1" value={checklist.progress} max={100} />
        <span className="text-sm font-medium">{checklist.progress}%</span>
      </div>

      <div className="space-y-2">
        {checklist.items?.map((item) => (
          <div
            key={item.id}
            className={`flex items-start gap-3 p-4 rounded-lg border transition-colors ${
              item.isCompleted ? 'bg-success/5 border-success/20' : 'border-base-300 hover:bg-base-200/50'
            }`}
          >
            <button
              type="button"
              className="mt-0.5 shrink-0"
              onClick={() => toggleItem.mutate({ itemId: item.id, completed: item.isCompleted })}
            >
              {item.isCompleted ? (
                <CheckCircle className="h-5 w-5 text-success" />
              ) : (
                <Circle className="h-5 w-5 text-base-content/30 hover:text-primary" />
              )}
            </button>
            <div className="flex-1 min-w-0">
              <p className={`font-medium ${item.isCompleted ? 'line-through text-base-content/50' : ''}`}>
                {item.title}
              </p>
              {item.description && <p className="text-sm text-base-content/60 mt-1">{item.description}</p>}
              <div className="flex gap-3 mt-1">
                {item.dueDate && <span className="text-xs text-base-content/40">Due: {item.dueDate}</span>}
                {item.isCompleted && item.completedAt && (
                  <span className="text-xs text-success">Done: {new Date(item.completedAt).toLocaleDateString()}</span>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
