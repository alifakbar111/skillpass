import { Check, FileUp, Sparkles, Wand2 } from 'lucide-react';
import { useRef, useState } from 'react';
import { LoadingSpinner } from '@/components/ui/LoadingFallback';
import { ApiError, api } from '@/lib/api';
import type { Experience } from '@/lib/api-types';
import { type ParsedExperience, type ParsedResume, parseResume, uploadResume } from '@/lib/resume';

interface Props {
  onExperienceAdded: (exp: Experience) => void;
  /** Controlled open state so onboarding can surface the importer. */
  open: boolean;
  onToggle: (open: boolean) => void;
}

export function ResumeImport({ onExperienceAdded, open, onToggle }: Props) {
  const [text, setText] = useState('');
  const [parsing, setParsing] = useState(false);
  const [parsed, setParsed] = useState<ParsedResume | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [addedIdx, setAddedIdx] = useState<Set<number>>(new Set());
  const [savingIdx, setSavingIdx] = useState<number | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  async function handleParse() {
    if (text.trim().length < 30 || parsing) return;
    setParsing(true);
    setError(null);
    setParsed(null);
    setAddedIdx(new Set());
    try {
      const result = await parseResume(text);
      setParsed(result);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to parse resume');
    } finally {
      setParsing(false);
    }
  }

  async function handleFileUpload(file: File) {
    if (parsing) return;
    setParsing(true);
    setError(null);
    setParsed(null);
    setAddedIdx(new Set());
    try {
      const result = await uploadResume(file);
      setParsed(result);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to read PDF');
    } finally {
      setParsing(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  }

  async function addOne(exp: ParsedExperience, idx: number) {
    setSavingIdx(idx);
    setError(null);
    try {
      const added = await api<Experience>('/profiles/me/experience', {
        method: 'POST',
        body: {
          type: exp.type,
          title: exp.title,
          organization: exp.organization,
          startDate: exp.startDate,
          endDate: exp.isCurrent ? undefined : exp.endDate || undefined,
          isCurrent: exp.isCurrent,
          description: exp.description || undefined,
          skillsUsed: exp.skillsUsed ?? [],
        },
      });
      onExperienceAdded(added);
      setAddedIdx((prev) => new Set(prev).add(idx));
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to add entry');
    } finally {
      setSavingIdx(null);
    }
  }

  async function addAll() {
    if (!parsed) return;
    for (let i = 0; i < parsed.experiences.length; i++) {
      if (!addedIdx.has(i)) {
        // Sequential to keep ordering and avoid rate spikes.
        // eslint-disable-next-line no-await-in-loop
        await addOne(parsed.experiences[i], i);
      }
    }
  }

  return (
    <div className="card bg-base-200 p-4">
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-2">
          <Sparkles size={18} className="text-primary" aria-hidden="true" />
          <h2 className="font-semibold">Import from Resume</h2>
        </div>
        <button type="button" className="btn btn-ghost btn-sm" onClick={() => onToggle(!open)}>
          {open ? 'Hide' : 'Open'}
        </button>
      </div>

      {open && (
        <div className="mt-3 space-y-3">
          <p className="text-sm opacity-70">
            Upload your resume PDF or paste its text — AI extracts your experiences. Review each entry before adding it
            to your profile.
          </p>
          <input
            ref={fileInputRef}
            type="file"
            accept="application/pdf"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (file) handleFileUpload(file);
            }}
          />
          <button
            type="button"
            className="btn btn-outline btn-sm gap-1"
            onClick={() => fileInputRef.current?.click()}
            disabled={parsing}
          >
            <FileUp size={16} aria-hidden="true" /> Upload PDF
          </button>
          <div className="divider text-xs opacity-50 my-1">or paste text</div>
          <textarea
            className="textarea textarea-bordered w-full h-40 font-mono text-xs"
            placeholder="Paste your resume text here…"
            value={text}
            onChange={(e) => setText(e.target.value)}
          />
          <button
            type="button"
            className="btn btn-primary btn-sm gap-1"
            onClick={handleParse}
            disabled={parsing || text.trim().length < 30}
          >
            {parsing ? <LoadingSpinner size="sm" /> : <Wand2 size={16} />} Parse Resume
          </button>

          {error && <p className="text-error text-sm">{error}</p>}

          {parsed && (
            <div className="space-y-3 mt-2">
              {(parsed.headline || parsed.about) && (
                <div className="bg-base-100 rounded-box p-3 text-sm">
                  {parsed.headline && (
                    <p>
                      <span className="opacity-50">Suggested headline: </span>
                      {parsed.headline}
                    </p>
                  )}
                  {parsed.about && (
                    <p className="mt-1">
                      <span className="opacity-50">Summary: </span>
                      {parsed.about}
                    </p>
                  )}
                </div>
              )}

              <div className="flex justify-between items-center">
                <h3 className="font-medium text-sm">
                  Found {parsed.experiences.length} {parsed.experiences.length === 1 ? 'entry' : 'entries'}
                </h3>
                {parsed.experiences.length > 0 && (
                  <button type="button" className="btn btn-outline btn-xs" onClick={addAll}>
                    Add all
                  </button>
                )}
              </div>

              <div className="space-y-2">
                {parsed.experiences.map((exp, idx) => (
                  <div
                    // biome-ignore lint/suspicious/noArrayIndexKey: parsed entries have no stable id; title+org+idx disambiguates duplicates
                    key={`${exp.title}-${exp.organization}-${idx}`}
                    className="bg-base-100 rounded-box p-3 flex justify-between items-start gap-3"
                  >
                    <div className="min-w-0">
                      <p className="font-medium text-sm">{exp.title}</p>
                      <p className="text-xs opacity-60">
                        {exp.organization} · {exp.startDate}
                        {exp.isCurrent ? ' - Present' : exp.endDate ? ` - ${exp.endDate}` : ''} · {exp.type}
                      </p>
                      {exp.skillsUsed?.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {exp.skillsUsed.slice(0, 6).map((s) => (
                            <span key={s} className="badge badge-xs badge-ghost">
                              {s}
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                    {addedIdx.has(idx) ? (
                      <span className="badge badge-success gap-1">
                        <Check size={12} /> Added
                      </span>
                    ) : (
                      <button
                        type="button"
                        className="btn btn-primary btn-xs"
                        onClick={() => addOne(exp, idx)}
                        disabled={savingIdx === idx}
                      >
                        {savingIdx === idx ? <span className="loading loading-spinner loading-xs" /> : 'Add'}
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
