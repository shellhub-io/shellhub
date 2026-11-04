import { useState } from "react";
import { ChevronDownIcon, FunnelIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Dropdown } from "@shellhub/design-system/primitives";
import { TagFilterList } from "@/components/common/TagFilterDropdown";
import { CountBadge, TOOLBAR_CONTROL } from "./toolbar";

/**
 * The Devices page's Filters button: one dropdown holding every filter the list takes, each in its
 * own section, with the count of active filters on the button.
 */
export default function DeviceFilters({
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
  onManageTags: () => void;
}) {
  const [open, setOpen] = useState(false);
  const active = filterTags.length;

  return (
    <Dropdown mode="content" portal open={open} onOpenChange={setOpen}>
      <Dropdown.Trigger>
        <button
          type="button"
          className={cn(
            TOOLBAR_CONTROL,
            active > 0 && "bg-primary/10 text-primary border-primary/25",
          )}
        >
          <FunnelIcon className="w-3.5 h-3.5" strokeWidth={2} />
          Filters {active > 0 && <CountBadge count={active} tone="active" />}
          <ChevronDownIcon
            className={cn("w-3 h-3 transition-transform", open && "rotate-180")}
            strokeWidth={2}
          />
        </button>
      </Dropdown.Trigger>

      <Dropdown.Panel aria-label="Filter devices" className="w-[260px]">
        <p className="px-3 pt-2.5 pb-1 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
          Tags
        </p>
        <TagFilterList
          filterTags={filterTags}
          onAdd={onAdd}
          onRemove={onRemove}
          onClearAll={onClearAll}
          onManageTags={onManageTags}
          close={() => setOpen(false)}
        />
      </Dropdown.Panel>
    </Dropdown>
  );
}
