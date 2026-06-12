import type { FieldError, UseFormRegisterReturn } from 'react-hook-form';
import { FormField } from './FormField';
import { useFieldBinding } from './useFieldBinding';

interface FormTextareaProps extends Omit<React.TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'> {
  label: string;
  registration?: UseFormRegisterReturn;
  error?: FieldError;
  name?: string;
}

export function FormTextarea({ label, registration, error, name, ...rest }: FormTextareaProps) {
  const {
    registration: resolvedRegistration,
    error: resolvedError,
    fieldName,
  } = useFieldBinding({ name, registration, error });
  return (
    <FormField label={label} name={fieldName} error={resolvedError}>
      <textarea
        id={fieldName}
        className="textarea textarea-bordered w-full"
        {...resolvedRegistration}
        {...rest}
        aria-describedby={resolvedError ? `${fieldName}-error` : undefined}
      />
    </FormField>
  );
}
