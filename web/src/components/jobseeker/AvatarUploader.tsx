import { Camera } from 'lucide-react';
import { useRef, useState } from 'react';
import { ApiError, apiUpload } from '../../lib/api';

interface Props {
  name: string;
  avatarUrl?: string | null;
  onUploaded: (url: string) => void;
}

// AvatarUploader shows the current avatar (or an initial placeholder) with a
// camera overlay to upload a new image (png/jpeg/webp, max 2MB).
export function AvatarUploader({ name, avatarUrl, onUploaded }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleFile(file: File) {
    if (uploading) return;
    setUploading(true);
    setError(null);
    try {
      const form = new FormData();
      form.append('file', file);
      const res = await apiUpload<{ avatarUrl: string }>('/profiles/me/avatar', form);
      onUploaded(res.avatarUrl);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Upload failed');
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = '';
    }
  }

  return (
    <div className="flex flex-col items-center gap-1">
      <button
        type="button"
        className="avatar placeholder relative group cursor-pointer"
        onClick={() => inputRef.current?.click()}
        aria-label="Change profile photo"
        disabled={uploading}
      >
        {avatarUrl ? (
          <div className="w-20 rounded-full">
            <img src={avatarUrl} alt={`${name} avatar`} />
          </div>
        ) : (
          <div className="bg-neutral text-neutral-content rounded-full w-20">
            <span className="text-2xl">{name?.charAt(0)}</span>
          </div>
        )}
        <span className="absolute inset-0 rounded-full bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
          {uploading ? (
            <span className="loading loading-spinner loading-sm text-white" />
          ) : (
            <Camera size={20} className="text-white" aria-hidden="true" />
          )}
        </span>
      </button>
      <input
        ref={inputRef}
        type="file"
        accept="image/png,image/jpeg,image/webp"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleFile(file);
        }}
      />
      {error && <p className="text-error text-xs">{error}</p>}
    </div>
  );
}
