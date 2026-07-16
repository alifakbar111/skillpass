import { Award, Briefcase, Check, Code, Download, FileUp, GraduationCap, Heart, Sparkles, Wand2 } from 'lucide-react';
import { type ReactNode, useRef, useState } from 'react';
import { LoadingSpinner } from '@/components/ui/LoadingFallback';
import { ApiError, api } from '@/lib/api';
import type { Experience } from '@/lib/api-types';
import { type ParsedExperience, type ParsedResume, parseResume, uploadResume } from '@/lib/resume';

/** Map raw parser types to valid experience type enum values. */
function normalizeType(raw: string): string {
  const aliases: Record<string, string> = {
    'side project': 'project',
    side_project: 'project',
    research: 'project',
    'research project': 'project',
    certificate: 'certification',
    volunteer: 'volunteering',
    'volunteer work': 'volunteering',
    freelance: 'gig',
    contract: 'gig',
    internship: 'employment',
    intern: 'employment',
  };
  return aliases[raw.toLowerCase().trim()] ?? raw;
}

interface SectionMeta {
  label: string;
  icon: ReactNode;
}

const SECTION_META: Record<string, SectionMeta> = {
  employment: { label: 'Work History', icon: <Briefcase size={14} /> },
  gig: { label: 'Work History', icon: <Briefcase size={14} /> },
  education: { label: 'Education', icon: <GraduationCap size={14} /> },
  certification: { label: 'Certifications & Licenses', icon: <Award size={14} /> },
  project: { label: 'Projects & Portfolio', icon: <Code size={14} /> },
  volunteering: { label: 'Volunteering', icon: <Heart size={14} /> },
};

/** The canonical section key for a type (employment+gig collapse to one section). */
function sectionKey(type: string): string {
  return type === 'employment' || type === 'gig' ? 'employment' : type;
}

/** Display order for sections. */
const SECTION_ORDER = ['employment', 'education', 'certification', 'project', 'volunteering'] as const;

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
  const [addingAll, setAddingAll] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  /** Reset all import state (parsed data, index tracking, errors). */
  function resetImport() {
    setParsed(null);
    setAddedIdx(new Set());
    setError(null);
  }

  async function handleParse() {
    if (text.trim().length < 30 || parsing) return;
    setParsing(true);
    setError(null);
    resetImport();
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
    resetImport();
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
          type: normalizeType(exp.type),
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

  function downloadMarkdown() {
    if (!parsed?.rawMarkdown) return;
    const blob = new Blob([parsed.rawMarkdown], { type: 'text/markdown' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'resume.md';
    a.click();
    URL.revokeObjectURL(url);
  }

  async function addAll() {
    if (!parsed) return;
    setAddingAll(true);
    setError(null);
    const indices = parsed.experiences.map((_, i) => i).filter((i) => !addedIdx.has(i));
    for (const i of indices) {
      try {
        await addOne(parsed.experiences[i], i);
      } catch {
        // Continue with remaining entries even if one fails.
      }
    }
    setAddingAll(false);
  }

  /** Group parsed experiences by section for display. */
  const grouped = parsed
    ? SECTION_ORDER.reduce(
        (acc, key) => {
          const entries = parsed.experiences
            .map((exp, idx) => ({ exp: { ...exp, type: normalizeType(exp.type) }, idx }))
            .filter(({ exp }) => sectionKey(exp.type) === key);
          if (entries.length > 0) acc.push({ key, entries });
          return acc;
        },
        [] as { key: string; entries: { exp: ParsedExperience; idx: number }[] }[],
      )
    : [];

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
            aria-label="pdf-input"
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
                <div className="flex gap-1">
                  {parsed.rawMarkdown && (
                    <button type="button" className="btn btn-outline btn-xs gap-1" onClick={downloadMarkdown}>
                      <Download size={12} /> Save .md
                    </button>
                  )}
                  {parsed.experiences.length > 0 && (
                    <button type="button" className="btn btn-outline btn-xs" onClick={addAll} disabled={addingAll}>
                      {addingAll ? <span className="loading loading-spinner loading-xs" /> : null}
                      {addingAll ? 'Adding…' : 'Add all'}
                    </button>
                  )}
                </div>
              </div>

              {/* Entries grouped by profile section */}
              <div className="space-y-4">
                {grouped.map(({ key, entries }) => {
                  const meta = SECTION_META[key];
                  return (
                    <div key={key}>
                      <h4 className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider opacity-60 mb-1.5">
                        {meta?.icon}
                        {meta?.label ?? key}
                        <span className="ml-auto font-normal normal-case opacity-50">
                          {entries.length} {entries.length === 1 ? 'entry' : 'entries'}
                        </span>
                      </h4>
                      <div className="space-y-1.5">
                        {entries.map(({ exp, idx }) => (
                          <div
                            key={`${exp.title}-${exp.organization}-${idx}`}
                            className="bg-base-100 rounded-box p-3 flex justify-between items-start gap-3"
                          >
                            <div className="min-w-0">
                              <p className="font-medium text-sm">{exp.title}</p>
                              <p className="text-xs opacity-60">
                                {exp.organization} · {exp.startDate}
                                {exp.isCurrent ? ' - Present' : exp.endDate ? ` - ${exp.endDate}` : ''}
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
                              <span className="badge badge-success gap-1 shrink-0">
                                <Check size={12} /> Added
                              </span>
                            ) : (
                              <button
                                type="button"
                                className="btn btn-primary btn-xs shrink-0"
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
                  );
                })}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
