import { ExternalLink } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { api } from '../lib/api';

interface PassportData {
  name: string;
  avatarUrl?: string;
  headline?: string;
  about?: string;
  yearsOfExperience?: number;
  experiences: Array<{
    type: string;
    title: string;
    organization: string;
    startDate: string;
    endDate?: string;
    isCurrent: boolean;
    description?: string;
    skillsUsed?: string[];
  }>;
}

export function JobseekerPassport() {
  const { user } = useAuth();
  const [data, setData] = useState<PassportData | null>(null);

  useEffect(() => {
    if (user) {
      api<PassportData>(`/profiles/${user.username}`).then(setData);
    }
  }, [user]);

  if (!data)
    return (
      <div className="flex justify-center p-8" role="status" aria-label="Loading passport">
        <span className="loading loading-spinner loading-lg" aria-hidden="true" />
      </div>
    );

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">My Passport</h1>
          <Link to={`/profiles/${user?.username}`} className="btn btn-outline btn-sm gap-2" target="_blank">
            <ExternalLink size={14} aria-hidden="true" /> View Public
        </Link>
      </div>

      <div className="card bg-base-200 p-6">
        <div className="flex items-center gap-4 mb-4">
          <div className="avatar placeholder">
            <div className="bg-neutral text-neutral-content rounded-full w-16">
              <span className="text-xl">{data.name?.charAt(0)}</span>
            </div>
          </div>
          <div>
            <h2 className="text-xl font-bold">{data.name}</h2>
            {data.headline && <p className="text-muted-strong">{data.headline}</p>}
            {data.yearsOfExperience !== undefined && (
              <p className="text-sm text-muted">{data.yearsOfExperience} years of experience</p>
            )}
          </div>
        </div>
        {data.about && <p className="text-muted-strong mb-4">{data.about}</p>}
      </div>

      <div className="card bg-base-200 p-4">
        <h3 className="font-semibold mb-3">Experience</h3>
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
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
