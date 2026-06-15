import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, Circle, UserPlus } from 'lucide-react';
import { useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import { listEmployees } from '@/lib/hris/employees';
import {
  assignChecklist,
  completeItem,
  getChecklist,
  listChecklists,
  listTemplates,
  uncompleteItem,
} from '@/lib/hris/onboarding';

export default function OnboardingChecklists() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('employee.update');
  const [showAssign, setShowAssign] = useState(false);
  const [selectedEmployee, setSelectedEmployee] = useState('');
  const [selectedTemplate, setSelectedTemplate] = useState('');
  const [detailId, setDetailId] = useState<string | null>(null);

  const { data: checklists, isLoading } = useQuery({
    queryKey: ['hris', 'onboarding-checklists'],
    queryFn: listChecklists,
  });

  const { data: detail } = useQuery({
    queryKey: ['hris', 'onboarding-checklist', detailId],
    queryFn: () => getChecklist(detailId ?? ''),
    enabled: !!detailId,
  });

  const { data: employees } = useQuery({
    queryKey: ['hris', 'employees'],
    queryFn: () => listEmployees(),
    enabled: showAssign,
  });

  const { data: templates } = useQuery({
    queryKey: ['hris', 'onboarding-templates'],
    queryFn: listTemplates,
    enabled: showAssign,
  });

  const assignMut = useMutation({
    mutationFn: () => assignChecklist(selectedEmployee, selectedTemplate),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'onboarding-checklists'] });
      setShowAssign(false);
      setSelectedEmployee('');
      setSelectedTemplate('');
    },
  });

  const toggleItem = useMutation({
    mutationFn: ({ itemId, completed }: { itemId: string; completed: boolean }) =>
      completed ? uncompleteItem(itemId) : completeItem(itemId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'onboarding-checklist', detailId] });
      qc.invalidateQueries({ queryKey: ['hris', 'onboarding-checklists'] });
    },
  });

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  if (detailId && detail)
    return (
      <div>
        <button type="button" className="btn btn-ghost btn-sm mb-4" onClick={() => setDetailId(null)}>
          Back to list
        </button>
        <div className="flex items-center gap-3 mb-4">
          <h2 className="text-xl font-bold">{detail.employeeName}</h2>
          <span className="badge badge-sm">{detail.employeeCode}</span>
          <span className={`badge badge-sm ${detail.status === 'completed' ? 'badge-success' : 'badge-info'}`}>
            {detail.status}
          </span>
          <span className="text-sm text-base-content/60">{detail.progress}% complete</span>
        </div>
        <progress className="progress progress-primary w-full mb-4" value={detail.progress} max={100} />

        <div className="space-y-2">
          {detail.items?.map((item) => (
            <div
              key={item.id}
              className={`flex items-start gap-3 p-3 rounded-lg border ${item.isCompleted ? 'bg-success/5 border-success/20' : 'border-base-300'}`}
            >
              <button
                type="button"
                className="mt-0.5"
                onClick={() => toggleItem.mutate({ itemId: item.id, completed: item.isCompleted })}
              >
                {item.isCompleted ? (
                  <CheckCircle className="h-5 w-5 text-success" />
                ) : (
                  <Circle className="h-5 w-5 text-base-content/30" />
                )}
              </button>
              <div className="flex-1">
                <p className={`font-medium ${item.isCompleted ? 'line-through text-base-content/50' : ''}`}>
                  {item.title}
                </p>
                {item.description && <p className="text-sm text-base-content/60">{item.description}</p>}
                {item.dueDate && <p className="text-xs text-base-content/40 mt-1">Due: {item.dueDate}</p>}
              </div>
            </div>
          ))}
        </div>
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Onboarding Checklists</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm" onClick={() => setShowAssign(true)}>
            <UserPlus className="h-4 w-4" /> Assign
          </button>
        )}
      </div>

      {showAssign && (
        <div className="card bg-base-100 border border-base-300 mb-6">
          <div className="card-body p-4">
            <h3 className="font-bold">Assign Onboarding</h3>
            <div className="grid gap-3 md:grid-cols-2">
              <select
                className="select select-bordered select-sm"
                value={selectedEmployee}
                onChange={(e) => setSelectedEmployee(e.target.value)}
              >
                <option value="">Select employee</option>
                {employees?.employees?.map((emp) => (
                  <option key={emp.id} value={emp.id}>
                    {emp.firstName} {emp.lastName}
                  </option>
                ))}
              </select>
              <select
                className="select select-bordered select-sm"
                value={selectedTemplate}
                onChange={(e) => setSelectedTemplate(e.target.value)}
              >
                <option value="">Select template</option>
                {templates
                  ?.filter((t) => t.isActive)
                  .map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name}
                    </option>
                  ))}
              </select>
            </div>
            <div className="flex gap-2 mt-2">
              <button
                type="button"
                className="btn btn-primary btn-sm"
                disabled={!selectedEmployee || !selectedTemplate}
                onClick={() => assignMut.mutate()}
              >
                Assign
              </button>
              <button type="button" className="btn btn-ghost btn-sm" onClick={() => setShowAssign(false)}>
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="table table-sm">
          <thead>
            <tr>
              <th>Employee</th>
              <th>Template</th>
              <th>Progress</th>
              <th>Status</th>
              <th>Started</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {checklists?.map((cl) => (
              <tr key={cl.id}>
                <td className="font-medium">{cl.employeeName}</td>
                <td className="text-sm">{cl.templateName}</td>
                <td>
                  <div className="flex items-center gap-2">
                    <progress className="progress progress-primary w-20" value={cl.progress} max={100} />
                    <span className="text-xs">{cl.progress}%</span>
                  </div>
                </td>
                <td>
                  <span className={`badge badge-sm ${cl.status === 'completed' ? 'badge-success' : 'badge-info'}`}>
                    {cl.status}
                  </span>
                </td>
                <td className="text-xs">{new Date(cl.startedAt).toLocaleDateString()}</td>
                <td>
                  <button type="button" className="btn btn-ghost btn-xs" onClick={() => setDetailId(cl.id)}>
                    View
                  </button>
                </td>
              </tr>
            ))}
            {checklists?.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center py-8 text-base-content/50">
                  No onboarding checklists assigned yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
