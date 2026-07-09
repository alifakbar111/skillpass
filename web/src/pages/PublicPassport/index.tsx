import { useQuery } from '@tanstack/react-query';
import { Award, Briefcase, Code, ExternalLink, Eye, GraduationCap, Heart } from 'lucide-react';
import { useRef } from 'react';
import { useParams } from 'react-router-dom';
import { SharePassport } from '@/components/passport/SharePassport';
import { LoadingFallback } from '@/components/ui/LoadingFallback';
import { ApiError, api } from '@/lib/api';
import type { PublicProfile } from '@/lib/api-types';

export function PublicPassport() {
  const { username } = useParams();
  const printRef = useRef<HTMLDivElement>(null);

  const { data, error, isLoading } = useQuery({
    queryKey: ['passport', username],
    enabled: !!username,
    queryFn: () => api<PublicProfile>(`/profiles/${encodeURIComponent(username as string)}`),
  });

  const errorMessage = error
    ? error instanceof ApiError && error.status >= 400 && error.status < 500
      ? (error.serverMessage ?? error.message)
      : 'Failed to load passport'
    : null;

  if (errorMessage)
    return (
      <div className="text-center p-8">
        <p className="text-error">{errorMessage}</p>
      </div>
    );
  if (isLoading || !data) return <LoadingFallback text="Loading profile" />;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="flex justify-end">
        {username && <SharePassport slug={username} name={data.name ?? ''} printRef={printRef} />}
      </div>

      <div ref={printRef} className="space-y-4">
        <div className="card bg-base-200 p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="avatar avatar-placeholder">
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

        {/* Work History - employment + gig */}
        <div className="card bg-base-200 p-4">
          <div className="flex justify-between items-center mb-3">
            <h2 className="font-semibold flex items-center gap-2">
              <Briefcase size={18} aria-hidden="true" /> Work History
            </h2>
          </div>
          {(data.experiences ?? []).filter((e) => e.type === 'employment' || e.type === 'gig').length === 0 ? (
            <p className="text-sm opacity-60 py-4 text-center">No work history listed.</p>
          ) : (
            <div className="space-y-2">
              {(data.experiences ?? [])
                .filter((e) => e.type === 'employment' || e.type === 'gig')
                .map((exp) => (
                  <div key={exp.id} className="p-3 bg-base-100 rounded-box">
                    <p className="font-medium">{exp.title}</p>
                    <p className="text-sm opacity-70">
                      {exp.organization} · {exp.startDate}
                      {exp.isCurrent ? ' - Present' : exp.endDate ? ` - ${exp.endDate}` : ''}
                    </p>
                    {exp.industry && <p className="text-xs opacity-50 mt-1">{exp.industry}</p>}
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
          )}
        </div>

        {/* Education */}
        <div className="card bg-base-200 p-4">
          <div className="flex justify-between items-center mb-3">
            <h2 className="font-semibold flex items-center gap-2">
              <GraduationCap size={18} aria-hidden="true" /> Education
            </h2>
          </div>
          {(data.experiences ?? []).filter((e) => e.type === 'education').length === 0 ? (
            <p className="text-sm opacity-60 py-4 text-center">No education listed.</p>
          ) : (
            <div className="space-y-2">
              {(data.experiences ?? [])
                .filter((e) => e.type === 'education')
                .map((exp) => (
                  <div key={exp.id} className="p-3 bg-base-100 rounded-box">
                    <p className="font-medium">{exp.title}</p>
                    <p className="text-sm opacity-70">
                      {exp.organization} · {exp.startDate}
                      {exp.isCurrent ? ' - Present' : exp.endDate ? ` - ${exp.endDate}` : ''}
                    </p>
                    {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
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
          )}
        </div>

        {/* Certifications & Licenses */}
        <div className="card bg-base-200 p-4">
          <div className="flex justify-between items-center mb-3">
            <h2 className="font-semibold flex items-center gap-2">
              <Award size={18} aria-hidden="true" /> Certifications & Licenses
            </h2>
          </div>
          {(data.experiences ?? []).filter((e) => e.type === 'certification').length === 0 ? (
            <p className="text-sm opacity-60 py-4 text-center">No certifications or licenses listed.</p>
          ) : (
            <div className="space-y-2">
              {(data.experiences ?? [])
                .filter((e) => e.type === 'certification')
                .map((exp) => (
                  <div key={exp.id} className="p-3 bg-base-100 rounded-box">
                    <p className="font-medium">{exp.title}</p>
                    <p className="text-sm opacity-70">{exp.organization}</p>
                    {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
                    {exp.url && (
                      <a
                        href={exp.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="link link-primary text-xs inline-flex items-center gap-1 mt-2"
                      >
                        <ExternalLink size={12} aria-hidden="true" /> Verify credential
                      </a>
                    )}
                  </div>
                ))}
            </div>
          )}
        </div>

        {/* Projects & Portfolio */}
        <div className="card bg-base-200 p-4">
          <div className="flex justify-between items-center mb-3">
            <h2 className="font-semibold flex items-center gap-2">
              <Code size={18} aria-hidden="true" /> Projects & Portfolio
            </h2>
          </div>
          {(data.experiences ?? []).filter((e) => e.type === 'project').length === 0 ? (
            <p className="text-sm opacity-60 py-4 text-center">No projects or portfolio items listed.</p>
          ) : (
            <div className="space-y-2">
              {(data.experiences ?? [])
                .filter((e) => e.type === 'project')
                .map((exp) => (
                  <div key={exp.id} className="p-3 bg-base-100 rounded-box">
                    <p className="font-medium">{exp.title}</p>
                    {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
                    {exp.url && (
                      <a
                        href={exp.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="link link-primary text-xs inline-flex items-center gap-1 mt-2"
                      >
                        <ExternalLink size={12} aria-hidden="true" /> View project
                      </a>
                    )}
                  </div>
                ))}
            </div>
          )}
        </div>

        {/* Volunteering */}
        <div className="card bg-base-200 p-4">
          <div className="flex justify-between items-center mb-3">
            <h2 className="font-semibold flex items-center gap-2">
              <Heart size={18} aria-hidden="true" /> Volunteering
            </h2>
          </div>
          {(data.experiences ?? []).filter((e) => e.type === 'volunteering').length === 0 ? (
            <p className="text-sm opacity-60 py-4 text-center">No volunteering experience listed.</p>
          ) : (
            <div className="space-y-2">
              {(data.experiences ?? [])
                .filter((e) => e.type === 'volunteering')
                .map((exp) => (
                  <div key={exp.id} className="p-3 bg-base-100 rounded-box">
                    <p className="font-medium">{exp.title}</p>
                    <p className="text-sm opacity-70">{exp.organization}</p>
                    {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
                  </div>
                ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
