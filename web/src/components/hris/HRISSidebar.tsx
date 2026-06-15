import {
  AlertTriangle,
  Banknote,
  Briefcase,
  Building2,
  Calculator,
  Calendar,
  CalendarCheck,
  CalendarDays,
  CalendarOff,
  Clock,
  ClockArrowUp,
  FileText,
  MapPin,
  Network,
  PalmtreeIcon,
  Receipt,
  Shield,
  Users,
} from 'lucide-react';
import { NavLink } from 'react-router-dom';

const links = [
  { to: '/hris/employees', label: 'Employees', icon: Users },
  { to: '/hris/departments', label: 'Departments', icon: Building2 },
  { to: '/hris/positions', label: 'Positions', icon: Briefcase },
  { to: '/hris/branches', label: 'Branches', icon: MapPin },
  { to: '/hris/org-chart', label: 'Org Chart', icon: Network },
  { to: '/hris/roles', label: 'Roles', icon: Shield },
  { to: '/hris/shifts', label: 'Shifts', icon: ClockArrowUp },
  { to: '/hris/clock-in', label: 'Clock In', icon: Clock },
  { to: '/hris/attendance', label: 'Attendance', icon: Calendar },
  { to: '/hris/my-attendance', label: 'My Attendance', icon: Calendar },
  { to: '/hris/attendance-exceptions', label: 'Exceptions', icon: AlertTriangle },
  { to: '/hris/leave-types', label: 'Leave Types', icon: FileText },
  { to: '/hris/leave-request', label: 'Leave Request', icon: CalendarOff },
  { to: '/hris/leave-approval', label: 'Leave Approval', icon: CalendarCheck },
  { to: '/hris/leave-balance', label: 'Leave Balance', icon: CalendarDays },
  { to: '/hris/holidays', label: 'Holidays', icon: PalmtreeIcon },
  { to: '/hris/salary-components', label: 'Salary Config', icon: Banknote },
  { to: '/hris/payroll-runs', label: 'Payroll Runs', icon: Calculator },
  { to: '/hris/my-payslips', label: 'My Payslips', icon: Receipt },
];

export default function HRISSidebar() {
  return (
    <nav className="menu bg-base-200 rounded-box w-56 min-h-full p-4">
      <li className="menu-title">HRIS</li>
      {links.map(({ to, label, icon: Icon }) => (
        <li key={to}>
          <NavLink to={to} className={({ isActive }) => (isActive ? 'active' : '')}>
            <Icon className="h-4 w-4" />
            {label}
          </NavLink>
        </li>
      ))}
    </nav>
  );
}
