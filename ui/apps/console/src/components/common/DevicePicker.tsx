import { useRef, useState } from "react";
import {
  MagnifyingGlassIcon,
  ChevronUpDownIcon,
} from "@heroicons/react/24/outline";
import { useDevices } from "@/hooks/useDevices";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { useClickOutside } from "@/hooks/useClickOutside";
import OnlineDot from "@/components/common/OnlineDot";
import { INPUT } from "@/utils/styles";

interface Props {
  value: string;
  /** Display name of the selected device, shown until a fresh pick is made.
   *  Falls back to a UID prefix when unknown (e.g. editing an existing entry). */
  valueLabel?: string;
  onChange: (uid: string, name: string) => void;
}

/**
 * Searchable device picker. The query is server-side and paginated (top matches
 * only), so it scales to fleets of thousands instead of loading every device.
 */
export default function DevicePicker({ value, valueLabel, onChange }: Props) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const debounced = useDebouncedValue(search.trim(), 300);
  const ref = useRef<HTMLDivElement>(null);
  useClickOutside(ref, () => setOpen(false));

  const { devices, isLoading } = useDevices({
    status: "accepted",
    perPage: 8,
    search: debounced,
    enabled: open,
  });

  const display = value
    ? valueLabel || `${value.slice(0, 12)}…`
    : "Choose a device…";

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={`${INPUT} flex items-center justify-between text-left`}
      >
        <span
          className={value ? "text-text-primary truncate" : "text-text-muted"}
        >
          {display}
        </span>
        <ChevronUpDownIcon className="w-4 h-4 text-text-muted shrink-0" />
      </button>

      {open && (
        <div className="absolute z-20 mt-1 w-full bg-card border border-border rounded-lg shadow-lg overflow-hidden">
          <div className="flex items-center gap-2 px-3 py-2 border-b border-border">
            <MagnifyingGlassIcon className="w-4 h-4 text-text-muted shrink-0" />
            <input
              // eslint-disable-next-line jsx-a11y/no-autofocus -- focus the search box when the picker opens
              autoFocus
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search devices..."
              className="flex-1 bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
            />
          </div>
          <div className="max-h-60 overflow-y-auto">
            {isLoading ? (
              <div className="px-3 py-3 text-xs font-mono text-text-muted">
                Loading...
              </div>
            ) : devices.length === 0 ? (
              <div className="px-3 py-3 text-xs font-mono text-text-muted">
                No devices found
              </div>
            ) : (
              devices.map((d) => (
                <button
                  key={d.uid}
                  type="button"
                  onClick={() => {
                    onChange(d.uid, d.name);
                    setSearch("");
                    setOpen(false);
                  }}
                  className="w-full flex items-center gap-2 px-3 py-2 text-left text-sm hover:bg-hover-subtle transition-colors"
                >
                  <OnlineDot online={d.online} />
                  <span className="flex-1 truncate text-text-primary">
                    {d.name}
                  </span>
                  {d.info?.platform && (
                    <span className="text-2xs text-text-muted shrink-0">
                      {d.info.platform}
                    </span>
                  )}
                </button>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}
