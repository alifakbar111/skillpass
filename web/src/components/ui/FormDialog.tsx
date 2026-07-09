import { type ReactNode, useEffect, useRef } from 'react';

interface Props {
  open: boolean;
  title: string;
  children: ReactNode;
  onClose: () => void;
  /** Rendered in the dialog footer alongside the Cancel button. */
  actions?: ReactNode;
}

export function FormDialog({ open, title, children, onClose, actions }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    if (open) {
      el.showModal();
    } else {
      el.close();
    }
  }, [open]);

  return (
    <dialog ref={dialogRef} className="modal" onClose={onClose}>
      <div className="modal-box max-w-lg">
        <h3 className="font-bold text-lg mb-4">{title}</h3>
        {children}
        <div className="modal-action flex gap-2">
          {actions}
          <button type="button" className="btn btn-ghost btn-sm" onClick={onClose}>
            Cancel
          </button>
        </div>
      </div>
      <form method="dialog" className="modal-backdrop">
        <button type="button" aria-label="Close" onClick={onClose}>
          close
        </button>
      </form>
    </dialog>
  );
}
