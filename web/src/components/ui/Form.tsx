import { type FieldValues, FormProvider, type UseFormReturn } from 'react-hook-form';

interface FormProps<T extends FieldValues> {
  methods: UseFormReturn<T>;
  onSubmit: (data: T) => void | Promise<void>;
  children: React.ReactNode;
  className?: string;
  'aria-label'?: string;
}

export function Form<T extends FieldValues>({
  methods,
  onSubmit,
  children,
  className,
  'aria-label': ariaLabel,
}: FormProps<T>) {
  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)} className={className} aria-label={ariaLabel}>
        {children}
      </form>
    </FormProvider>
  );
}
