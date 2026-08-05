import {
  AlertTriangle,
  Banknote,
  BarChart3,
  Briefcase,
  Building2,
  Calculator,
  Calendar,
  CalendarCheck,
  CalendarDays,
  CalendarOff,
  ClipboardCheck,
  ClipboardList,
  Clock,
  ClockArrowUp,
  Download,
  FileArchive,
  FileText,
  type LucideIcon,
  MapPin,
  Network,
  PalmtreeIcon,
  Percent,
  Receipt,
  ScanFace,
  Shield,
  UserCircle,
  Users,
} from 'lucide-react';
import { NavLink } from 'react-router-dom';

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const groups: NavGroup[] = [
  {
    title: 'People',
    items: [
      { to: '/hris/my-info', label: 'My Info', icon: UserCircle },
      { to: '/hris/employees', label: 'Employees', icon: Users },
      { to: '/hris/documents', label: 'Documents', icon: FileArchive },
      { to: '/hris/org-chart', label: 'Org Chart', icon: Network },
      { to: '/hris/departments', label: 'Departments', icon: Building2 },
      { to: '/hris/positions', label: 'Positions', icon: Briefcase },
      { to: '/hris/branches', label: 'Branches', icon: MapPin },
    ],
  },
  {
    title: 'Time & Attendance',
    items: [
      { to: '/hris/clock-in', label: 'Clock In', icon: Clock },
      { to: '/hris/attendance', label: 'Attendance', icon: Calendar },
      { to: '/hris/my-attendance', label: 'My Attendance', icon: CalendarDays },
      { to: '/hris/shifts', label: 'Shifts', icon: ClockArrowUp },
      { to: '/hris/attendance-exceptions', label: 'Exceptions', icon: AlertTriangle },
    ],
  },
  {
    title: 'Leave',
    items: [
      { to: '/hris/leave-request', label: 'Leave Request', icon: CalendarOff },
      { to: '/hris/leave-approval', label: 'Leave Approval', icon: CalendarCheck },
      { to: '/hris/leave-balance', label: 'Leave Balance', icon: CalendarDays },
      { to: '/hris/leave-types', label: 'Leave Types', icon: FileText },
      { to: '/hris/holidays', label: 'Holidays', icon: PalmtreeIcon },
    ],
  },
  {
    title: 'Payroll',
    items: [
      { to: '/hris/salary-components', label: 'Salary Config', icon: Banknote },
      { to: '/hris/tax-config', label: 'Tax & BPJS', icon: Percent },
      { to: '/hris/payroll-runs', label: 'Payroll Runs', icon: Calculator },
      { to: '/hris/my-payslips', label: 'My Payslips', icon: Receipt },
    ],
  },
  {
    title: 'Reports',
    items: [
      { to: '/hris/analytics', label: 'HR Analytics', icon: BarChart3 },
      { to: '/hris/attendance-export', label: 'Export Attendance', icon: Download },
    ],
  },
  {
    title: 'Onboarding',
    items: [
      { to: '/hris/onboarding-templates', label: 'Templates', icon: ClipboardList },
      { to: '/hris/onboarding-checklists', label: 'Onboarding', icon: ClipboardCheck },
      { to: '/hris/my-onboarding', label: 'My Onboarding', icon: ClipboardCheck },
    ],
  },
  {
    title: 'Trust & Security',
    items: [{ to: '/hris/face-enrollment', label: 'Face ID', icon: ScanFace }],
  },
  {
    title: 'Settings',
    items: [{ to: '/hris/roles', label: 'Roles & Permissions', icon: Shield }],
  },
];

export default function HRISSidebar() {
  return (
    <nav
      aria-label="HRIS navigation"
      className="hidden md:flex w-60 shrink-0 flex-col gap-5 overflow-y-auto border-r border-base-300 bg-base-100 px-3 py-5"
    >
      {groups.map((group) => (
        <div key={group.title}>
          <p className="px-3 pb-1.5 text-xs font-semibold uppercase tracking-wider text-base-content/40">
            {group.title}
          </p>
          <ul className="flex flex-col gap-0.5">
            {group.items.map(({ to, label, icon: Icon }) => (
              <li key={to}>
                <NavLink
                  to={to}
                  className={({ isActive }) =>
                    [
                      'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-primary/10 text-primary'
                        : 'text-base-content/70 hover:bg-base-200 hover:text-base-content',
                    ].join(' ')
                  }
                >
                  <Icon className="h-[18px] w-[18px] shrink-0" />
                  <span className="truncate">{label}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </nav>
  );
}
