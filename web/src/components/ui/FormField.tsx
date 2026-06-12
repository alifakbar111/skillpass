import type { FieldError } from 'react-hook-form';

interface FormFieldProps {
  label: string;
  name: string;
  error?: FieldError;
  children: React.ReactNode;
}

/**
 * Low-level layout primitive: renders a label, the control, and an error message.
 * Field components (FormInput, FormSelect, FormTextarea) compose this.
 */
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
