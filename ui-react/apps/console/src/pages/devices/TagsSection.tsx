import { useState } from "react";
import { TagIcon, XMarkIcon, PlusIcon } from "@heroicons/react/24/outline";
import { useAddDeviceTag, useRemoveDeviceTag } from "@/hooks/useDeviceMutations";
import { useHasPermission } from "@/hooks/useHasPermission";

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";

interface TagsSectionProps {
  uid: string;
  tags: string[];
}

export default function TagsSection({ uid, tags }: TagsSectionProps) {
  const addTagMutation = useAddDeviceTag();
  const removeTagMutation = useRemoveDeviceTag();
  const canEditTags = useHasPermission("tag:edit");
  const [input, setInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAdd = async () => {
    const tag = input.trim();
    if (!tag) return;
    setError(null);

    if (tags && tags.includes(tag)) {
      setError("This tag is already added.");
      return;
    }
    if (tags && tags.length >= 3) return;
    if (tag.length < 3) {
      setError("Tag must be at least 3 characters.");
      return;
    }
    if (tag.length > 255) {
      setError("Tag must be at most 255 characters.");
      return;
    }
    if (!/^[a-zA-Z0-9]+$/.test(tag)) {
      setError("Tag must contain only letters and numbers.");
      return;
    }

    setAdding(true);
    try {
      await addTagMutation.mutateAsync({ path: { uid, name: tag } });
      setInput("");
    } catch {
      setError("Failed to add tag.");
    }
    setAdding(false);
  };

  const handleRemove = async (tag: string) => {
    try {
      await removeTagMutation.mutateAsync({ path: { uid, name: tag } });
    } catch {
      /* invalidation handles UI update */
    }
  };

  return (
    <div>
      <h3 className={LABEL + " mb-2"}>Tags</h3>
      <div className="flex flex-wrap items-center gap-2">
        {tags &&
          tags.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary text-xs rounded-md font-medium"
            >
              <TagIcon className="w-3 h-3" strokeWidth={2} />
              {tag}
              {canEditTags && (
                <button
                  type="button"
                  onClick={() => void handleRemove(tag)}
                  aria-label={`Remove tag ${tag}`}
                  className="hover:text-white transition-colors"
                >
                  <XMarkIcon className="w-3 h-3" strokeWidth={2} />
                </button>
              )}
            </span>
          ))}
        {canEditTags && (!tags || tags.length < 3) && (
          <div className="flex items-center gap-1.5">
            <input
              type="text"
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                setError(null);
              }}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  void handleAdd();
                }
              }}
              placeholder="Add tag..."
              aria-label="New tag"
              className="w-28 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
            />
            <button
              type="button"
              onClick={() => void handleAdd()}
              disabled={adding || !input.trim()}
              aria-label="Add tag"
              className="p-1 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 disabled:opacity-soft transition-all"
            >
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            </button>
          </div>
        )}
      </div>
      {tags && tags.length >= 3 && (
        <p className="text-2xs text-text-muted mt-1.5">
          Maximum of 3 tags reached.
        </p>
      )}
      {error && <p className="text-2xs text-accent-red mt-1.5">{error}</p>}
    </div>
  );
}
