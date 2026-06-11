import { useEffect, useRef } from 'react';
import { Outlet, useLocation } from 'react-router-dom';
import { Navbar } from './Navbar';
import { VerifyEmailBanner } from './VerifyEmailBanner';

export function RootLayout() {
  const mainRef = useRef<HTMLElement>(null);
  const _location = useLocation();

  useEffect(() => {
    mainRef.current?.focus();
  }, []);

  return (
    <div className="min-h-screen flex flex-col">
      <a href="#main-content" className="skip-link">
        Skip to content
      </a>
      <nav aria-label="Main navigation">
        <Navbar />
      </nav>
      <VerifyEmailBanner />
      <main id="main-content" ref={mainRef} className="flex-1 outline-none" tabIndex={-1}>
        <Outlet />
      </main>
      <footer className="footer footer-center bg-base-200 text-base-content p-4 text-sm text-muted">
        <p>SkillPass — Build your career passport</p>
      </footer>
    </div>
  );
}
