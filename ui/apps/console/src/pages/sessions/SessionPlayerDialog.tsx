import { XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import SessionPlayer from "@/components/sessions/SessionPlayer";

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
      <IconButton
        onClick={onClose}
        aria-label="Close"
        className="absolute top-2 right-2 z-10 bg-black/40 hover:bg-black/60"
      >
        <XMarkIcon className="w-4 h-4" strokeWidth={2} />
      </IconButton>

      {/* Player */}
      <div className="flex-1 min-h-0">
        <SessionPlayer logs={logs} onClose={onClose} />
      </div>
    </div>
  );
}
