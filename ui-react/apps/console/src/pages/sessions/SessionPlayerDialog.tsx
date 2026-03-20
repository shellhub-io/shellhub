import { XMarkIcon } from "@heroicons/react/24/outline";
import SessionPlayer from "../../components/sessions/SessionPlayer";

interface SessionPlayerDialogProps {
  open: boolean;
  onClose: () => void;
  logs: string;
}

export default function SessionPlayerDialog({
  open,
  onClose,
  logs,
}: SessionPlayerDialogProps) {
  if (!open) return null;

  return (
    <div className="absolute inset-0 z-[80] flex flex-col bg-[#121314]">
      {/* Close button overlay */}
      <button
        onClick={onClose}
        className="absolute top-2 right-2 z-10 w-7 h-7 flex items-center justify-center rounded-md bg-black/40 hover:bg-black/60 text-text-muted hover:text-text-primary transition-colors"
        aria-label="Close"
      >
        <XMarkIcon className="w-4 h-4" strokeWidth={2} />
      </button>

      {/* Player */}
      <div className="flex-1 min-h-0">
        <SessionPlayer logs={logs} onClose={onClose} />
      </div>
    </div>
  );
}
