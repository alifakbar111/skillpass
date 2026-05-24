import { useCallback, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { ThemeToggle } from '../ui/ThemeToggle';

export function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  const toggleDropdown = () => {
    setDropdownOpen((prev) => !prev);
  };

  const handleDropdownKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        setDropdownOpen(false);
        const toggle = dropdownRef.current?.querySelector('[aria-haspopup]') as HTMLElement | null;
        toggle?.focus();
      }
    },
    [],
  );

  return (
    <div className="navbar bg-base-100 shadow-sm sticky top-0 z-50">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl font-bold">
          SkillPass
        </Link>
      </div>
      <div className="flex-none gap-2">
        {user ? (
          <>
            {user.role === 'jobseeker' && (
              <Link to="/jobseeker/profile" className="btn btn-ghost btn-sm">
                My Profile
              </Link>
            )}
            {user.role === 'company' && (
              <>
                <Link to="/company/search" className="btn btn-ghost btn-sm">
                  Search
                </Link>
                <Link to="/company/jobs" className="btn btn-ghost btn-sm">
                  Jobs
                </Link>
              </>
            )}
            <div className="dropdown dropdown-end" ref={dropdownRef}>
              <button
                type="button"
                className="btn btn-ghost btn-circle avatar placeholder"
                onClick={toggleDropdown}
                aria-haspopup="true"
                aria-expanded="false"
                aria-label="User menu"
              >
                <div className="bg-neutral text-neutral-content rounded-full w-10">
                  <span>{user.name?.charAt(0)?.toLocaleUpperCase() ?? '?'}</span>
                </div>
              </button>
              {dropdownOpen && (
                <ul
                  className="menu dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow-sm"
                  onKeyDown={handleDropdownKeyDown}
                >
                  <li className="menu-label text-xs text-muted">{user.email}</li>
                  <li><hr className="divider my-1" /></li>
                  {user.role === 'company' && (
                    <li>
                      <Link to="/company/profile" onClick={() => setDropdownOpen(false)}>
                        Company Profile
                      </Link>
                    </li>
                  )}
                  <li>
                    <button type="button" onClick={handleLogout} className="text-error">
                      Logout
                    </button>
                  </li>
                </ul>
              )}
            </div>
          </>
        ) : (
          <>
            <Link to="/auth/login" className="btn btn-ghost btn-sm">
              Login
            </Link>
            <Link to="/auth/register" className="btn btn-primary btn-sm">
              Register
            </Link>
          </>
        )}
        <ThemeToggle />
      </div>
    </div>
  );
}