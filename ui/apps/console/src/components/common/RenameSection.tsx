import { useState } from "react";
import {
  PencilSquareIcon,
  CheckIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";

export interface RenameSectionProps {
  uid: string;
  currentName: string;
  rename: (opts: {
    path: { uid: string };
    body: { name: string };
  }) => Promise<unknown>;
  entityLabel: string;
  canRename?: boolean;
}

export default function RenameSection({
  uid,
  currentName,
  rename,
  entityLabel,
  canRename = true,
}: RenameSectionProps) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(currentName);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    if (saving) return;
    if (!name.trim() || name.trim() === currentName) {
      setEditing(false);
      return;
    }
    setSaving(true);
    setError(null);
    try {
      await rename({
        path: { uid },
        body: { name: name.trim() },
      });
      setEditing(false);
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      const errors: Record<number, string> = {
        400: `Invalid ${entityLabel} name.`,
        409: `A ${entityLabel} with that name already exists.`,
      };
      setError((status && errors[status]) || `Failed to rename ${entityLabel}.`);
    }
    setSaving(false);
  };

  if (!editing) {
    return (
      <div className="flex items-center gap-2">
        <h1 className="text-2xl font-bold text-text-primary">{currentName}</h1>
        {canRename && (
          <IconButton
            variant="primary"
            aria-label={`Rename ${entityLabel}`}
            title={`Rename ${entityLabel}`}
            onClick={() => {
              setName(currentName);
              setEditing(true);
            }}
          >
            <PencilSquareIcon className="w-4 h-4" />
          </IconButton>
        )}
      </div>
    );
  }

  return (
    <>
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") void handleSave();
            if (e.key === "Escape") setEditing(false);
          }}
          aria-label={`${entityLabel.charAt(0).toUpperCase() + entityLabel.slice(1)} name`}
          className="text-2xl font-bold text-text-primary bg-transparent border-b-2 border-primary/50 focus:outline-none focus:border-primary w-full max-w-md"
        />
        <IconButton
          variant="primary"
          type="button"
          aria-label={`Save ${entityLabel} name`}
          disabled={saving}
          onClick={() => void handleSave()}
        >
          <CheckIcon className="w-4 h-4" strokeWidth={2} />
        </IconButton>
        <IconButton
          type="button"
          aria-label="Cancel rename"
          onClick={() => setEditing(false)}
        >
          <XMarkIcon className="w-4 h-4" strokeWidth={2} />
        </IconButton>
      </div>
      {error && (
        <p role="alert" className="text-2xs text-accent-red mt-1">
          {error}
        </p>
      )}
    </>
  );
}
