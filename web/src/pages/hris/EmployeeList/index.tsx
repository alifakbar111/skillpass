import { useQuery } from '@tanstack/react-query';
import { ChevronLeft, ChevronRight, Plus, Search } from 'lucide-react';
import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import { listEmployees } from '@/lib/hris/employees';
import { listDepartments } from '@/lib/hris/org';

export default function EmployeeList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get('page') ?? '1');
  const [search, setSearch] = useState(searchParams.get('search') ?? '');
  const status = searchParams.get('status') ?? '';
  const departmentId = searchParams.get('departmentId') ?? '';

  const { hasPermission } = usePermissions();

  const { data, isLoading } = useQuery({
    queryKey: ['hris', 'employees', { page, search: searchParams.get('search'), status, departmentId }],
    queryFn: () =>
      listEmployees({
        page,
        pageSize: 20,
        search: searchParams.get('search') ?? undefined,
        status: status || undefined,
        departmentId: departmentId || undefined,
      }),
  });

  const { data: departments } = useQuery({
    queryKey: ['hris', 'departments'],
    queryFn: listDepartments,
  });

  function applyFilters() {
    const params = new URLSearchParams();
    if (search) params.set('search', search);
    if (status) params.set('status', status);
    if (departmentId) params.set('departmentId', departmentId);
    params.set('page', '1');
    setSearchParams(params);
  }

  function goToPage(p: number) {
    const params = new URLSearchParams(searchParams);
    params.set('page', String(p));
    setSearchParams(params);
  }

  const totalPages = data ? Math.ceil(data.total / data.pageSize) : 0;

  const statusBadge = (s: string) => {
    const map: Record<string, string> = {
      active: 'badge-success',
      resigned: 'badge-warning',
      terminated: 'badge-error',
      on_leave: 'badge-info',
    };
    return map[s] ?? 'badge-ghost';
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Employees</h1>
        {hasPermission('employee.create') && (
          <Link to="/hris/employees/new" className="btn btn-primary btn-sm gap-2">
            <Plus className="h-4 w-4" />
            Add Employee
          </Link>
        )}
      </div>

      <div className="flex flex-wrap gap-3 mb-4">
        <div className="join">
          <input
            type="text"
            placeholder="Search name or email..."
            className="input input-bordered input-sm join-item w-64"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && applyFilters()}
          />
          <button type="button" className="btn btn-sm join-item" onClick={applyFilters}>
            <Search className="h-4 w-4" />
          </button>
        </div>
        <select
          className="select select-bordered select-sm"
          value={status}
          onChange={(e) => {
            const params = new URLSearchParams(searchParams);
            params.set('status', e.target.value);
            params.set('page', '1');
            setSearchParams(params);
          }}
        >
          <option value="">All Status</option>
          <option value="active">Active</option>
          <option value="resigned">Resigned</option>
          <option value="terminated">Terminated</option>
          <option value="on_leave">On Leave</option>
        </select>
        <select
          className="select select-bordered select-sm"
          value={departmentId}
          onChange={(e) => {
            const params = new URLSearchParams(searchParams);
            params.set('departmentId', e.target.value);
            params.set('page', '1');
            setSearchParams(params);
          }}
        >
          <option value="">All Departments</option>
          {departments?.map((d) => (
            <option key={d.id} value={d.id}>
              {d.name}
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center p-12">
          <span className="loading loading-spinner loading-lg" />
        </div>
      ) : (
        <>
          <div className="overflow-x-auto">
            <table className="table table-zebra">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Department</th>
                  <th>Position</th>
                  <th>Status</th>
                  <th>Join Date</th>
                </tr>
              </thead>
              <tbody>
                {data?.employees.length === 0 && (
                  <tr>
                    <td colSpan={7} className="text-center text-base-content/50 py-8">
                      No employees found
                    </td>
                  </tr>
                )}
                {data?.employees.map((emp) => (
                  <tr key={emp.id} className="hover">
                    <td className="font-mono text-sm">{emp.employeeIdNumber}</td>
                    <td>
                      <Link to={`/hris/employees/${emp.id}`} className="link link-hover font-medium">
                        {emp.firstName} {emp.lastName}
                      </Link>
                    </td>
                    <td className="text-sm">{emp.email}</td>
                    <td className="text-sm">{emp.departmentName ?? '-'}</td>
                    <td className="text-sm">{emp.positionName ?? '-'}</td>
                    <td>
                      <span className={`badge badge-sm ${statusBadge(emp.employmentStatus)}`}>
                        {emp.employmentStatus}
                      </span>
                    </td>
                    <td className="text-sm">{new Date(emp.joinDate).toLocaleDateString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center mt-4">
              <div className="join">
                <button
                  type="button"
                  className="join-item btn btn-sm"
                  disabled={page <= 1}
                  onClick={() => goToPage(page - 1)}
                >
                  <ChevronLeft className="h-4 w-4" />
                </button>
                <span className="join-item btn btn-sm btn-disabled">
                  Page {page} of {totalPages}
                </span>
                <button
                  type="button"
                  className="join-item btn btn-sm"
                  disabled={page >= totalPages}
                  onClick={() => goToPage(page + 1)}
                >
                  <ChevronRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
