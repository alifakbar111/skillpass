import type { FieldError, UseFormRegisterReturn } from 'react-hook-form';

interface FormFieldProps {
  label: string;
  name: string;
  error?: FieldError;
  children: React.ReactNode;
}

export function FormField({ label, name, error, children }: FormFieldProps) {
  return (
    <div className="form-control w-full">
      <label className="label" htmlFor={name}>
        <span className="label-text">{label}</span>
      </label>
      {children}
      {error && (
        <span className="text-error text-xs mt-1" role="alert" id={`${name}-error`}>
          {error.message}
        </span>
      )}
    </div>
  );
}

interface FormInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size'> {
  label: string;
  registration: UseFormRegisterReturn;
  error?: FieldError;
}

export function FormInput({ label, registration, error, ...rest }: FormInputProps) {
  return (
    <FormField label={label} name={registration.name} error={error}>
      <input
        id={registration.name}
        className="input input-bordered w-full"
        {...registration}
        {...rest}
        aria-describedby={error ? `${registration.name}-error` : undefined}
      />
    </FormField>
  );
}

interface FormSelectProps extends Omit<React.SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label: string;
  registration: UseFormRegisterReturn;
  error?: FieldError;
  options: { value: string; label: string }[];
  placeholder?: string;
}

export function FormSelect({ label, registration, error, options, placeholder }: FormSelectProps) {
  return (
    <FormField label={label} name={registration.name} error={error}>
      <select
        id={registration.name}
        className="select select-bordered w-full"
        {...registration}
        aria-describedby={error ? `${registration.name}-error` : undefined}
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </FormField>
  );
}

interface FormTextareaProps extends Omit<React.TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'> {
  label: string;
  registration: UseFormRegisterReturn;
  error?: FieldError;
}

export function FormTextarea({ label, registration, error, ...rest }: FormTextareaProps) {
  return (
    <FormField label={label} name={registration.name} error={error}>
      <textarea
        id={registration.name}
        className="textarea textarea-bordered w-full"
        {...registration}
        {...rest}
        aria-describedby={error ? `${registration.name}-error` : undefined}
      />
    </FormField>
  );
}
