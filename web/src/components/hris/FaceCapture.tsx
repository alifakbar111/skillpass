import { Camera, RefreshCw } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

interface Props {
  /** Called with a JPEG data URL when the user captures a frame. */
  onCapture: (dataUrl: string) => void;
  /** The most recently captured image, if any (controlled by the parent). */
  captured?: string | null;
}

/**
 * Reusable webcam capture using the MediaDevices API. Streams the camera,
 * captures a still frame to a canvas, and returns it as a JPEG data URL.
 * Requires a secure context (https or localhost).
 */
export function FaceCapture({ onCapture, captured }: Props) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    let cancelled = false;
    async function start() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'user' }, audio: false });
        if (cancelled) {
          for (const t of stream.getTracks()) t.stop();
          return;
        }
        streamRef.current = stream;
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
          setReady(true);
        }
      } catch {
        setError('Camera access was denied or is unavailable. Allow camera permission and reload.');
      }
    }
    start();
    return () => {
      cancelled = true;
      if (streamRef.current) {
        for (const t of streamRef.current.getTracks()) t.stop();
      }
    };
  }, []);

  function capture() {
    const video = videoRef.current;
    if (!video) return;
    const canvas = document.createElement('canvas');
    canvas.width = video.videoWidth || 480;
    canvas.height = video.videoHeight || 360;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
    onCapture(canvas.toDataURL('image/jpeg', 0.85));
  }

  if (error) {
    return <div className="alert alert-error text-sm">{error}</div>;
  }

  return (
    <div className="flex flex-col items-center gap-3">
      <div className="relative w-full max-w-sm overflow-hidden rounded-box border border-base-300 bg-base-200 aspect-[4/3]">
        {captured ? (
          // biome-ignore lint/a11y/useAltText: preview of the just-captured frame
          <img src={captured} alt="Captured face" className="h-full w-full object-cover" />
        ) : (
          <>
            {/* biome-ignore lint/a11y/useMediaCaption: live camera preview, no captions */}
            <video ref={videoRef} autoPlay muted playsInline className="h-full w-full object-cover" />
            <div className="pointer-events-none absolute inset-0 m-auto h-2/3 w-1/2 rounded-full border-2 border-dashed border-primary/50" />
          </>
        )}
      </div>

      {captured ? (
        <button type="button" className="btn btn-outline btn-sm gap-2" onClick={() => onCapture('')}>
          <RefreshCw className="h-4 w-4" /> Retake
        </button>
      ) : (
        <button type="button" className="btn btn-primary btn-sm gap-2" disabled={!ready} onClick={capture}>
          <Camera className="h-4 w-4" /> Capture
        </button>
      )}
    </div>
  );
}
