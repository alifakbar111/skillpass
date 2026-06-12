import { useController, useFormContext } from 'react-hook-form';

interface ToggleButtonGroupProps {
  name: string;
  legend: string;
  options: { value: string; label: string }[];
  'aria-label'?: string;
}

export function ToggleButtonGroup({ name, legend, options, 'aria-label': ariaLabel }: ToggleButtonGroupProps) {
  const { control } = useFormContext();
  const { field } = useController({ name, control });

  return (
    <fieldset className="fieldset" aria-label={ariaLabel}>
      <legend className="fieldset-legend">{legend}</legend>
      <div className="flex gap-2">
        {options.map((opt) => (
          <button
            key={opt.value}
            type="button"
            className={`btn flex-1 ${field.value === opt.value ? 'btn-primary' : 'btn-outline'}`}
            onClick={() => field.onChange(opt.value)}
            // aria-pressed={field.value === opt.value}
            aria-label={opt.label}
          >
            {opt.label}
          </button>
        ))}
      </div>
    </fieldset>
  );
}
