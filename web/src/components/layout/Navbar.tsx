import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { ThemeToggle } from '../ui/ThemeToggle';

export function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

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
            <div className="dropdown dropdown-end">
              {/* biome-ignore lint/a11y/useSemanticElements: DaisyUI dropdown toggle pattern */}
              <div tabIndex={0} role="button" className="btn btn-ghost btn-circle avatar placeholder">
                <div className="bg-neutral text-neutral-content rounded-full w-10">
                  <span>{user.name.charAt(0).toUpperCase()}</span>
                </div>
              </div>
              <ul className="menu dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow-sm">
                <li className="menu-label text-xs opacity-60">{user.email}</li>
                <div className="divider my-1" />
                {user.role === 'company' && (
                  <li>
                    <Link to="/company/profile">Company Profile</Link>
                  </li>
                )}
                <li>
                  <button type="button" onClick={handleLogout} className="text-error">
                    Logout
                  </button>
                </li>
              </ul>
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
