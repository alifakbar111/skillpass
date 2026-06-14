import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Clock, MapPin } from 'lucide-react';
import { useEffect, useState } from 'react';
import { clockIn, clockOut } from '@/lib/hris/attendance';

export default function ClockInPage() {
  const qc = useQueryClient();
  const [lat, setLat] = useState<number | null>(null);
  const [lng, setLng] = useState<number | null>(null);
  const [geoError, setGeoError] = useState('');
  const [now, setNow] = useState(new Date());

  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!navigator.geolocation) {
      setGeoError('Geolocation is not supported by your browser');
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLat(pos.coords.latitude);
        setLng(pos.coords.longitude);
      },
      () => setGeoError('Unable to get location. Please enable GPS.'),
      { enableHighAccuracy: true },
    );
  }, []);

  const clockInMut = useMutation({
    mutationFn: () => clockIn({ lat: lat ?? 0, lng: lng ?? 0 }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'my-attendance'] });
      qc.invalidateQueries({ queryKey: ['hris', 'attendance-dashboard'] });
    },
  });

  const clockOutMut = useMutation({
    mutationFn: () => clockOut({ lat: lat ?? 0, lng: lng ?? 0 }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'my-attendance'] });
      qc.invalidateQueries({ queryKey: ['hris', 'attendance-dashboard'] });
    },
  });

  return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <div className="card bg-base-100 shadow-xl w-full max-w-md border border-base-300">
        <div className="card-body items-center text-center">
          <Clock className="h-16 w-16 text-primary mb-2" />
          <h1 className="text-4xl font-bold font-mono tabular-nums">{now.toLocaleTimeString()}</h1>
          <p className="text-base-content/60">
            {now.toLocaleDateString(undefined, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}
          </p>

          {geoError ? (
            <div className="alert alert-warning mt-4">
              <MapPin className="h-4 w-4" /> {geoError}
            </div>
          ) : lat && lng ? (
            <div className="flex items-center gap-1 text-sm text-success mt-2">
              <MapPin className="h-4 w-4" /> GPS: {lat.toFixed(5)}, {lng.toFixed(5)}
            </div>
          ) : (
            <div className="flex items-center gap-2 mt-2 text-sm">
              <span className="loading loading-spinner loading-xs" /> Getting location...
            </div>
          )}

          <div className="grid grid-cols-2 gap-3 w-full mt-6">
            <button
              type="button"
              className="btn btn-primary btn-lg"
              disabled={lat === null || clockInMut.isPending}
              onClick={() => clockInMut.mutate()}
            >
              {clockInMut.isPending ? <span className="loading loading-spinner loading-sm" /> : 'Clock In'}
            </button>
            <button
              type="button"
              className="btn btn-secondary btn-lg"
              disabled={lat === null || clockOutMut.isPending}
              onClick={() => clockOutMut.mutate()}
            >
              {clockOutMut.isPending ? <span className="loading loading-spinner loading-sm" /> : 'Clock Out'}
            </button>
          </div>

          {clockInMut.isSuccess && (
            <div className="alert alert-success mt-4">
              Clocked in at {new Date(clockInMut.data.clockIn ?? '').toLocaleTimeString()}
              {clockInMut.data.isLate && ` (Late by ${clockInMut.data.lateMinutes} min)`}
            </div>
          )}
          {clockOutMut.isSuccess && (
            <div className="alert alert-success mt-4">
              Clocked out at {new Date(clockOutMut.data.clockOut ?? '').toLocaleTimeString()}
            </div>
          )}
          {(clockInMut.isError || clockOutMut.isError) && (
            <div className="alert alert-error mt-4">
              {(clockInMut.error as Error)?.message || (clockOutMut.error as Error)?.message}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
