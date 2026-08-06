import { useQuery } from '@tanstack/react-query';
import {
  CalendarPlus,
  ChevronLeft,
  ChevronRight,
  Download,
  Plus,
  Search,
  SlidersHorizontal,
  UserCheck,
  Users,
  UsersRound,
} from 'lucide-react';
import { Fragment, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { EmptyState, PageHeader, StatCard } from '@/components/hris/ui';
import { usePermissions } from '@/hooks/usePermissions';
import { type Employee, listEmployees } from '@/lib/hris/employees';
import { listDepartments } from '@/lib/hris/org';
import { getHeadcountStats } from '@/lib/hris/report';

const PAGE_SIZE = 20;

type ColKey = 'employeeId' | 'department' | 'position' | 'branch' | 'employmentType' | 'status' | 'joinDate';

const OPTIONAL_COLUMNS: { key: ColKey; label: string }[] = [
  { key: 'employeeId', label: 'Employee ID' },
  { key: 'department', label: 'Department' },
  { key: 'position', label: 'Position' },
  { key: 'branch', label: 'Branch' },
  { key: 'employmentType', label: 'Employment type' },
  { key: 'status', label: 'Status' },
  { key: 'joinDate', label: 'Join date' },
];

const statusBadge = (s: string) => {
  const map: Record<string, string> = {
    active: 'badge-success',
    resigned: 'badge-warning',
    terminated: 'badge-error',
    on_leave: 'badge-info',
  };
  return map[s] ?? 'badge-ghost';
};

function csvCell(v: unknown): string {
  const s = v == null ? '' : String(v);
  return `"${s.replace(/"/g, '""')}"`;
}

export default function EmployeeList() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get('page') ?? '1');
  const [search, setSearch] = useState(searchParams.get('search') ?? '');
  const status = searchParams.get('status') ?? '';
  const departmentId = searchParams.get('departmentId') ?? '';

  const [visibleCols, setVisibleCols] = useState<Set<ColKey>>(
    new Set(['employeeId', 'department', 'position', 'status', 'joinDate']),
  );
  const [exporting, setExporting] = useState(false);

  const { hasPermission } = usePermissions();

  const { data, isLoading } = useQuery({
    queryKey: ['hris', 'employees', { page, search: searchParams.get('search'), status, departmentId }],
    queryFn: () =>
      listEmployees({
        page,
        pageSize: PAGE_SIZE,
        search: searchParams.get('search') ?? undefined,
        status: status || undefined,
        departmentId: departmentId || undefined,
      }),
  });

  const { data: departments } = useQuery({
    queryKey: ['hris', 'departments'],
    queryFn: listDepartments,
  });

  const canViewAnalytics = hasPermission('analytics.view') || hasPermission('analytics.view_team');
  const { data: stats } = useQuery({
    queryKey: ['hris', 'headcount'],
    queryFn: getHeadcountStats,
    enabled: canViewAnalytics,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  const onLeaveCount = stats?.byStatus.find((s) => s.status === 'on_leave')?.count ?? 0;

  function applyFilters() {
    const params = new URLSearchParams(searchParams);
    if (search) params.set('search', search);
    else params.delete('search');
    params.set('page', '1');
    setSearchParams(params);
  }

  function setParam(key: string, value: string) {
    const params = new URLSearchParams(searchParams);
    if (value) params.set(key, value);
    else params.delete(key);
    params.set('page', '1');
    setSearchParams(params);
  }

  function goToPage(p: number) {
    const params = new URLSearchParams(searchParams);
    params.set('page', String(p));
    setSearchParams(params);
  }

  function toggleCol(key: ColKey) {
    setVisibleCols((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }

  async function exportCsv() {
    setExporting(true);
    try {
      // The server clamps pageSize to 100 (M-6/F3), so a single request with
      // pageSize 1000 silently truncated the CSV. Loop pages instead.
      const perPage = 100;
      const filters = {
        pageSize: perPage,
        search: searchParams.get('search') ?? undefined,
        status: status || undefined,
        departmentId: departmentId || undefined,
      };
      const rows: Employee[] = [];
      let pageNum = 1;
      for (;;) {
        const res = await listEmployees({ page: pageNum, ...filters });
        rows.push(...res.employees);
        if (res.employees.length < perPage || rows.length >= res.total) break;
        pageNum++;
      }
      const header = [
        'Employee ID',
        'First name',
        'Last name',
        'Email',
        'Department',
        'Position',
        'Branch',
        'Employment type',
        'Status',
        'Join date',
      ];
      const lines = rows.map((e) =>
        [
          e.employeeIdNumber,
          e.firstName,
          e.lastName,
          e.email,
          e.departmentName ?? '',
          e.positionName ?? '',
          e.branchName ?? '',
          e.employmentType,
          e.employmentStatus,
          e.joinDate,
        ]
          .map(csvCell)
          .join(','),
      );
      const csv = [header.map(csvCell).join(','), ...lines].join('\r\n');
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `employees-${new Date().toISOString().slice(0, 10)}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } finally {
      setExporting(false);
    }
  }

  const total = data?.total ?? 0;
  const totalPages = data ? Math.ceil(total / data.pageSize) : 0;
  const rangeStart = total === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
  const rangeEnd = Math.min(page * PAGE_SIZE, total);

  function renderCell(emp: Employee, key: ColKey) {
    switch (key) {
      case 'employeeId':
        return <td className="font-mono text-sm text-base-content/70">{emp.employeeIdNumber}</td>;
      case 'department':
        return <td className="text-sm">{emp.departmentName ?? '-'}</td>;
      case 'position':
        return <td className="text-sm">{emp.positionName ?? '-'}</td>;
      case 'branch':
        return <td className="text-sm">{emp.branchName ?? '-'}</td>;
      case 'employmentType':
        return <td className="text-sm capitalize">{emp.employmentType?.replace('_', ' ') ?? '-'}</td>;
      case 'status':
        return (
          <td>
            <span className={`badge badge-sm ${statusBadge(emp.employmentStatus)}`}>
              {emp.employmentStatus.replace('_', ' ')}
            </span>
          </td>
        );
      case 'joinDate':
        return <td className="text-sm">{new Date(emp.joinDate).toLocaleDateString()}</td>;
      default:
        return <td />;
    }
  }

  const shownCols = OPTIONAL_COLUMNS.filter((c) => visibleCols.has(c.key));

  return (
    <div>
      <PageHeader
        title="Employees"
        subtitle="Manage your organisation's people and employment records."
        actions={
          hasPermission('employee.create') && (
            <Link to="/hris/employees/new" className="btn btn-primary btn-sm gap-2">
              <Plus className="h-4 w-4" />
              Add Employee
            </Link>
          )
        }
      />

      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Total employees" value={total} icon={Users} tone="primary" hint="Across all statuses" />
        <StatCard
          label="Active employees"
          value={stats?.totalActive ?? (canViewAnalytics ? '—' : total)}
          icon={UserCheck}
          tone="success"
          hint={canViewAnalytics ? 'Currently employed' : undefined}
        />
        <StatCard
          label="On leave"
          value={canViewAnalytics ? onLeaveCount : '—'}
          icon={CalendarPlus}
          tone="warning"
          hint="This period"
        />
      </div>

      <div className="rounded-box border border-base-300 bg-base-100">
        <div className="flex flex-wrap items-center gap-3 border-b border-base-300 p-4">
          <div className="relative min-w-0 flex-1 sm:max-w-xs">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-base-content/40" />
            <input
              type="search"
              placeholder="Search name or email…"
              className="input input-sm w-full rounded-lg border-base-300 pl-9"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && applyFilters()}
            />
          </div>
          <select
            className="select select-sm rounded-lg border-base-300"
            name="select-status"
            aria-label="Filter by status"
            value={status}
            onChange={(e) => setParam('status', e.target.value)}
          >
            <option value="">All statuses</option>
            <option value="active">Active</option>
            <option value="resigned">Resigned</option>
            <option value="terminated">Terminated</option>
            <option value="on_leave">On leave</option>
          </select>
          <select
            className="select select-sm rounded-lg border-base-300"
            name="select-departments"
            aria-label="Filter by department"
            value={departmentId}
            onChange={(e) => setParam('departmentId', e.target.value)}
          >
            <option value="">All departments</option>
            {departments?.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>

          <div className="ml-auto flex items-center gap-2">
            <div className="dropdown dropdown-end">
              <button type="button" tabIndex={0} className="btn btn-sm btn-outline gap-2">
                <SlidersHorizontal className="h-4 w-4" /> Columns
              </button>
              <ul className="dropdown-content menu z-10 mt-1 w-56 rounded-box border border-base-300 bg-base-100 p-2 shadow">
                {OPTIONAL_COLUMNS.map((col) => (
                  <li key={col.key}>
                    <label className="flex cursor-pointer items-center gap-2">
                      <input
                        type="checkbox"
                        className="checkbox checkbox-sm"
                        checked={visibleCols.has(col.key)}
                        onChange={() => toggleCol(col.key)}
                      />
                      <span className="text-sm">{col.label}</span>
                    </label>
                  </li>
                ))}
              </ul>
            </div>
            <button
              type="button"
              className="btn btn-sm btn-outline gap-2"
              onClick={exportCsv}
              disabled={exporting || total === 0}
            >
              <Download className="h-4 w-4" />
              {exporting ? 'Exporting…' : 'Export CSV'}
            </button>
          </div>
        </div>

        {isLoading ? (
          <div className="flex justify-center p-16">
            <span className="loading loading-spinner loading-lg text-primary" />
          </div>
        ) : data?.employees.length === 0 ? (
          <EmptyState
            icon={UsersRound}
            title="No employees found"
            description="Try adjusting your filters, or add your first employee to get started."
            action={
              hasPermission('employee.create') && (
                <Link to="/hris/employees/new" className="btn btn-primary btn-sm gap-2">
                  <Plus className="h-4 w-4" />
                  Add Employee
                </Link>
              )
            }
          />
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="table">
                <thead>
                  <tr className="text-xs uppercase tracking-wide text-base-content/50">
                    <th>Employee</th>
                    {shownCols.map((c) => (
                      <th key={c.key}>{c.label}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {data?.employees.map((emp) => (
                    <tr key={emp.id} className="hover:bg-base-200/60">
                      <td>
                        <div className="flex items-center gap-3">
                          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
                            {(emp.firstName[0] ?? '').toUpperCase()}
                            {(emp.lastName?.[0] ?? '').toUpperCase()}
                          </span>
                          <div className="min-w-0">
                            <Link
                              to={`/hris/employees/${emp.id}`}
                              className="font-medium text-base-content hover:text-primary"
                            >
                              {emp.firstName} {emp.lastName}
                            </Link>
                            <p className="truncate text-xs text-base-content/50">{emp.email}</p>
                          </div>
                        </div>
                      </td>
                      {shownCols.map((c) => (
                        <Fragment key={c.key}>{renderCell(emp, c.key)}</Fragment>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3 border-t border-base-300 px-4 py-3">
              <p className="text-sm text-base-content/60">
                Showing <span className="font-medium text-base-content">{rangeStart}</span>–
                <span className="font-medium text-base-content">{rangeEnd}</span> of{' '}
                <span className="font-medium text-base-content">{total}</span>
              </p>
              {totalPages > 1 && (
                <div className="join">
                  <button
                    type="button"
                    className="btn btn-sm join-item"
                    title="Previous page"
                    aria-label="Previous page"
                    disabled={page <= 1}
                    onClick={() => goToPage(page - 1)}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </button>
                  <span className="btn btn-sm join-item btn-disabled">
                    Page {page} of {totalPages}
                  </span>
                  <button
                    type="button"
                    className="btn btn-sm join-item"
                    title="Next page"
                    aria-label="Next page"
                    disabled={page >= totalPages}
                    onClick={() => goToPage(page + 1)}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
