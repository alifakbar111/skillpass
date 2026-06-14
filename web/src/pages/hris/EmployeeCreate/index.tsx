import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { type CreateEmployeeRequest, createEmployee } from '@/lib/hris/employees';
import { listBranches, listDepartments, listPositions } from '@/lib/hris/org';

export default function EmployeeCreate() {
  const navigate = useNavigate();
  const qc = useQueryClient();
  const [error, setError] = useState('');

  const { data: departments } = useQuery({ queryKey: ['hris', 'departments'], queryFn: listDepartments });
  const { data: positions } = useQuery({ queryKey: ['hris', 'positions'], queryFn: listPositions });
  const { data: branches } = useQuery({ queryKey: ['hris', 'branches'], queryFn: listBranches });

  const mutation = useMutation({
    mutationFn: createEmployee,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'employees'] });
      navigate('/hris/employees');
    },
    onError: (err: Error) => setError(err.message),
  });

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError('');
    const fd = new FormData(e.currentTarget);
    const data: CreateEmployeeRequest = {
      firstName: fd.get('firstName') as string,
      lastName: (fd.get('lastName') as string) || undefined,
      email: fd.get('email') as string,
      phone: (fd.get('phone') as string) || undefined,
      employmentType: fd.get('employmentType') as string,
      joinDate: fd.get('joinDate') as string,
      departmentId: (fd.get('departmentId') as string) || undefined,
      positionId: (fd.get('positionId') as string) || undefined,
      branchId: (fd.get('branchId') as string) || undefined,
      baseSalary: fd.get('baseSalary') ? Number(fd.get('baseSalary')) : undefined,
    };
    mutation.mutate(data);
  }

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-bold mb-6">Add Employee</h1>

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="form-control">
            <label className="label" htmlFor="firstName">
              <span className="label-text">First Name *</span>
            </label>
            <input id="firstName" name="firstName" type="text" className="input input-bordered" required />
          </div>
          <div className="form-control">
            <label className="label" htmlFor="lastName">
              <span className="label-text">Last Name</span>
            </label>
            <input id="lastName" name="lastName" type="text" className="input input-bordered" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="form-control">
            <label className="label" htmlFor="email">
              <span className="label-text">Email *</span>
            </label>
            <input id="email" name="email" type="email" className="input input-bordered" required />
          </div>
          <div className="form-control">
            <label className="label" htmlFor="phone">
              <span className="label-text">Phone</span>
            </label>
            <input id="phone" name="phone" type="tel" className="input input-bordered" />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="form-control">
            <label className="label" htmlFor="employmentType">
              <span className="label-text">Employment Type *</span>
            </label>
            <select id="employmentType" name="employmentType" className="select select-bordered" required>
              <option value="permanent">Permanent</option>
              <option value="contract">Contract</option>
              <option value="probation">Probation</option>
              <option value="intern">Intern</option>
            </select>
          </div>
          <div className="form-control">
            <label className="label" htmlFor="joinDate">
              <span className="label-text">Join Date *</span>
            </label>
            <input id="joinDate" name="joinDate" type="date" className="input input-bordered" required />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="form-control">
            <label className="label" htmlFor="departmentId">
              <span className="label-text">Department</span>
            </label>
            <select id="departmentId" name="departmentId" className="select select-bordered">
              <option value="">-- None --</option>
              {departments?.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </div>
          <div className="form-control">
            <label className="label" htmlFor="positionId">
              <span className="label-text">Position</span>
            </label>
            <select id="positionId" name="positionId" className="select select-bordered">
              <option value="">-- None --</option>
              {positions?.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="form-control">
            <label className="label" htmlFor="branchId">
              <span className="label-text">Branch</span>
            </label>
            <select id="branchId" name="branchId" className="select select-bordered">
              <option value="">-- None --</option>
              {branches?.map((b) => (
                <option key={b.id} value={b.id}>
                  {b.name}
                </option>
              ))}
            </select>
          </div>
          <div className="form-control">
            <label className="label" htmlFor="baseSalary">
              <span className="label-text">Base Salary</span>
            </label>
            <input id="baseSalary" name="baseSalary" type="number" className="input input-bordered" min="0" />
          </div>
        </div>

        <div className="flex gap-2 pt-4">
          <button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
            {mutation.isPending && <span className="loading loading-spinner loading-sm" />}
            Create Employee
          </button>
          <button type="button" className="btn btn-ghost" onClick={() => navigate('/hris/employees')}>
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}
