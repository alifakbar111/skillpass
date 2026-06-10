import { Check, Copy, Linkedin, MessageCircle, Printer, Share2, Twitter } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';
import { type RefObject, useRef, useState } from 'react';
import { useReactToPrint } from 'react-to-print';

interface Props {
  /** Public passport slug (username). */
  slug: string;
  name: string;
  /** Ref to the passport content to print/save as PDF. */
  printRef: RefObject<HTMLDivElement | null>;
}

// SharePassport: the passport is SkillPass's viral artifact — make sharing
// one click. QR for in-person, intents for social, print-to-PDF for email.
export function SharePassport({ slug, name, printRef }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [copied, setCopied] = useState(false);

  const url = `${window.location.origin}/profiles/${slug}`;
  const shareText = `Check out ${name}'s skill passport on SkillPass`;

  const handlePrint = useReactToPrint({
    contentRef: printRef,
    documentTitle: `${name} — SkillPass`,
  });

  async function copyLink() {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard unavailable (http, permissions) — the input below remains selectable.
    }
  }

  const intents = [
    {
      label: 'LinkedIn',
      icon: Linkedin,
      href: `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(url)}`,
    },
    {
      label: 'X',
      icon: Twitter,
      href: `https://twitter.com/intent/tweet?url=${encodeURIComponent(url)}&text=${encodeURIComponent(shareText)}`,
    },
    {
      label: 'WhatsApp',
      icon: MessageCircle,
      href: `https://wa.me/?text=${encodeURIComponent(`${shareText} ${url}`)}`,
    },
  ];

  return (
    <>
      <div className="flex gap-2">
        <button type="button" className="btn btn-primary btn-sm gap-2" onClick={() => dialogRef.current?.showModal()}>
          <Share2 size={14} aria-hidden="true" /> Share
        </button>
        <button type="button" className="btn btn-outline btn-sm gap-2" onClick={handlePrint}>
          <Printer size={14} aria-hidden="true" /> Save PDF
        </button>
      </div>

      <dialog ref={dialogRef} className="modal">
        <div className="modal-box max-w-sm text-center">
          <h3 className="font-bold text-lg mb-4">Share this passport</h3>

          <div className="bg-white rounded-box p-4 inline-block mb-4">
            <QRCodeSVG value={url} size={160} aria-label={`QR code linking to ${name}'s passport`} />
          </div>

          <div className="join w-full mb-4">
            <input type="text" readOnly value={url} className="input input-bordered input-sm join-item flex-1" />
            <button type="button" className="btn btn-sm btn-primary join-item gap-1" onClick={copyLink}>
              {copied ? <Check size={14} /> : <Copy size={14} />}
              {copied ? 'Copied' : 'Copy'}
            </button>
          </div>

          <div className="flex justify-center gap-2">
            {intents.map(({ label, icon: Icon, href }) => (
              <a
                key={label}
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                className="btn btn-outline btn-sm gap-1"
              >
                <Icon size={14} aria-hidden="true" /> {label}
              </a>
            ))}
          </div>

          <div className="modal-action">
            <form method="dialog">
              <button type="submit" className="btn btn-ghost btn-sm">
                Close
              </button>
            </form>
          </div>
        </div>
        <form method="dialog" className="modal-backdrop">
          <button type="submit" aria-label="Close">
            close
          </button>
        </form>
      </dialog>
    </>
  );
}
