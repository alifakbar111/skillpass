import { useEffect, useState } from 'react';
import { api } from '../../lib/api';
import { ChecklistCard, type ChecklistStep } from './ChecklistCard';

// CompanyOnboarding composes its state from existing endpoints — no
// dedicated backend. Shown on the verification page (the company's first stop).
export function CompanyOnboarding() {
  const [steps, setSteps] = useState<ChecklistStep[]>([]);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      const [profileRes, statusRes, jobsRes] = await Promise.allSettled([
        api<{ description?: string | null; website?: string | null }>('/company/profile'),
        api<{ verificationStatus: string }>('/company/verification-status'),
        api<unknown[]>('/jobs/me'),
      ]);
      if (cancelled) return;

      const profileDone =
        profileRes.status === 'fulfilled' && Boolean(profileRes.value.description || profileRes.value.website);
      const verified = statusRes.status === 'fulfilled' && statusRes.value.verificationStatus === 'verified';
      // /jobs/me 403s until verified — that simply counts as "not done yet".
      const hasJob = jobsRes.status === 'fulfilled' && jobsRes.value.length > 0;

      setSteps([
        {
          id: 'profile',
          label: 'Complete your company profile (description, website)',
          done: profileDone,
          to: '/company/profile',
          actionLabel: 'Edit profile',
        },
        {
          id: 'verify',
          label: 'Get your company verified',
          done: verified,
        },
        {
          id: 'job',
          label: 'Post your first job',
          done: hasJob,
          to: verified ? '/company/jobs' : undefined,
          actionLabel: 'Post job',
        },
      ]);
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  return <ChecklistCard title="Get hiring on SkillPass" steps={steps} />;
}
