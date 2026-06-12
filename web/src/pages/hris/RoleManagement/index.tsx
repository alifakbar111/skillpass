import { useQuery } from '@tanstack/react-query';
import { Shield } from 'lucide-react';
import { listRoles } from '@/lib/hris/rbac';

export default function RoleManagement() {
  const { data: roles, isLoading } = useQuery({
    queryKey: ['hris', 'roles'],
    queryFn: listRoles,
  });

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Roles</h1>

      {isLoading ? (
        <div className="flex justify-center p-12">
          <span className="loading loading-spinner loading-lg" />
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {roles?.map((role) => (
            <div key={role.id} className="card bg-base-200">
              <div className="card-body">
                <div className="flex items-center gap-2">
                  <Shield className="h-5 w-5 text-primary" />
                  <h3 className="card-title text-base">{role.name}</h3>
                </div>
                {role.description && <p className="text-sm text-base-content/60">{role.description}</p>}
                {role.isSystem && <span className="badge badge-sm badge-outline">System Role</span>}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
