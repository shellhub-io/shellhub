import { useEffect, useState, FormEvent } from "react";
import { useTagsStore } from "../stores/tagsStore";
import axios from "axios";
import Drawer from "./common/Drawer";
import ConfirmDialog from "./common/ConfirmDialog";
import {
  TagIcon,
  PlusIcon,
  XMarkIcon,
  PencilSquareIcon,
  TrashIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";

const TAG_PATTERN = /^[a-zA-Z0-9\-_]+$/;

export default function ManageTagsDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const { tags, loading, fetch, create, update, remove } = useTagsStore();
  const [newName, setNewName] = useState("");
  const [editingTag, setEditingTag] = useState<string | null>(null);
  const [editName, setEditName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [deletingTag, setDeletingTag] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      fetch(1, 100);
      setNewName("");
      setEditingTag(null);
      setError(null);
    }
  }, [open, fetch]);

  const newNameValid =
    newName.trim().length >= 3 && TAG_PATTERN.test(newName.trim());

  const handleCreate = async (e: FormEvent) => {
    e.preventDefault();
    if (!newNameValid) return;
    setSubmitting(true);
    setError(null);
    try {
      await create(newName.trim());
      setNewName("");
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError("A tag with this name already exists.");
      } else {
        setError("Failed to create tag.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleRename = async (currentName: string) => {
    const trimmed = editName.trim();
    if (
      !trimmed ||
      trimmed.length < 3 ||
      !TAG_PATTERN.test(trimmed) ||
      trimmed === currentName
    ) {
      setEditingTag(null);
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await update(currentName, trimmed);
      setEditingTag(null);
    } catch {
      setError(`Failed to rename "${currentName}".`);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (name: string) => {
    setSubmitting(true);
    setError(null);
    try {
      await remove(name);
      setDeletingTag(null);
    } catch {
      setError(`Failed to delete "${name}".`);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      <Drawer
        open={open}
        onClose={onClose}
        title="Manage Tags"
        subtitle={`${tags.length} tag${tags.length !== 1 ? "s" : ""}`}
        icon={<TagIcon className="w-4 h-4 text-primary" />}
        width="sm"
        bodyClassName="flex-1 flex flex-col overflow-hidden"
        footer={
          <button
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Done
          </button>
        }
      >
        {/* Create input */}
        <form
          onSubmit={handleCreate}
          className="px-6 py-3 border-b border-border shrink-0"
        >
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Enter tag name..."
              className="flex-1 px-3 py-2 bg-card border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
            />
            <button
              type="submit"
              disabled={!newNameValid || submitting}
              className="px-3 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all shrink-0"
            >
              {submitting ? (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin block" />
              ) : (
                <PlusIcon className="w-4 h-4" strokeWidth={2} />
              )}
            </button>
          </div>
          {newName.trim() && !newNameValid && (
            <p className="mt-1.5 text-2xs text-text-muted">
              {newName.trim().length < 3
                ? "At least 3 characters"
                : "Only letters, numbers, hyphens, underscores"}
            </p>
          )}
        </form>

        {/* Error */}
        {error && (
          <div className="px-6 py-2 shrink-0">
            <div className="flex items-center justify-between gap-2">
              <p className="text-xs text-accent-red flex items-center gap-1.5">
                <ExclamationCircleIcon
                  className="w-3.5 h-3.5 shrink-0"
                  strokeWidth={2}
                />
                {error}
              </p>
              <button
                onClick={() => setError(null)}
                className="p-0.5 rounded text-accent-red/60 hover:text-accent-red transition-colors shrink-0"
              >
                <XMarkIcon className="w-3.5 h-3.5" strokeWidth={2} />
              </button>
            </div>
          </div>
        )}

        {/* Tag list */}
        <div className="flex-1 overflow-y-auto">
          {loading && tags.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
            </div>
          ) : tags.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <TagIcon className="w-8 h-8 text-text-muted/30 mx-auto mb-3" />
              <p className="text-sm text-text-muted">No tags yet</p>
              <p className="text-2xs text-text-muted/60 mt-1">
                Create your first tag above.
              </p>
            </div>
          ) : (
            <div className="divide-y divide-border/60">
              {tags.map((tag) => (
                <div
                  key={tag.name}
                  className="group flex items-center gap-2 px-6 py-2.5 hover:bg-hover-subtle transition-colors"
                >
                  {editingTag === tag.name ? (
                    <input
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") handleRename(tag.name);
                        if (e.key === "Escape") setEditingTag(null);
                      }}
                      onBlur={() => handleRename(tag.name)}
                      autoFocus
                      className="flex-1 px-2.5 py-1 bg-card border border-primary/50 rounded-md text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-primary/20 transition-all"
                    />
                  ) : (
                    <div className="flex-1 flex items-center gap-2 min-w-0">
                      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-md font-medium">
                        <TagIcon className="w-3 h-3" strokeWidth={2} />
                        {tag.name}
                      </span>
                    </div>
                  )}
                  {editingTag !== tag.name && (
                    <div className="flex items-center gap-0.5 shrink-0">
                      <button
                        onClick={() => {
                          setEditingTag(tag.name);
                          setEditName(tag.name);
                        }}
                        className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
                        title="Rename"
                      >
                        <PencilSquareIcon className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={() => setDeletingTag(tag.name)}
                        className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
                        title="Delete"
                      >
                        <TrashIcon className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </Drawer>

      <ConfirmDialog
        open={!!deletingTag}
        onClose={() => setDeletingTag(null)}
        onConfirm={() => handleDelete(deletingTag!)}
        title="Delete Tag"
        description={
          <>
            Are you sure you want to delete{" "}
            <span className="font-medium text-text-primary">{deletingTag}</span>
            ? This will remove the tag from all devices.
          </>
        }
        confirmLabel="Delete"
      />
    </>
  );
}
