import { useState } from "react";
import { PlusIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import {
  useSetDeviceCustomField,
  useDeleteDeviceCustomField,
} from "@/hooks/useDeviceMutations";
import { useHasPermission } from "@/hooks/useHasPermission";

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";

interface CustomFieldsSectionProps {
  uid: string;
  customFields: Record<string, string>;
}

export default function CustomFieldsSection({
  uid,
  customFields,
}: CustomFieldsSectionProps) {
  const setMutation = useSetDeviceCustomField();
  const deleteMutation = useDeleteDeviceCustomField();
  const canEdit = useHasPermission("device:customField:update");
  const [keyInput, setKeyInput] = useState("");
  const [valueInput, setValueInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmKey, setConfirmKey] = useState<string | null>(null);

  const handleAdd = async () => {
    const key = keyInput.trim();
    const value = valueInput.trim();
    if (!key || !value) return;
    if (key in customFields) {
      setError("This key already exists.");
      return;
    }
    setError(null);
    setAdding(true);
    try {
      await setMutation.mutateAsync({
        path: { uid, key },
        body: { value },
      });
      setKeyInput("");
      setValueInput("");
    } catch {
      setError("Failed to add custom field.");
    }
    setAdding(false);
  };

  const handleRemove = async (key: string) => {
    try {
      await deleteMutation.mutateAsync({
        path: { uid, key },
      });
    } catch {
      /* invalidation handles UI update */
    }
  };

  return (
    <div>
      <h3 className={LABEL + " mb-3"}>Custom Fields</h3>
      <dl className="space-y-2 mb-3">
        {Object.entries(customFields).map(([key, value]) => (
          <div key={key} className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <span className="text-xs font-mono text-text-muted shrink-0">
                {key}:
              </span>
              <span className="text-sm text-text-primary font-medium truncate">
                {value}
              </span>
            </div>
            {canEdit &&
              (confirmKey === key ? (
                <div className="flex items-center gap-1 shrink-0">
                  <span className="text-2xs text-text-muted">Remove?</span>
                  <button
                    type="button"
                    onClick={() => {
                      void handleRemove(key);
                      setConfirmKey(null);
                    }}
                    className="px-1.5 py-0.5 rounded text-2xs font-semibold bg-accent-red/90 hover:bg-accent-red text-white transition-all"
                  >
                    Yes
                  </button>
                  <button
                    type="button"
                    onClick={() => setConfirmKey(null)}
                    className="px-1.5 py-0.5 rounded text-2xs font-semibold text-text-muted hover:text-text-primary hover:bg-hover-subtle transition-all"
                  >
                    No
                  </button>
                </div>
              ) : (
                <IconButton
                  variant="danger"
                  size="sm"
                  type="button"
                  aria-label={`Remove custom field ${key}`}
                  onClick={() => setConfirmKey(key)}
                >
                  <XMarkIcon className="w-3 h-3" strokeWidth={2} />
                </IconButton>
              ))}
          </div>
        ))}
      </dl>
      {canEdit && (
        <div className="flex items-center gap-1.5">
          <input
            type="text"
            value={keyInput}
            onChange={(e) => {
              setKeyInput(e.target.value);
              setError(null);
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                void handleAdd();
              }
            }}
            placeholder="key"
            aria-label="Custom field key"
            className="w-24 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
          />
          <span className="text-text-muted text-xs">:</span>
          <input
            type="text"
            value={valueInput}
            onChange={(e) => {
              setValueInput(e.target.value);
              setError(null);
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                void handleAdd();
              }
            }}
            placeholder="value"
            aria-label="Custom field value"
            className="w-32 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
          />
          <IconButton
            variant="primary"
            size="sm"
            disabled={adding || !keyInput.trim() || !valueInput.trim()}
            aria-label="Add custom field"
            onClick={() => void handleAdd()}
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
          </IconButton>
        </div>
      )}
      {error && <p className="text-2xs text-accent-red mt-1.5">{error}</p>}
    </div>
  );
}
