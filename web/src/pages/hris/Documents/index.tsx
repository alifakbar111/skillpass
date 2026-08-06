import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Download, FileText, FolderOpen, ShieldCheck, Trash2, Upload, X } from 'lucide-react';
import { useRef, useState } from 'react';
import { EmptyState, PageHeader, StatCard } from '@/components/hris/ui';
import { usePermissions } from '@/hooks/usePermissions';
import { ApiError } from '@/lib/api';
import {
  type DocumentCategory,
  deleteDocument,
  downloadDocument,
  type HrisDocument,
  listDocuments,
  uploadDocument,
} from '@/lib/hris/documents';

const CATEGORIES: DocumentCategory[] = ['identity', 'contract', 'certificate', 'payslip', 'tax', 'other'];

const SCAN_BADGE: Record<string, string> = {
  clean: 'badge-success',
  pending: 'badge-ghost',
  infected: 'badge-error',
  error: 'badge-warning',
};

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export default function Documents() {
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canUpload = hasPermission('documents.upload');
  const canDelete = hasPermission('documents.delete');

  const [category, setCategory] = useState('');
  const [uploadOpen, setUploadOpen] = useState(false);

  const { data: docs = [], isLoading } = useQuery({
    queryKey: ['hris', 'documents', { category }],
    queryFn: () => listDocuments({ category: category || undefined }),
  });

  const delMutation = useMutation({
    mutationFn: (id: string) => deleteDocument(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hris', 'documents'] }),
  });

  const total = docs.length;
  const clean = docs.filter((d) => d.scanStatus === 'clean').length;
  const flagged = docs.filter((d) => d.scanStatus === 'infected' || d.scanStatus === 'error').length;

  return (
    <div>
      <PageHeader
        title="Documents"
        subtitle="Upload, scan, and manage company and employee documents."
        actions={
          canUpload && (
            <button type="button" className="btn btn-primary btn-sm gap-2" onClick={() => setUploadOpen(true)}>
              <Upload className="h-4 w-4" /> Upload
            </button>
          )
        }
      />

      <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard label="Total documents" value={total} icon={FileText} tone="primary" />
        <StatCard label="Scanned clean" value={clean} icon={ShieldCheck} tone="success" />
        <StatCard label="Flagged" value={flagged} icon={FolderOpen} tone="error" hint="Infected or scan error" />
      </div>

      <div className="rounded-box border border-base-300 bg-base-100">
        <div className="flex flex-wrap items-center gap-3 border-b border-base-300 p-4">
          <select
            className="select select-sm rounded-lg border-base-300"
            aria-label="Filter by category"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
          >
            <option value="">All categories</option>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c.charAt(0).toUpperCase() + c.slice(1)}
              </option>
            ))}
          </select>
        </div>

        {isLoading ? (
          <div className="flex justify-center p-16">
            <span className="loading loading-spinner loading-lg text-primary" />
          </div>
        ) : docs.length === 0 ? (
          <EmptyState
            icon={FolderOpen}
            title="No documents yet"
            description="Upload contracts, certificates, or identity documents — each is malware-scanned on upload."
            action={
              canUpload && (
                <button type="button" className="btn btn-primary btn-sm gap-2" onClick={() => setUploadOpen(true)}>
                  <Upload className="h-4 w-4" /> Upload
                </button>
              )
            }
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr className="text-xs uppercase tracking-wide text-base-content/50">
                  <th>Document</th>
                  <th>Category</th>
                  <th>Employee</th>
                  <th>Size</th>
                  <th>Scan</th>
                  <th>Uploaded</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {docs.map((d: HrisDocument) => (
                  <tr key={d.id} className="hover:bg-base-200/60">
                    <td>
                      <div className="flex items-center gap-2">
                        <FileText className="h-4 w-4 shrink-0 text-base-content/40" />
                        <span className="font-medium">{d.filename}</span>
                      </div>
                    </td>
                    <td className="text-sm capitalize">{d.category}</td>
                    <td className="text-sm">{d.employeeName || '-'}</td>
                    <td className="text-sm text-base-content/70">{formatSize(d.fileSize)}</td>
                    <td>
                      <span className={`badge badge-sm ${SCAN_BADGE[d.scanStatus] ?? 'badge-ghost'}`}>
                        {d.scanStatus}
                      </span>
                    </td>
                    <td className="text-sm text-base-content/60">{d.createdAt.slice(0, 10)}</td>
                    <td>
                      <div className="flex items-center justify-end gap-1">
                        <button
                          type="button"
                          className="btn btn-ghost btn-xs"
                          title="Download"
                          disabled={d.scanStatus === 'infected'}
                          onClick={() => downloadDocument(d.id, d.filename)}
                        >
                          <Download className="h-4 w-4" />
                        </button>
                        {canDelete && (
                          <button
                            type="button"
                            className="btn btn-ghost btn-xs text-error"
                            title="Delete"
                            onClick={() => {
                              if (confirm(`Delete "${d.filename}"?`)) delMutation.mutate(d.id);
                            }}
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {uploadOpen && <UploadModal onClose={() => setUploadOpen(false)} />}
    </div>
  );
}

function UploadModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient();
  const fileRef = useRef<HTMLInputElement>(null);
  const [category, setCategory] = useState<DocumentCategory>('other');
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: () => uploadDocument(file as File, category),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'documents'] });
      onClose();
    },
    onError: (err) => setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Upload failed'),
  });

  return (
    <div className="modal modal-open" role="dialog" aria-modal="true">
      <div className="modal-box">
        <div className="flex items-start justify-between">
          <h3 className="text-lg font-bold flex items-center gap-2">
            <Upload className="h-5 w-5 text-primary" /> Upload document
          </h3>
          <button type="button" className="btn btn-sm btn-ghost btn-circle" onClick={onClose} aria-label="Close">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-4 space-y-4">
          <div>
            <label className="label" htmlFor="doc-category">
              <span className="label-text">Category</span>
            </label>
            <select
              id="doc-category"
              className="select select-bordered w-full"
              value={category}
              onChange={(e) => setCategory(e.target.value as DocumentCategory)}
            >
              {CATEGORIES.map((c) => (
                <option key={c} value={c}>
                  {c.charAt(0).toUpperCase() + c.slice(1)}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="label" htmlFor="doc-file">
              <span className="label-text">File (max 25MB)</span>
            </label>
            <input
              id="doc-file"
              ref={fileRef}
              type="file"
              className="file-input file-input-bordered w-full"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
          </div>

          {error && <div className="alert alert-error text-sm">{error}</div>}
        </div>

        <div className="modal-action">
          <button type="button" className="btn btn-ghost" onClick={onClose} disabled={mutation.isPending}>
            Cancel
          </button>
          <button
            type="button"
            className="btn btn-primary"
            disabled={!file || mutation.isPending}
            onClick={() => mutation.mutate()}
          >
            {mutation.isPending ? 'Uploading…' : 'Upload & scan'}
          </button>
        </div>
      </div>
      <button type="button" className="modal-backdrop" onClick={onClose} aria-label="Close">
        close
      </button>
    </div>
  );
}
