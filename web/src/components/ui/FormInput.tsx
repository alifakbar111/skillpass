import type { FieldError, UseFormRegisterReturn } from 'react-hook-form';
import { FormField } from './FormField';
import { useFieldBinding } from './useFieldBinding';

interface FormInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size' | 'type'> {
  label: string;
  registration?: UseFormRegisterReturn;
  error?: FieldError;
  name?: string;
  // Text-style inputs only; use FormNumberInput for numeric fields.
  type?: 'text' | 'email' | 'password' | 'url' | 'tel' | 'search' | 'date';
}

export function FormInput({ label, registration, error, name, ...rest }: FormInputProps) {
  const {
    registration: resolvedRegistration,
    error: resolvedError,
    fieldName,
  } = useFieldBinding({ name, registration, error });
  return (
    <FormField label={label} name={fieldName} error={resolvedError}>
      <input
        id={fieldName}
        className="input input-bordered w-full"
        {...resolvedRegistration}
        {...rest}
        aria-describedby={resolvedError ? `${fieldName}-error` : undefined}
      />
    </FormField>
  );
}
