import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
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

export function PublicPassport() {
  const { username } = useParams();
  const [data, setData] = useState<PassportData | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!username) return;
    api<PassportData>(`/profiles/${username}`)
      .then(setData)
      .catch(() => setError('Profile not found'));
  }, [username]);

  if (error)
    return (
      <div className="text-center p-8">
        <p className="text-error">{error}</p>
      </div>
    );
  if (!data)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="card bg-base-200 p-6">
        <div className="flex items-center gap-4 mb-4">
          <div className="avatar placeholder">
            <div className="bg-neutral text-neutral-content rounded-full w-20">
              <span className="text-2xl">{data.name?.charAt(0)}</span>
            </div>
          </div>
          <div>
            <h1 className="text-2xl font-bold">{data.name}</h1>
            {data.headline && <p className="opacity-70">{data.headline}</p>}
            {data.yearsOfExperience !== undefined && (
              <p className="text-sm opacity-50">{data.yearsOfExperience} years of experience</p>
            )}
          </div>
        </div>
        {data.about && <p className="opacity-70 mb-4">{data.about}</p>}
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
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
