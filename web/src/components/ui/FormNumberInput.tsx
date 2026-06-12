import type { FieldError, UseFormRegisterReturn } from 'react-hook-form';
import { FormField } from './FormField';
import { useFieldBinding } from './useFieldBinding';

interface FormNumberInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'size' | 'type'> {
  label: string;
  registration?: UseFormRegisterReturn;
  error?: FieldError;
  name?: string;
}

/**
 * Number field. Native number inputs emit strings, so in context mode it registers
 * with `setValueAs` to hand the form a real `number` (empty/invalid -> undefined,
 * which satisfies optional number schemas).
 */
export function FormNumberInput({ label, registration, error, name, ...rest }: FormNumberInputProps) {
  const {
    registration: resolvedRegistration,
    error: resolvedError,
    fieldName,
  } = useFieldBinding({
    name,
    registration,
    error,
    registerOptions: {
      setValueAs: (value) => {
        if (value === '' || value === null || value === undefined) return undefined;
        const parsed = Number(value);
        return Number.isNaN(parsed) ? undefined : parsed;
      },
    },
  });
  return (
    <FormField label={label} name={fieldName} error={resolvedError}>
      <input
        id={fieldName}
        type="number"
        className="input input-bordered w-full"
        {...resolvedRegistration}
        {...rest}
        aria-describedby={resolvedError ? `${fieldName}-error` : undefined}
      />
    </FormField>
  );
}
