import { CalendarClock, X } from 'lucide-react';
import { useState } from 'react';
import { ApiError } from '@/lib/api';
import { type Application, scheduleInterview } from '@/lib/application';

interface Props {
  applicationId: string;
  candidateName: string;
  onClose: () => void;
  onScheduled: (updated: Application) => void;
}

export function InterviewScheduleModal({ applicationId, candidateName, onClose, onScheduled }: Props) {
  const [scheduledAt, setScheduledAt] = useState('');
  const [mode, setMode] = useState<'onsite' | 'online'>('onsite');
  const [location, setLocation] = useState('');
  const [meetingLink, setMeetingLink] = useState('');
  const [interviewer, setInterviewer] = useState('');
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isUrlValid = (url: string) => {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  };

  const placeValid =
    mode === 'onsite' ? location.trim() !== '' : meetingLink.trim() !== '' && isUrlValid(meetingLink.trim());
  const canSubmit = scheduledAt !== '' && placeValid && !submitting;

  async function handleSubmit() {
    if (!canSubmit) return;
    setSubmitting(true);
    setError(null);
    try {
      // datetime-local yields a local "YYYY-MM-DDTHH:mm"; convert to RFC3339.
      const iso = new Date(scheduledAt).toISOString();
      const updated = await scheduleInterview(applicationId, {
        scheduledAt: iso,
        mode,
        location: mode === 'onsite' ? location.trim() : undefined,
        meetingLink: mode === 'online' ? meetingLink.trim() : undefined,
        interviewer: interviewer.trim() || undefined,
        notes: notes.trim() || undefined,
      });
      onScheduled(updated);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to schedule interview');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="modal modal-open" role="dialog" aria-modal="true">
      <div className="modal-box">
        <div className="flex items-start justify-between">
          <h3 className="text-lg font-bold flex items-center gap-2">
            <CalendarClock size={20} className="text-primary" />
            Schedule interview
          </h3>
          <button type="button" className="btn btn-sm btn-ghost btn-circle" onClick={onClose} aria-label="Close">
            <X size={18} />
          </button>
        </div>
        <p className="text-sm opacity-70 mt-1">
          Invite <span className="font-medium">{candidateName}</span> to interview. They'll be notified by email and
          in-app.
        </p>

        <div className="mt-4 space-y-4">
          <div>
            <label className="label" htmlFor="interview-datetime">
              <span className="label-text">Date &amp; time</span>
            </label>
            <input
              id="interview-datetime"
              type="datetime-local"
              className="input input-bordered w-full"
              value={scheduledAt}
              onChange={(e) => setScheduledAt(e.target.value)}
            />
          </div>

          <div>
            <span className="label-text">Mode</span>
            <div className="join mt-1 w-full">
              <button
                type="button"
                className={`btn join-item flex-1 ${mode === 'onsite' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setMode('onsite')}
              >
                Onsite
              </button>
              <button
                type="button"
                className={`btn join-item flex-1 ${mode === 'online' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setMode('online')}
              >
                Online
              </button>
            </div>
          </div>

          {mode === 'onsite' ? (
            <div>
              <label className="label" htmlFor="interview-location">
                <span className="label-text">Address</span>
              </label>
              <input
                id="interview-location"
                type="text"
                className="input input-bordered w-full"
                placeholder="Office address / meeting room"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
              />
            </div>
          ) : (
            <div>
              <label className="label" htmlFor="interview-link">
                <span className="label-text">Meeting link</span>
              </label>
              <input
                id="interview-link"
                type="url"
                className="input input-bordered w-full"
                placeholder="https://meet.example.com/..."
                value={meetingLink}
                onChange={(e) => setMeetingLink(e.target.value)}
              />
              {mode === 'online' && meetingLink.trim() !== '' && !isUrlValid(meetingLink.trim()) && (
                <p className="text-error text-xs mt-1">Please enter a valid URL (e.g. https://meet.google.com/...)</p>
              )}
            </div>
          )}

          <div>
            <label className="label" htmlFor="interview-interviewer">
              <span className="label-text">
                Interviewer <span className="opacity-50">(optional)</span>
              </span>
            </label>
            <input
              id="interview-interviewer"
              type="text"
              className="input input-bordered w-full"
              placeholder="e.g. Head of People"
              value={interviewer}
              onChange={(e) => setInterviewer(e.target.value)}
            />
          </div>

          <div>
            <label className="label" htmlFor="interview-notes">
              <span className="label-text">
                Note to candidate <span className="opacity-50">(optional)</span>
              </span>
            </label>
            <textarea
              id="interview-notes"
              className="textarea textarea-bordered w-full"
              rows={2}
              placeholder="Anything they should prepare or bring."
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>

          {error && <div className="alert alert-error text-sm">{error}</div>}
        </div>

        <div className="modal-action">
          <button type="button" className="btn btn-ghost" onClick={onClose} disabled={submitting}>
            Cancel
          </button>
          <button type="button" className="btn btn-primary" onClick={handleSubmit} disabled={!canSubmit}>
            {submitting ? 'Scheduling…' : 'Schedule & notify'}
          </button>
        </div>
      </div>
      <button type="button" className="modal-backdrop" onClick={onClose} aria-label="Close">
        close
      </button>
    </div>
  );
}
