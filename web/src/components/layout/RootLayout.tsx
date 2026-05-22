import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';

export function RootLayout() {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="footer footer-center bg-base-200 text-base-content p-4 text-sm opacity-60">
        <p>SkillPass — Build your career passport</p>
      </footer>
    </div>
  );
}
