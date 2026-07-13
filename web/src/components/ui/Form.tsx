import { type FieldValues, FormProvider, type UseFormReturn } from 'react-hook-form';

interface FormProps<T extends FieldValues> {
  id?: string;
  methods: UseFormReturn<T>;
  onSubmit: (data: T) => void | Promise<void>;
  children: React.ReactNode;
  className?: string;
  'aria-label'?: string;
}

export function Form<T extends FieldValues>({
  id,
  methods,
  onSubmit,
  children,
  className,
  'aria-label': ariaLabel,
}: FormProps<T>) {
  return (
    <FormProvider {...methods}>
      <form id={id} onSubmit={methods.handleSubmit(onSubmit)} className={className} aria-label={ariaLabel}>
        {children}
      </form>
    </FormProvider>
  );
}
