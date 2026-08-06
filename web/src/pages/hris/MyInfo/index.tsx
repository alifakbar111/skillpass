import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { PageHeader } from '@/components/hris/ui';
import { ApiError } from '@/lib/api';
import { type Employee, getMyEmployee, type SelfUpdateRequest, updateMyEmployee } from '@/lib/hris/employees';

const EDITABLE_FIELDS: { key: keyof SelfUpdateRequest; label: string; type?: string }[] = [
  { key: 'phone', label: 'Phone' },
  { key: 'dateOfBirth', label: 'Date of birth', type: 'date' },
  { key: 'gender', label: 'Gender' },
  { key: 'maritalStatus', label: 'Marital status' },
  { key: 'address', label: 'Address' },
  { key: 'city', label: 'City' },
  { key: 'province', label: 'Province' },
  { key: 'postalCode', label: 'Postal code' },
];

const EMERGENCY_FIELDS: { key: keyof SelfUpdateRequest; label: string }[] = [
  { key: 'emergencyContactName', label: 'Contact name' },
  { key: 'emergencyContactPhone', label: 'Contact phone' },
  { key: 'emergencyContactRelation', label: 'Relationship' },
];

function initials(e: Employee) {
  return `${(e.firstName[0] ?? '').toUpperCase()}${(e.lastName?.[0] ?? '').toUpperCase()}`;
}

function Field({ label, value }: { label: string; value?: string | null }) {
  return (
    <div>
      <dt className="text-xs font-medium uppercase tracking-wide text-base-content/50">{label}</dt>
      <dd className="mt-0.5 text-sm text-base-content">{value || '—'}</dd>
    </div>
  );
}

