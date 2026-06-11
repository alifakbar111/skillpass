import type { Application, ApplicationStatus } from '../../lib/application';

interface Props {
  applications: Application[];
}

const columns: { status: ApplicationStatus; title: string; color: string }[] = [
  { status: 'applied', title: 'Applied', color: 'border-l-info' },
  { status: 'reviewed', title: 'Reviewing', color: 'border-l-warning' },
  { status: 'interviewed', title: 'Interviewing', color: 'border-l-accent' },
  { status: 'offered', title: 'Offers', color: 'border-l-success' },
];

export function ApplicationKanban({ applications }: Props) {
  const grouped = columns.map((col) => ({
    ...col,
    items: applications.filter((a) => a.status === col.status),
  }));

  const rejected = applications.filter((a) => a.status === 'rejected');

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {grouped.map((col) => (
          <div key={col.status} className={`card bg-base-200 border-l-4 ${col.color}`}>
            <div className="card-body p-4">
              <div className="flex justify-between items-center mb-3">
                <h3 className="font-semibold">{col.title}</h3>
                <span className="badge badge-ghost badge-sm">{col.items.length}</span>
              </div>
              <div className="space-y-2 min-h-30">
                {col.items.length === 0 ? (
                  <p className="text-sm opacity-50 italic">No applications</p>
                ) : (
                  col.items.map((app) => (
                    <div key={app.id} className="p-3 bg-base-100 rounded-box">
                      <p className="font-medium text-sm">{app.jobTitle || 'Unknown Position'}</p>
                      <p className="text-xs opacity-60">{app.companyName || 'Unknown Company'}</p>
                      <p className="text-xs opacity-40 mt-1">
                        Applied {app.createdAt ? new Date(app.createdAt).toLocaleDateString() : 'recently'}
                      </p>
                      {app.latestNote && (
                        <div className="mt-2 text-xs bg-base-200 rounded p-2 border-l-2 border-l-primary">
                          <span className="opacity-50">Note: </span>
                          {app.latestNote}
                        </div>
                      )}
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        ))}
      </div>

      {rejected.length > 0 && (
        <details className="collapse collapse-arrow bg-base-200">
          <summary className="collapse-title font-medium text-sm opacity-60">Rejected ({rejected.length})</summary>
          <div className="collapse-content space-y-2">
            {rejected.map((app) => (
              <div key={app.id} className="p-3 bg-base-100 rounded-box">
                <p className="font-medium text-sm">{app.jobTitle || 'Unknown Position'}</p>
                <p className="text-xs opacity-60">{app.companyName || 'Unknown Company'}</p>
                {app.latestNote && (
                  <div className="mt-2 text-xs bg-base-200 rounded p-2 border-l-2 border-l-primary">
                    <span className="opacity-50">Note: </span>
                    {app.latestNote}
                  </div>
                )}
              </div>
            ))}
          </div>
        </details>
      )}
    </div>
  );
}
