import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';

export function RootLayout() {
  return (
    <div className="min-h-screen flex flex-col">
      <a href="#main-content" className="skip-link">
        Skip to content
      </a>
      <nav aria-label="Main navigation">
        <Navbar />
      </nav>
      <main id="main-content" className="flex-1" tabIndex={-1}>
        <Outlet />
      </main>
      <footer className="footer footer-center bg-base-200 text-base-content p-4 text-sm text-muted">
        <p>SkillPass — Build your career passport</p>
      </footer>
    </div>
  );
}
