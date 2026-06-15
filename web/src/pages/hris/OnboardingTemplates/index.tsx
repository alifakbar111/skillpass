import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2 } from 'lucide-react';
import { useState } from 'react';
import { usePermissions } from '@/hooks/usePermissions';
import {
  createTemplate,
  deleteTemplate,
  listTemplates,
  type Template,
  type TemplateTask,
  updateTemplate,
} from '@/lib/hris/onboarding';

export default function OnboardingTemplates() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('org.manage');
  const [showCreate, setShowCreate] = useState(false);
  const [editId, setEditId] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isActive, setIsActive] = useState(true);
  const [tasks, setTasks] = useState<Partial<TemplateTask>[]>([]);
  const [error, setError] = useState('');

  const { data: templates, isLoading } = useQuery({
    queryKey: ['hris', 'onboarding-templates'],
    queryFn: listTemplates,
  });

  const createMut = useMutation({
    mutationFn: () => createTemplate({ name, description, tasks }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'onboarding-templates'] });
      resetForm();
    },
    onError: (e: Error) => setError(e.message),
  });

  const updateMut = useMutation({
    mutationFn: () => updateTemplate(editId ?? '', { name, description, isActive }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'onboarding-templates'] });
      resetForm();
    },
    onError: (e: Error) => setError(e.message),
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteTemplate(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'onboarding-templates'] }),
  });

  const resetForm = () => {
    setShowCreate(false);
    setEditId(null);
    setName('');
    setDescription('');
    setIsActive(true);
    setTasks([]);
    setError('');
  };

  const startEdit = (t: Template) => {
    setEditId(t.id);
    setName(t.name);
    setDescription(t.description);
    setIsActive(t.isActive);
    setShowCreate(true);
  };

  const addTask = () => {
    setTasks([...tasks, { title: '', description: '', dueDays: 0, assigneeRole: 'employee' }]);
  };

  const updateTask = (idx: number, field: string, value: string | number) => {
    const updated = [...tasks];
    updated[idx] = { ...updated[idx], [field]: value };
    setTasks(updated);
  };

  const removeTask = (idx: number) => {
    setTasks(tasks.filter((_, i) => i !== idx));
  };

  if (isLoading)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Onboarding Templates</h1>
        {canManage && (
          <button type="button" className="btn btn-primary btn-sm" onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4" /> New Template
          </button>
        )}
      </div>

      {showCreate && (
        <div className="card bg-base-100 border border-base-300 mb-6">
          <div className="card-body p-4">
            <h3 className="font-bold">{editId ? 'Edit Template' : 'New Template'}</h3>
            {error && <div className="alert alert-error text-sm py-2">{error}</div>}
            <div className="grid gap-3 md:grid-cols-2">
              <input
                className="input input-bordered input-sm"
                placeholder="Template name"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
              <input
                className="input input-bordered input-sm"
                placeholder="Description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
            {editId && (
              <label className="label cursor-pointer justify-start gap-2">
                <input
                  type="checkbox"
                  className="checkbox checkbox-sm"
                  checked={isActive}
                  onChange={(e) => setIsActive(e.target.checked)}
                />
                <span className="label-text">Active</span>
              </label>
            )}

            {!editId && (
              <>
                <div className="divider text-sm">Tasks</div>
                {tasks.map((t, i) => (
                  // biome-ignore lint/suspicious/noArrayIndexKey: form fields are not reordered
                  <div key={i} className="flex gap-2 items-center">
                    <input
                      className="input input-bordered input-xs flex-1"
                      placeholder="Task title"
                      value={t.title ?? ''}
                      onChange={(e) => updateTask(i, 'title', e.target.value)}
                    />
                    <input
                      type="number"
                      className="input input-bordered input-xs w-20"
                      placeholder="Due days"
                      value={t.dueDays ?? 0}
                      onChange={(e) => updateTask(i, 'dueDays', Number(e.target.value))}
                    />
                    <select
                      className="select select-bordered select-xs"
                      value={t.assigneeRole ?? 'employee'}
                      onChange={(e) => updateTask(i, 'assigneeRole', e.target.value)}
                    >
                      <option value="employee">Employee</option>
                      <option value="manager">Manager</option>
                      <option value="hr">HR</option>
                      <option value="it">IT</option>
                    </select>
                    <button type="button" className="btn btn-ghost btn-xs" onClick={() => removeTask(i)}>
                      <Trash2 className="h-3 w-3" />
                    </button>
                  </div>
                ))}
                <button type="button" className="btn btn-ghost btn-xs self-start" onClick={addTask}>
                  <Plus className="h-3 w-3" /> Add Task
                </button>
              </>
            )}

            <div className="flex gap-2 mt-2">
              <button
                type="button"
                className="btn btn-primary btn-sm"
                onClick={() => (editId ? updateMut.mutate() : createMut.mutate())}
              >
                {editId ? 'Update' : 'Create'}
              </button>
              <button type="button" className="btn btn-ghost btn-sm" onClick={resetForm}>
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
              <th>Name</th>
              <th>Description</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {templates?.map((t) => (
              <tr key={t.id}>
                <td className="font-medium">{t.name}</td>
                <td className="text-sm text-base-content/70">{t.description}</td>
                <td>
                  <span className={`badge badge-sm ${t.isActive ? 'badge-success' : 'badge-ghost'}`}>
                    {t.isActive ? 'Active' : 'Inactive'}
                  </span>
                </td>
                <td className="flex gap-1">
                  {canManage && (
                    <>
                      <button type="button" className="btn btn-ghost btn-xs" onClick={() => startEdit(t)}>
                        Edit
                      </button>
                      <button
                        type="button"
                        className="btn btn-ghost btn-xs text-error"
                        onClick={() => deleteMut.mutate(t.id)}
                      >
                        <Trash2 className="h-3 w-3" />
                      </button>
                    </>
                  )}
                </td>
              </tr>
            ))}
            {templates?.length === 0 && (
              <tr>
                <td colSpan={4} className="text-center py-8 text-base-content/50">
                  No onboarding templates yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
