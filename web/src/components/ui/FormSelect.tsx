import type { FieldError, UseFormRegisterReturn } from 'react-hook-form';
import { FormField } from './FormField';
import { useFieldBinding } from './useFieldBinding';

interface FormSelectProps extends Omit<React.SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  label: string;
  registration?: UseFormRegisterReturn;
  error?: FieldError;
  name?: string;
  options: { value: string; label: string }[];
  placeholder?: string;
}

export function FormSelect({ label, registration, error, name, options, placeholder, ...rest }: FormSelectProps) {
  const {
    registration: resolvedRegistration,
    error: resolvedError,
    fieldName,
  } = useFieldBinding({ name, registration, error });
  return (
    <FormField label={label} name={fieldName} error={resolvedError}>
      <select
        id={fieldName}
        className="select select-bordered w-full"
        {...resolvedRegistration}
        {...rest}
        aria-describedby={resolvedError ? `${fieldName}-error` : undefined}
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
