import { useState } from "react";
import {
  TagIcon,
  ChevronDownIcon,
  CheckIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown } from "@shellhub/design-system/primitives";
import { useTags } from "@/hooks/useTags";

function TagFilterDropdown({
  filterTags,
  onAdd,
  onRemove,
  onClearAll,
  onManageTags,
}: {
  filterTags: string[];
  onAdd: (tag: string) => void;
  onRemove: (tag: string) => void;
  onClearAll: () => void;
  onManageTags?: () => void;
}) {
  const { tags: tagObjects } = useTags();
  const allTags = tagObjects.map((t) => t.name);
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");

  const filtered = allTags.filter((t) =>
    t.toLowerCase().includes(search.toLowerCase()),
  );
  const hasActive = filterTags.length > 0;

  return (
    <Dropdown
      mode="content"
      portal
      open={open}
      onOpenChange={(v) => {
        setOpen(v);
        if (v) setSearch("");
      }}
    >
      <Dropdown.Trigger>
        <button
          type="button"
          className={cn(
            "flex items-center gap-1.5 h-8 px-3 text-xs font-medium rounded-md border transition-all duration-150",
            hasActive
              ? "bg-primary/15 text-primary border-primary/25"
              : "bg-card text-text-muted border-border hover:text-text-secondary hover:border-border",
          )}
        >
          <TagIcon className="w-3 h-3" strokeWidth={2} />
          Tags
          {hasActive && (
            <span className="w-4 h-4 rounded-full bg-primary text-white text-3xs font-bold flex items-center justify-center leading-none">
              {filterTags.length}
            </span>
          )}
          <ChevronDownIcon
            className={cn("w-3 h-3 transition-transform", open && "rotate-180")}
            strokeWidth={2}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel aria-label="Filter by tags" className="w-[240px]">
        <div className="p-2 border-b border-border">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search tags..."
            className="w-full px-2.5 py-1.5 bg-card border border-border rounded-lg text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
          />
        </div>

        <div className="max-h-[200px] overflow-y-auto p-1">
          {filtered.length === 0 ? (
            <p className="px-2.5 py-3 text-2xs text-text-muted text-center">
              No tags found
            </p>
          ) : (
            filtered.map((tag) => {
              const active = filterTags.includes(tag);
              return (
                <button
                  type="button"
                  key={tag}
                  onClick={() => {
                    if (active) onRemove(tag);
                    else onAdd(tag);
                  }}
                  className="w-full flex items-center gap-2 px-2.5 py-1.5 text-xs rounded-md hover:bg-hover-medium transition-colors"
                >
                  <span
                    className={cn(
                      "w-3.5 h-3.5 rounded border flex items-center justify-center shrink-0 transition-all",
                      active
                        ? "bg-primary border-primary"
                        : "border-text-muted/30",
                    )}
                  >
                    {active && (
                      <CheckIcon
                        className="w-2.5 h-2.5 text-white"
                        strokeWidth={3}
                      />
                    )}
                  </span>
                  <span
                    className={cn("truncate", active ? "text-primary font-medium" : "text-text-secondary")}
                  >
                    {tag}
                  </span>
                </button>
              );
            })
          )}
        </div>

        <div className="p-2 border-t border-border space-y-1">
          {hasActive && (
            <button
              type="button"
              onClick={() => {
                onClearAll();
                setOpen(false);
              }}
              className="w-full text-left px-2.5 py-1.5 text-2xs text-text-muted hover:text-text-primary hover:bg-hover-medium rounded-md transition-colors"
            >
              Clear all
            </button>
          )}
          {onManageTags && (
            <button
              type="button"
              onClick={() => {
                setOpen(false);
                onManageTags();
              }}
              className="w-full text-left px-2.5 py-1.5 text-2xs text-primary hover:text-primary/80 hover:bg-hover-medium rounded-md transition-colors font-medium flex items-center gap-1"
            >
              <Cog6ToothIcon className="w-3 h-3" strokeWidth={2} />
              Manage tags
            </button>
          )}
        </div>
      </Dropdown.Panel>
    </Dropdown>
  );
}

export default TagFilterDropdown;
