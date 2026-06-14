import { Outlet } from 'react-router-dom';
import HRISSidebar from './HRISSidebar';

export default function HRISLayout() {
  return (
    <div className="flex min-h-[calc(100vh-4rem)]">
      <HRISSidebar />
      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  );
}
