import {
  type FieldError,
  get,
  type RegisterOptions,
  type UseFormRegisterReturn,
  useFormContext,
} from 'react-hook-form';

export interface FieldBindingProps {
  name?: string;
  registration?: UseFormRegisterReturn;
  error?: FieldError;
  /** Forwarded to `register` in context mode (e.g. `{ valueAsNumber: true }`). */
  registerOptions?: RegisterOptions;
}

/**
 * Resolves a field's registration + error from either FormProvider context
 * (when `name` is supplied) or explicit `registration`/`error` props.
 * Falls back gracefully when used outside a provider.
 */
export function useFieldBinding({ name, registration, error, registerOptions }: FieldBindingProps): {
  registration?: UseFormRegisterReturn;
  error?: FieldError;
  fieldName: string;
} {
  const ctx = useFormContext();
  if (name && ctx) {
    const err = get(ctx.formState.errors, name) as FieldError | undefined;
    return { registration: ctx.register(name, registerOptions), error: err, fieldName: name };
  }
  return {
    registration,
    error: registration ? error : undefined,
    fieldName: registration?.name ?? name ?? '',
  };
}
