import {
  AlertTriangle,
  Briefcase,
  Building2,
  Calendar,
  Clock,
  ClockArrowUp,
  MapPin,
  Network,
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
