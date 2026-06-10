import { useCallback, useEffect, useId, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { ThemeToggle } from '../ui/ThemeToggle';
import { NotificationBell } from './NotificationBell';

export function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  const handleLogout = async () => {
    setDropdownOpen(false);
    await logout();
    navigate('/');
  };

  const toggleDropdown = () => {
    setDropdownOpen((prev) => !prev);
  };

  useEffect(() => {
    if (!dropdownOpen) return;
    const handleClick = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [dropdownOpen]);

  const handleDropdownKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      setDropdownOpen(false);
      const toggle = dropdownRef.current?.querySelector('[aria-haspopup]') as HTMLElement | null;
      toggle?.focus();
    }
  }, []);

  return (
    <div className="navbar bg-base-100 shadow-sm sticky top-0 z-50">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl font-bold">
          SkillPass
        </Link>
      </div>
      <div className="flex flex-wrap gap-2">
        {user ? (
          <>
            {user.role === 'jobseeker' && (
              <>
                <Link to="/jobseeker/profile" className="btn btn-ghost btn-sm">
                  My Profile
                </Link>
                <Link to="/jobseeker/matches" className="btn btn-ghost btn-sm">
                  Matches
                </Link>
                <Link to="/jobseeker/applications" className="btn btn-ghost btn-sm">
                  Applications
                </Link>
              </>
            )}
            {user.role === 'company' && (
              <>
                <Link to="/company/search" className="btn btn-ghost btn-sm">
                  Search
                </Link>
                <Link to="/company/jobs" className="btn btn-ghost btn-sm">
                  Jobs
                </Link>
                <Link to="/company/applications" className="btn btn-ghost btn-sm">
                  Applications
                </Link>
              </>
            )}
            {user.role === 'admin' && (
              <Link to="/admin/verifications" className="btn btn-ghost btn-sm">
                Verifications
              </Link>
            )}
            <NotificationBell />
            <div className="dropdown dropdown-end" ref={dropdownRef}>
              <button
                type="button"
                className="btn btn-ghost btn-circle avatar avatar-placeholder"
                onClick={toggleDropdown}
                aria-haspopup="menu"
                aria-expanded={dropdownOpen}
                aria-controls={menuId}
                aria-label="User menu"
              >
                <div className="bg-neutral text-neutral-content rounded-full w-10">
                  <span>{user.name?.charAt(0)?.toLocaleUpperCase() ?? '?'}</span>
                </div>
              </button>
              {dropdownOpen && (
                <ul
                  id={menuId}
                  aria-label="User menu"
                  className="menu dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow-sm"
                  onKeyDown={handleDropdownKeyDown}
                >
                  <li role="presentation" className="menu-label text-xs text-muted">
                    {user.email}
                  </li>
                  <li>
                    <hr className="divider my-1" />
                  </li>
                  {user.role === 'company' && (
                    <li role="none">
                      <Link role="menuitem" to="/company/profile" onClick={() => setDropdownOpen(false)}>
                        Company Profile
                      </Link>
                    </li>
                  )}
                  <li role="none">
                    <button type="button" role="menuitem" onClick={handleLogout} className="text-error">
                      Logout
                    </button>
                  </li>
                </ul>
              )}
            </div>
          </>
        ) : (
          <div className="flex space-x-4">
            <Link to="/auth/login" className="btn btn-ghost btn-sm">
              Login
            </Link>
            <Link to="/auth/register" className="btn btn-primary btn-sm">
              Register
            </Link>
          </div>
        )}
        <ThemeToggle />
      </div>
    </div>
  );
}
