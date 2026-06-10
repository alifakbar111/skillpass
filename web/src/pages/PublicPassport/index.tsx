import { Eye, ExternalLink } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { useParams } from 'react-router-dom';
import { SharePassport } from '../../components/passport/SharePassport';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { api } from '../../lib/api';
import type { PassportData } from './type';

export function PublicPassport() {
  const { username } = useParams();
  const [data, setData] = useState<PassportData | null>(null);
  const [error, setError] = useState('');
  const printRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!username) return;
    const safe = encodeURIComponent(username);
    let cancelled = false;
    api<PassportData>(`/profiles/${safe}`)
      .then((d) => {
        if (!cancelled) setData(d);
      })
      .catch(() => {
        if (!cancelled) setError('Profile not found');
      });
    return () => {
      cancelled = true;
    };
  }, [username]);

  if (error)
    return (
      <div className="text-center p-8">
        <p className="text-error">{error}</p>
      </div>
    );
  if (!data) return <LoadingFallback text="Loading profile" />;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="flex justify-end">
        {username && <SharePassport slug={username} name={data.name} printRef={printRef} />}
      </div>

      <div ref={printRef} className="space-y-4">
        <div className="card bg-base-200 p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="avatar placeholder">
              {data.avatarUrl ? (
                <div className="w-20 rounded-full">
                  <img src={data.avatarUrl} alt={`${data.name} avatar`} />
                </div>
              ) : (
                <div className="bg-neutral text-neutral-content rounded-full w-20">
                  <span className="text-2xl">{data.name?.charAt(0)}</span>
                </div>
              )}
            </div>
            <div>
              <h1 className="text-2xl font-bold">{data.name}</h1>
              {data.headline && <p className="text-muted-strong">{data.headline}</p>}
              {data.yearsOfExperience !== undefined && (
                <p className="text-sm text-muted">{data.yearsOfExperience} years of experience</p>
              )}
              {data.viewCount !== undefined && (
                <p className="text-xs text-muted flex items-center gap-1 mt-1">
                  <Eye size={12} aria-hidden="true" /> {data.viewCount} {data.viewCount === 1 ? 'view' : 'views'}
                </p>
              )}
            </div>
          </div>
          {data.about && <p className="text-muted-strong mb-4">{data.about}</p>}
        </div>

        <div className="card bg-base-200 p-4">
          <h2 className="font-semibold mb-3">Experience</h2>
          <div className="space-y-2">
            {data.experiences.map((exp, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: experiences array has no stable id in this view
              <div key={i} className="p-3 bg-base-100 rounded-box">
                <p className="font-medium">{exp.title}</p>
                <p className="text-sm opacity-70">
                  {exp.organization} · {exp.startDate}{' '}
                  {exp.isCurrent ? '- Present' : exp.endDate ? `- ${exp.endDate}` : ''}
                </p>
                {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
                {exp.skillsUsed && exp.skillsUsed.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {exp.skillsUsed.map((s) => (
                      <span key={s} className="badge badge-sm">
                        {s}
                      </span>
                    ))}
                  </div>
                )}
                {exp.url && (
                  <a
                    href={exp.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="link link-primary text-xs inline-flex items-center gap-1 mt-2"
                  >
                    <ExternalLink size={12} aria-hidden="true" /> View evidence
                  </a>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