export default function MyInfo() {
  const qc = useQueryClient();
  const { data: emp, isLoading, error } = useQuery({ queryKey: ['hris', 'me', 'employee'], queryFn: getMyEmployee });

  const [editing, setEditing] = useState(false);
  const [form, setForm] = useState<SelfUpdateRequest>({});
  const [saveError, setSaveError] = useState<string | null>(null);

  // Seed the form once when entering edit mode. Seeding from every refetch of
  // `emp` (window-focus refetch, invalidations) clobbered in-progress edits
  // while the user was typing (F11).
  // biome-ignore lint/correctness/useExhaustiveDependencies: intentionally seed only on the editing transition
  useEffect(() => {
    if (!editing || !emp) return;
    setForm({
      phone: emp.phone ?? '',
      dateOfBirth: emp.dateOfBirth?.slice(0, 10) ?? '',
      gender: emp.gender ?? '',
      maritalStatus: emp.maritalStatus ?? '',
      address: emp.address ?? '',
      city: emp.city ?? '',
      province: emp.province ?? '',
      postalCode: emp.postalCode ?? '',
      emergencyContactName: emp.emergencyContactName ?? '',
      emergencyContactPhone: emp.emergencyContactPhone ?? '',
      emergencyContactRelation: emp.emergencyContactRelation ?? '',
    });
  }, [editing]);

  const saveMutation = useMutation({
    mutationFn: () => updateMyEmployee(form),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'me', 'employee'] });
      setEditing(false);
      setSaveError(null);
    },
    onError: (err) => setSaveError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save'),
  });

  if (isLoading) {
    return (
      <div className="flex justify-center p-16">
        <span className="loading loading-spinner loading-lg text-primary" />
      </div>
    );
  }
  if (error || !emp) {
    return (
      <div className="rounded-box border border-base-300 bg-base-100 p-8 text-center text-base-content/60">
        We couldn't load your employee record. If you're not registered as an employee, ask your HR admin.
      </div>
    );
  }

  const set = (key: keyof SelfUpdateRequest, value: string) => setForm((f) => ({ ...f, [key]: value }));

  return (
    <div>
      <PageHeader
        title="My Info"
        subtitle="Your personal, employment and emergency details."
        actions={
          !editing ? (
            <button type="button" className="btn btn-primary btn-sm" onClick={() => setEditing(true)}>
              Edit personal info
            </button>
          ) : (
            <div className="flex gap-2">
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={() => {
                  setEditing(false);
                  setSaveError(null);
                }}
              >
                Cancel
              </button>
              <button
                type="button"
                className="btn btn-primary btn-sm"
                disabled={saveMutation.isPending}
                onClick={() => saveMutation.mutate()}
              >
                {saveMutation.isPending ? 'Saving…' : 'Save changes'}
              </button>
            </div>
          )
        }
      />

      {/* Identity header */}
      <div className="mb-6 flex items-center gap-4 rounded-box border border-base-300 bg-base-100 p-5">
        <span className="grid h-16 w-16 place-items-center rounded-full bg-primary/10 text-xl font-semibold text-primary">
          {initials(emp)}
        </span>
        <div>
          <h2 className="text-xl font-semibold">
            {emp.firstName} {emp.lastName}
          </h2>
          <p className="text-sm text-base-content/60">
            {emp.positionName ?? '—'} · {emp.departmentName ?? '—'}
          </p>
          <p className="mt-1 font-mono text-xs text-base-content/50">{emp.employeeIdNumber}</p>
        </div>
        <span className="badge badge-success badge-sm ml-auto capitalize">
          {emp.employmentStatus.replace('_', ' ')}
        </span>
      </div>

      {saveError && <div className="alert alert-error mb-4 text-sm">{saveError}</div>}

      {/* Employment (read-only) */}
      <div className="mb-6 rounded-box border border-base-300 bg-base-100 p-5">
        <h3 className="mb-4 text-sm font-semibold uppercase tracking-wider text-base-content/50">Employment</h3>
        <dl className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
          <Field label="Email" value={emp.email} />
          <Field label="Department" value={emp.departmentName} />
          <Field label="Position" value={emp.positionName} />
          <Field label="Branch" value={emp.branchName} />
          <Field label="Employment type" value={emp.employmentType?.replace('_', ' ')} />
          <Field label="Join date" value={emp.joinDate ? new Date(emp.joinDate).toLocaleDateString() : undefined} />
        </dl>
        <p className="mt-4 text-xs text-base-content/40">Employment details are managed by HR.</p>
      </div>

      {/* Personal (editable) */}
      <div className="mb-6 rounded-box border border-base-300 bg-base-100 p-5">
        <h3 className="mb-4 text-sm font-semibold uppercase tracking-wider text-base-content/50">
          Personal &amp; Identity
        </h3>
        {editing ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {EDITABLE_FIELDS.map((f) => (
              <label key={f.key} className="form-control">
                <span className="label-text mb-1 text-xs text-base-content/60">{f.label}</span>
                <input
                  type={f.type ?? 'text'}
                  className="input input-sm input-bordered"
                  value={(form[f.key] as string) ?? ''}
                  onChange={(e) => set(f.key, e.target.value)}
                />
              </label>
            ))}
          </div>
        ) : (
          <dl className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            <Field label="Phone" value={emp.phone} />
            <Field label="Date of birth" value={emp.dateOfBirth?.slice(0, 10)} />
            <Field label="Gender" value={emp.gender} />
            <Field label="Marital status" value={emp.maritalStatus} />
            <Field label="National ID (NIK)" value={emp.nationalId} />
            <Field label="NPWP" value={emp.npwp} />
            <Field label="Address" value={emp.address} />
            <Field label="City" value={emp.city} />
            <Field label="Province" value={emp.province} />
            <Field label="Postal code" value={emp.postalCode} />
            <Field label="Bank name" value={emp.bankName} />
            <Field label="Bank account" value={emp.bankAccountNumber} />
          </dl>
        )}
        {editing && (
          <p className="mt-4 text-xs text-base-content/40">
            Bank details and tax identifiers (NIK/NPWP) are managed by HR — contact them to update those fields.
          </p>
        )}
      </div>

      {/* Emergency contact (editable) */}
      <div className="rounded-box border border-base-300 bg-base-100 p-5">
        <h3 className="mb-4 text-sm font-semibold uppercase tracking-wider text-base-content/50">Emergency Contact</h3>
        {editing ? (
          <div className="grid gap-4 sm:grid-cols-3">
            {EMERGENCY_FIELDS.map((f) => (
              <label key={f.key} className="form-control">
                <span className="label-text mb-1 text-xs text-base-content/60">{f.label}</span>
                <input
                  type="text"
                  className="input input-sm input-bordered"
                  value={(form[f.key] as string) ?? ''}
                  onChange={(e) => set(f.key, e.target.value)}
                />
              </label>
            ))}
          </div>
        ) : (
          <dl className="grid gap-5 sm:grid-cols-3">
            <Field label="Contact name" value={emp.emergencyContactName} />
            <Field label="Contact phone" value={emp.emergencyContactPhone} />
            <Field label="Relationship" value={emp.emergencyContactRelation} />
          </dl>
        )}
      </div>
    </div>
  );
}
