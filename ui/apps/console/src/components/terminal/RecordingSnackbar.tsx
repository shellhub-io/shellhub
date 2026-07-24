import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { CheckCircleIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { useRecordingsStore } from "@/stores/recordingsStore";

// Lightweight post-session notice. Driven by the recordings store, so it shows
// the same way regardless of how the session ended (X button, exit, or dropped
// connection). Non-blocking, announces politely, and auto-dismisses.
export default function RecordingSnackbar() {
  const notice = useRecordingsStore((s) => s.notice);
  const clearNotice = useRecordingsStore((s) => s.clearNotice);
  const navigate = useNavigate();

  useEffect(() => {
    if (!notice) return;
    const t = setTimeout(() => clearNotice(), 6000);
    return () => clearTimeout(t);
  }, [notice, clearNotice]);

  if (!notice) return null;

  return (
    <div
      role="status"
      aria-live="polite"
      className="fixed bottom-16 right-4 z-toast w-80 max-w-[calc(100vw-2rem)] bg-card border border-accent-green/30 rounded-lg shadow-lg shadow-black/30 px-4 py-3 flex items-start gap-3 animate-slide-up"
    >
      <CheckCircleIcon
        className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
        strokeWidth={1.5}
      />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-text-primary">
          Session recorded
        </p>
        <p className="text-2xs text-text-muted mt-0.5 truncate">
          {notice.deviceName}
        </p>
        <button
          type="button"
          onClick={() => {
            clearNotice();
            void navigate("/sessions");
          }}
          className="mt-1.5 text-sm text-primary hover:text-primary-600 font-medium transition-colors"
        >
          View recordings
        </button>
      </div>
      <button
        type="button"
        aria-label="Dismiss"
        onClick={clearNotice}
        className="text-text-muted hover:text-text-primary transition-colors p-0.5 shrink-0"
      >
        <XMarkIcon className="w-3.5 h-3.5" />
      </button>
    </div>
  );
}
