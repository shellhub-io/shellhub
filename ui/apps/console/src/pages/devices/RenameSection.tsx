import { useState } from "react";
import {
  PencilSquareIcon,
  CheckIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { useRenameDevice } from "@/hooks/useDeviceMutations";
import { useHasPermission } from "@/hooks/useHasPermission";

interface RenameSectionProps {
  uid: string;
  currentName: string;
}

export default function RenameSection({
  uid,
  currentName,
}: RenameSectionProps) {
  const renameMutation = useRenameDevice();
  const canRename = useHasPermission("device:rename");
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(currentName);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (!name.trim() || name.trim() === currentName) {
      setEditing(false);
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await renameMutation.mutateAsync({
        path: { uid },
        body: { name: name.trim() },
      });
      setEditing(false);
    } catch {
      setError("Failed to rename device.");
    }
    setSaving(false);
  };

  if (!editing) {
    return (
      <div className="flex items-center gap-2">
        <h1 className="text-2xl font-bold text-text-primary">{currentName}</h1>
        {canRename && (
          <button
            type="button"
            onClick={() => {
              setName(currentName);
              setEditing(true);
            }}
            className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
            aria-label="Rename device"
            title="Rename"
          >
            <PencilSquareIcon className="w-4 h-4" />
          </button>
        )}
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void handleSave();
            if (e.key === "Escape") setEditing(false);
          }}
          aria-label="Device name"

          className="text-2xl font-bold text-text-primary bg-transparent border-b-2 border-primary/50 focus:outline-none focus:border-primary w-full max-w-md"
        />
        <button
          type="button"
          onClick={() => void handleSave()}
          disabled={saving}
          aria-label="Save device name"
          className="p-1.5 rounded-md text-accent-green hover:bg-accent-green/10 transition-all"
        >
          <CheckIcon className="w-4 h-4" strokeWidth={2} />
        </button>
        <button
          type="button"
          onClick={() => setEditing(false)}
          aria-label="Cancel rename"
          className="p-1.5 rounded-md text-text-muted hover:bg-hover-medium transition-all"
        >
          <XMarkIcon className="w-4 h-4" strokeWidth={2} />
        </button>
      </div>
      {error && <p className="text-2xs text-accent-red mt-1">{error}</p>}
    </div>
  );
}
