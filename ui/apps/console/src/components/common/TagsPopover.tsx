import { useId, useState } from "react";
import {
  TagIcon,
  XMarkIcon,
  PlusIcon,
  PencilIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useTags } from "@/hooks/useTagRequests";
import { useHasPermission } from "@/hooks/useHasPermission";
import { tagNameProblem } from "@/utils/tagName";
import {
  ContextualDialogHeader,
  KeyHint,
} from "@/components/common/ContextualDialog";

interface TagsPopoverProps {
  uid: string;
  tags: string[];
  addTag: (opts: { path: { uid: string; name: string } }) => Promise<unknown>;
  removeTag: (opts: {
    path: { uid: string; name: string };
  }) => Promise<unknown>;
  onFilterTag: (tag: string) => void;
  editLabel?: string;
}

/**
 * Adds and removes an entity's tags in place, from a row, without opening its detail page.
 */
export default function TagsPopover({
  uid,
  tags: entityTags,
  addTag,
  removeTag,
  onFilterTag,
  editLabel = "Manage tags",
}: TagsPopoverProps) {
  const { tags: tagObjects } = useTags();
  const allTags = tagObjects.map((t) => t.name);

  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const titleId = useId();
  const descriptionId = useId();
  const canEditTags = useHasPermission(["TagCreate", "TagDelete"]);
  const tags = entityTags || [];

  const handleAdd = async (tag: string) => {
    if (
      tags.includes(tag) ||
      tags.length >= 3 ||
      tagNameProblem(tag)
    )
      return;
    setLoading(true);
    setError(null);
    try {
      await addTag({ path: { uid, name: tag } });
      setInput("");
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 403) setError("You don't have permission to add tags.");
      else if (status === 400) setError(`"${tag}" is not a valid tag name.`);
      else setError(`Failed to add "${tag}".`);
    }
    setLoading(false);
  };

  const handleRemove = async (tag: string) => {
    setLoading(true);
    setError(null);
    try {
      await removeTag({ path: { uid, name: tag } });
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 403) setError("You don't have permission to remove tags.");
      else setError(`Failed to remove "${tag}".`);
    }
    setLoading(false);
  };

  const suggestions = allTags.filter(
    (t) => !tags.includes(t) && t.toLowerCase().includes(input.toLowerCase()),
  );
  const inputProblem = input.trim() ? tagNameProblem(input.trim()) : null;
  const isNew =
    !inputProblem &&
    !allTags.includes(input.trim()) &&
    !tags.includes(input.trim());
  const inputMalformed = !!inputProblem && !inputProblem.incomplete;

  return (
    <>
      <div className="flex items-center gap-1 min-h-[28px] group/tags">
        {tags.length > 0 ? (
          <div className="flex flex-wrap items-center gap-1">
            {tags.map((tag) => (
              <button
                type="button"
                key={tag}
                onClick={(e) => {
                  e.stopPropagation();
                  onFilterTag(tag);
                }}
                title={`Filter by "${tag}"`}
                className="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-primary/10 text-primary text-2xs rounded font-medium hover:bg-primary/20 transition-all cursor-pointer"
              >
                <TagIcon className="w-2 h-2" strokeWidth={2} />
                {tag}
              </button>
            ))}
          </div>
        ) : (
          <span className="text-2xs text-text-muted/30 group-hover/tags:text-text-muted transition-colors">
            No tags
          </span>
        )}
        {canEditTags && (
          <span
            role="presentation"
            onClick={(e) => e.stopPropagation()}
          >
            <Dropdown
              mode="content"
              portal
              onOpenChange={(open) => {
                if (open) {
                  setInput("");
                  setError(null);
                }
              }}
            >
              <Dropdown.Trigger>
                <button
                  type="button"
                  className="p-0.5 rounded text-text-muted/20 group-hover/tags:text-text-muted hover:text-primary hover:bg-primary/10 transition-all shrink-0"
                  title={editLabel}
                  aria-label={editLabel}
                >
                  <PencilIcon className="w-3 h-3" strokeWidth={2} />
                </button>
              </Dropdown.Trigger>

              <Dropdown.Panel
                aria-labelledby={titleId}
                aria-describedby={descriptionId}
                className="w-[300px] bg-card border-border-light"
              >
                <ContextualDialogHeader
                  icon={<TagIcon />}
                  title="Manage tags"
                  description="Up to three tags. Click one in the list to filter by it."
                  titleId={titleId}
                  descriptionId={descriptionId}
                />
                <div className="px-3.5 py-3 space-y-3">
                  {tags.length < 3 ? (
                    <div>
                      <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={(e) => {
                          if (
                            e.key === "Enter" &&
                            input.trim() &&
                            !inputProblem
                          ) {
                            e.preventDefault();
                            void handleAdd(input.trim());
                          }
                        }}
                        placeholder="Search or create tag..."
                        aria-label="Search or create tag"
                        className="w-full px-2.5 py-1.5 bg-card border border-border rounded-lg text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                      />
                      {inputProblem && (
                        <p
                          className={cn(
                            "text-2xs mt-1",
                            inputProblem.incomplete
                              ? "text-text-muted"
                              : "text-accent-red",
                          )}
                        >
                          {inputProblem.hint}
                        </p>
                      )}

                      {(suggestions.length > 0 || isNew) &&
                        input.trim() &&
                        !inputMalformed && (
                          <div className="mt-1.5 max-h-[140px] overflow-y-auto border border-border rounded-lg divide-y divide-border/60">
                            {suggestions.map((tag) => (
                              <button
                                type="button"
                                key={tag}
                                onClick={() => void handleAdd(tag)}
                                disabled={loading}
                                className="w-full text-left px-2.5 py-1.5 text-2xs text-text-primary hover:bg-hover-medium transition-colors disabled:opacity-dim flex items-center gap-1.5"
                              >
                                <TagIcon
                                  className="w-2.5 h-2.5 text-primary shrink-0"
                                  strokeWidth={2}
                                />
                                {tag}
                              </button>
                            ))}
                            {isNew && (
                              <button
                                type="button"
                                onClick={() => void handleAdd(input.trim())}
                                disabled={loading}
                                className="w-full text-left px-2.5 py-1.5 text-2xs text-accent-green hover:bg-hover-medium transition-colors disabled:opacity-dim flex items-center gap-1.5"
                              >
                                <PlusIcon
                                  className="w-2.5 h-2.5 shrink-0"
                                  strokeWidth={2}
                                />
                                Create &ldquo;
                                {input.trim()}
                                &rdquo;
                              </button>
                            )}
                          </div>
                        )}
                    </div>
                  ) : (
                    <p className="text-2xs text-text-muted">
                      Max 3 tags. Remove one to add another.
                    </p>
                  )}

                  {tags.length > 0 && (
                    <div className="flex flex-wrap gap-1.5">
                      {tags.map((tag) => (
                        <span
                          key={tag}
                          className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/10 text-primary text-2xs rounded-md font-medium"
                        >
                          <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
                          {tag}
                          <button
                            type="button"
                            onClick={() => void handleRemove(tag)}
                            disabled={loading}
                            aria-label={`Remove tag ${tag}`}
                            className="hover:text-white transition-colors disabled:opacity-dim ml-0.5"
                          >
                            <XMarkIcon
                              className="w-2.5 h-2.5"
                              strokeWidth={2}
                            />
                          </button>
                        </span>
                      ))}
                    </div>
                  )}

                  {error && (
                    <p role="alert" className="text-2xs text-accent-red">
                      {error}
                    </p>
                  )}
                </div>
                {tags.length < 3 && (
                  <div className="px-3.5 py-2.5 border-t border-border">
                    <KeyHint action="add" />
                  </div>
                )}
              </Dropdown.Panel>
            </Dropdown>
          </span>
        )}
      </div>
    </>
  );
}
