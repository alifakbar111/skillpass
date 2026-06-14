import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Save } from 'lucide-react';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import { getEmployee, type UpdateEmployeeRequest, updateEmployee } from '@/lib/hris/employees';
import { listBranches, listDepartments, listPositions } from '@/lib/hris/org';

export default function EmployeeDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canEdit = hasPermission('employee.update');
  const [editing, setEditing] = useState(false);
  const [error, setError] = useState('');

  const { data: emp, isLoading } = useQuery({
    queryKey: ['hris', 'employee', id],
    queryFn: () => getEmployee(id ?? ''),
    enabled: !!id,
  });

  const { data: departments } = useQuery({ queryKey: ['hris', 'departments'], queryFn: listDepartments });
  const { data: positions } = useQuery({ queryKey: ['hris', 'positions'], queryFn: listPositions });
  const { data: branches } = useQuery({ queryKey: ['hris', 'branches'], queryFn: listBranches });

  const mutation = useMutation({
    mutationFn: (data: UpdateEmployeeRequest) => updateEmployee(id ?? '', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'employee', id] });
      qc.invalidateQueries({ queryKey: ['hris', 'employees'] });
      setEditing(false);
      setError('');
    },
    onError: (err: Error) => setError(err.message),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const fd = new FormData(e.currentTarget);
    const data: UpdateEmployeeRequest = {};
    const nullableFields = ['departmentId', 'positionId', 'branchId'];
    for (const [key, val] of fd.entries()) {
      if (nullableFields.includes(key)) {
        (data as Record<string, unknown>)[key] = val === '' ? null : val;
      } else if (val !== '' && val !== emp?.[key as keyof typeof emp]?.toString()) {
        (data as Record<string, unknown>)[key] = key === 'baseSalary' ? Number(val) : val;
      }
    }
    mutation.mutate(data);
  }

  if (isLoading) {
    return (
      <div className="flex justify-center p-12">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );
  }

  if (!emp) {
    return <div className="text-center p-12 text-base-content/50">Employee not found</div>;
  }

  return (
    <div className="max-w-3xl">
      <div className="flex items-center gap-4 mb-6">
        <button type="button" className="btn btn-ghost btn-sm" onClick={() => navigate('/hris/employees')}>
          <ArrowLeft className="h-4 w-4" />
        </button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold">
            {emp.firstName} {emp.lastName}
          </h1>
          <p className="text-sm text-base-content/60">
            {emp.employeeIdNumber} · {emp.positionName ?? 'No position'} · {emp.departmentName ?? 'No department'}
          </p>
        </div>
        {canEdit && !editing && (
          <button type="button" className="btn btn-primary btn-sm" onClick={() => setEditing(true)}>
            Edit
          </button>
        )}
      </div>

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="card bg-base-200">
            <div className="card-body">
              <h2 className="card-title text-base">Personal Info</h2>
              <Field label="First Name" name="firstName" value={emp.firstName} editing={editing} />
              <Field label="Last Name" name="lastName" value={emp.lastName} editing={editing} />
              <Field label="Email" name="email" value={emp.email} type="email" editing={editing} />
              <Field label="Phone" name="phone" value={emp.phone} type="tel" editing={editing} />
              <Field label="Date of Birth" name="dateOfBirth" value={emp.dateOfBirth} type="date" editing={editing} />
              <Field label="Gender" name="gender" value={emp.gender} editing={editing} />
              <Field label="Marital Status" name="maritalStatus" value={emp.maritalStatus} editing={editing} />
            </div>
          </div>

          <div className="card bg-base-200">
            <div className="card-body">
              <h2 className="card-title text-base">Employment</h2>
              <div className="form-control">
                <label className="label" htmlFor="empStatus">
                  <span className="label-text text-xs">Status</span>
                </label>
                {editing ? (
                  <select
                    id="empStatus"
                    name="employmentStatus"
                    className="select select-bordered select-sm"
                    defaultValue={emp.employmentStatus}
                  >
                    <option value="active">Active</option>
                    <option value="resigned">Resigned</option>
                    <option value="terminated">Terminated</option>
                    <option value="on_leave">On Leave</option>
                  </select>
                ) : (
                  <p className="text-sm">{emp.employmentStatus}</p>
                )}
              </div>
              <div className="form-control">
                <label className="label" htmlFor="empType">
                  <span className="label-text text-xs">Type</span>
                </label>
                {editing ? (
                  <select
                    id="empType"
                    name="employmentType"
                    className="select select-bordered select-sm"
                    defaultValue={emp.employmentType}
                  >
                    <option value="permanent">Permanent</option>
                    <option value="contract">Contract</option>
                    <option value="probation">Probation</option>
                    <option value="intern">Intern</option>
                  </select>
                ) : (
                  <p className="text-sm">{emp.employmentType}</p>
                )}
              </div>
              <Field label="Join Date" name="joinDate" value={emp.joinDate} type="date" editing={editing} />
              <div className="form-control">
                <label className="label" htmlFor="empDept">
                  <span className="label-text text-xs">Department</span>
                </label>
                {editing ? (
                  <select
                    id="empDept"
                    name="departmentId"
                    className="select select-bordered select-sm"
                    defaultValue={emp.departmentId ?? ''}
                  >
                    <option value="">-- None --</option>
                    {departments?.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <p className="text-sm">{emp.departmentName ?? '-'}</p>
                )}
              </div>
              <div className="form-control">
                <label className="label" htmlFor="empPos">
                  <span className="label-text text-xs">Position</span>
                </label>
                {editing ? (
                  <select
                    id="empPos"
                    name="positionId"
                    className="select select-bordered select-sm"
                    defaultValue={emp.positionId ?? ''}
                  >
                    <option value="">-- None --</option>
                    {positions?.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <p className="text-sm">{emp.positionName ?? '-'}</p>
                )}
              </div>
              <div className="form-control">
                <label className="label" htmlFor="empBranch">
                  <span className="label-text text-xs">Branch</span>
                </label>
                {editing ? (
                  <select
                    id="empBranch"
                    name="branchId"
                    className="select select-bordered select-sm"
                    defaultValue={emp.branchId ?? ''}
                  >
                    <option value="">-- None --</option>
                    {branches?.map((b) => (
                      <option key={b.id} value={b.id}>
                        {b.name}
                      </option>
                    ))}
                  </select>
                ) : (
                  <p className="text-sm">{emp.branchName ?? '-'}</p>
                )}
              </div>
              <Field
                label="Base Salary"
                name="baseSalary"
                value={emp.baseSalary?.toString()}
                type="number"
                editing={editing}
              />
            </div>
          </div>

          <div className="card bg-base-200">
            <div className="card-body">
              <h2 className="card-title text-base">Address</h2>
              <Field label="Address" name="address" value={emp.address} editing={editing} />
              <Field label="City" name="city" value={emp.city} editing={editing} />
              <Field label="Province" name="province" value={emp.province} editing={editing} />
              <Field label="Postal Code" name="postalCode" value={emp.postalCode} editing={editing} />
            </div>
          </div>

          <div className="card bg-base-200">
            <div className="card-body">
              <h2 className="card-title text-base">Emergency Contact</h2>
              <Field label="Name" name="emergencyContactName" value={emp.emergencyContactName} editing={editing} />
              <Field label="Phone" name="emergencyContactPhone" value={emp.emergencyContactPhone} editing={editing} />
              <Field
                label="Relation"
                name="emergencyContactRelation"
                value={emp.emergencyContactRelation}
                editing={editing}
              />
            </div>
          </div>
        </div>

        {editing && (
          <div className="flex gap-2 mt-6">
            <button type="submit" className="btn btn-primary btn-sm gap-2" disabled={mutation.isPending}>
              {mutation.isPending ? (
                <span className="loading loading-spinner loading-sm" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Changes
            </button>
            <button
              type="button"
              className="btn btn-ghost btn-sm"
              onClick={() => {
                setEditing(false);
                setError('');
              }}
            >
              Cancel
            </button>
          </div>
        )}
      </form>
    </div>
  );
}

function Field({
  label,
  name,
  value,
  type = 'text',
  editing,
}: {
  label: string;
  name: string;
  value?: string;
  type?: string;
  editing: boolean;
}) {
  return (
    <div className="form-control">
      <label className="label" htmlFor={name}>
        <span className="label-text text-xs">{label}</span>
      </label>
      {editing ? (
        <input id={name} name={name} type={type} className="input input-bordered input-sm" defaultValue={value ?? ''} />
      ) : (
        <p className="text-sm">{value || '-'}</p>
      )}
    </div>
  );
}
