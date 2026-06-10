import { Bell } from 'lucide-react';
import { useCallback, useEffect, useId, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  getNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  type Notification,
} from '../../lib/notifications';

const POLL_INTERVAL_MS = 60_000;

export function NotificationBell() {
  const navigate = useNavigate();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unread, setUnread] = useState(0);
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const menuId = useId();

  const load = useCallback(async () => {
    try {
      const data = await getNotifications();
      setNotifications(data.notifications);
      setUnread(data.unreadCount);
    } catch {
      // Silent — bell is non-critical UI.
    }
  }, []);

  useEffect(() => {
    load();
    const interval = setInterval(load, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [load]);

  useEffect(() => {
    if (!open) return;
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [open]);

  async function handleItemClick(n: Notification) {
    setOpen(false);
    if (!n.readAt) {
      try {
        await markNotificationRead(n.id);
        setUnread((u) => Math.max(0, u - 1));
        setNotifications((prev) => prev.map((x) => (x.id === n.id ? { ...x, readAt: new Date().toISOString() } : x)));
      } catch {
        // ignore
      }
    }
    if (n.link) navigate(n.link);
  }

  async function handleMarkAll() {
    try {
      await markAllNotificationsRead();
      setUnread(0);
      setNotifications((prev) => prev.map((x) => ({ ...x, readAt: x.readAt ?? new Date().toISOString() })));
    } catch {
      // ignore
    }
  }

  return (
    <div className="dropdown dropdown-end" ref={ref}>
      <button
        type="button"
        className="btn btn-ghost btn-circle"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        aria-label={unread > 0 ? `Notifications, ${unread} unread` : 'Notifications'}
      >
        <div className="indicator">
          <Bell size={20} aria-hidden="true" />
          {unread > 0 && (
            <span className="badge badge-xs badge-primary indicator-item">{unread > 9 ? '9+' : unread}</span>
          )}
        </div>
      </button>

      {open && (
        <div id={menuId} className="dropdown-content bg-base-100 rounded-box z-50 mt-3 w-80 p-2 shadow-md">
          <div className="flex justify-between items-center px-2 py-1">
            <span className="font-semibold text-sm">Notifications</span>
            {unread > 0 && (
              <button type="button" className="btn btn-ghost btn-xs" onClick={handleMarkAll}>
                Mark all read
              </button>
            )}
          </div>
          <div className="divider my-0" />
          {notifications.length === 0 ? (
            <p className="text-center text-sm opacity-60 py-6">No notifications yet.</p>
          ) : (
            <ul className="max-h-96 overflow-y-auto">
              {notifications.map((n) => (
                <li key={n.id}>
                  <button
                    type="button"
                    className={`w-full text-left rounded-lg p-2 hover:bg-base-200 transition-colors ${
                      n.readAt ? 'opacity-60' : ''
                    }`}
                    onClick={() => handleItemClick(n)}
                  >
                    <div className="flex items-start gap-2">
                      {!n.readAt && <span className="badge badge-xs badge-primary mt-1.5 shrink-0" />}
                      <div className="min-w-0">
                        <div className="font-medium text-sm">{n.title}</div>
                        <div className="text-xs opacity-70">{n.body}</div>
                        <div className="text-[10px] opacity-40 mt-0.5">{n.createdAt.slice(0, 10)}</div>
                      </div>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
