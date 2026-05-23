import { Moon, Sun } from 'lucide-react';
import { useEffect, useState } from 'react';

export function ThemeToggle() {
  const [dark, setDark] = useState(() => localStorage.getItem('theme') === 'dark');

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'winter');
    localStorage.setItem('theme', dark ? 'dark' : 'winter');
  }, [dark]);

  return (
    <button type="button" className="btn btn-ghost btn-circle" onClick={() => setDark(!dark)} aria-label="Toggle theme">
      {dark ? <Sun size={20} /> : <Moon size={20} />}
    </button>
  );
}
