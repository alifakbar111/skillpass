import { Code, Pencil, Plus, Trash2 } from 'lucide-react';
import type { Experience } from '@/lib/api-types';

interface Props {
  experiences: Experience[];
  onAdd: () => void;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

export function ProjectSection({ experiences, onAdd, onEdit, onDelete }: Props) {
  return (
    <div className="card bg-base-200 p-4">
      <div className="flex justify-between items-center mb-3">
        <h2 className="font-semibold flex items-center gap-2">
          <Code size={18} aria-hidden="true" /> Projects & Portfolio
        </h2>
        <button type="button" className="btn btn-outline btn-sm gap-1" onClick={onAdd}>
          <Plus size={16} aria-hidden="true" /> Add Project
        </button>
      </div>
      {experiences.length === 0 ? (
        <p className="text-sm opacity-60 py-4 text-center">No projects or portfolio items added yet.</p>
      ) : (
        <div className="space-y-2">
          {experiences.map((exp) => (
            <div key={exp.id} className="p-3 bg-base-100 rounded-box flex justify-between items-start">
              <div>
                <p className="font-medium">{exp.title}</p>
                {exp.description && <p className="text-sm opacity-60 mt-1">{exp.description}</p>}
                {exp.url && (
                  <a
                    href={exp.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-xs link link-primary mt-1 inline-block"
                  >
                    View project
                  </a>
                )}
              </div>
              <div className="flex gap-1">
                <button
                  type="button"
                  className="btn btn-ghost btn-xs"
                  onClick={() => exp.id && onEdit(exp.id)}
                  aria-label="Edit"
                >
                  <Pencil size={14} />
                </button>
                <button
                  type="button"
                  className="btn btn-ghost btn-xs text-error"
                  onClick={() => exp.id && onDelete(exp.id)}
                  aria-label="Delete"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
