import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle2, ScanFace, Sun, User } from 'lucide-react';
import { useState } from 'react';
import { FaceCapture } from '@/components/hris/FaceCapture';
import { PageHeader } from '@/components/hris/ui';
import { ApiError } from '@/lib/api';
import { enrollFace, getFaceStatus } from '@/lib/hris/face';

const GUIDE = [
  { icon: User, text: 'Center your face inside the oval' },
  { icon: Sun, text: 'Find even, front-on lighting' },
  { icon: ScanFace, text: 'Look straight at the camera' },
];

export default function FaceEnrollment() {
  const qc = useQueryClient();
  const { data: status, isLoading } = useQuery({ queryKey: ['hris', 'face', 'status'], queryFn: getFaceStatus });

  const [captured, setCaptured] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => enrollFace(captured as string),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'face', 'status'] });
      setCaptured(null);
      setError(null);
    },
    onError: (err) => setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Enrollment failed'),
  });

  return (
    <div className="max-w-3xl">
      <PageHeader title="Face ID" subtitle="Enrol your face for attendance verification." />

      {!isLoading && status?.enrolled && (
        <div className="mb-6 flex items-center gap-3 rounded-box border border-success/30 bg-success/5 p-4">
          <CheckCircle2 className="h-5 w-5 text-success" />
          <div className="text-sm">
            <p className="font-medium text-success">Your face is enrolled.</p>
            {status.enrolledAt && (
              <p className="text-base-content/60">Last updated {new Date(status.enrolledAt).toLocaleDateString()}</p>
            )}
          </div>
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-[1fr_16rem]">
        <div className="rounded-box border border-base-300 bg-base-100 p-6">
          <FaceCapture captured={captured} onCapture={(d) => setCaptured(d || null)} />

          {error && <div className="alert alert-error mt-4 text-sm">{error}</div>}

          <div className="mt-5 flex justify-center">
            <button
              type="button"
              className="btn btn-primary gap-2"
              disabled={!captured || mutation.isPending}
              onClick={() => mutation.mutate()}
            >
              <ScanFace className="h-4 w-4" />
              {mutation.isPending ? 'Enrolling…' : status?.enrolled ? 'Re-enrol face' : 'Enrol face'}
            </button>
          </div>
        </div>

        <aside className="rounded-box border border-base-300 bg-base-100 p-5 h-fit">
          <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-base-content/50">For a good scan</h3>
          <ul className="space-y-3">
            {GUIDE.map(({ icon: Icon, text }) => (
              <li key={text} className="flex items-start gap-3 text-sm">
                <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
                  <Icon className="h-4 w-4" />
                </span>
                <span className="text-base-content/70">{text}</span>
              </li>
            ))}
          </ul>
          <p className="mt-4 text-xs text-base-content/40">
            Your facial data is encrypted and used only for identity verification.
          </p>
        </aside>
      </div>
    </div>
  );
}
