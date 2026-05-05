import {
  forwardRef,
  KeyboardEvent,
  startTransition,
  useId,
  useRef,
  useState,
} from "react";
import { useNavigate } from "react-router-dom";
import {
  ExclamationCircleIcon,
  MagnifyingGlassIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "../common/BaseDialog";
import DataTable, { type Column } from "../common/DataTable";
import DistroIcon from "../common/DistroIcon";
import OnlineDot from "../common/OnlineDot";
import LastSeenCell from "../common/LastSeenCell";
import { useDevices, type NormalizedDevice } from "@/hooks/useDevices";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import {
  useChoiceDevices,
  useSuggestedDevices,
} from "@/hooks/useDeviceChooser";
import { isSdkError } from "@/api/errors";
import { FREE_TIER_DEVICE_LIMIT } from "./DeviceChooserTrigger";

const PER_PAGE = 5;
const SEARCH_DEBOUNCE_MS = 300;

type TabId = "suggested" | "all";

interface DeviceChooserDialogProps {
  open: boolean;
  onClose: () => void;
}

function OsCell({ info }: { info?: NormalizedDevice["info"] }) {
  return (
    <div className="flex items-center gap-2 min-w-0">
      <DistroIcon id={info?.id ?? ""} className="text-base shrink-0" />
      <span className="text-xs text-text-secondary truncate">
        {info?.pretty_name ?? "Unknown"}
      </span>
    </div>
  );
}

export default function DeviceChooserDialog({
  open,
  onClose,
}: DeviceChooserDialogProps) {
  const navigate = useNavigate();
  const titleId = useId();
  const descriptionId = useId();
  const suggestedTabId = useId();
  const allTabId = useId();
  const suggestedPanelId = useId();
  const allPanelId = useId();

  const choice = useChoiceDevices();
  const {
    devices: suggested,
    isLoading: suggestedLoading,
    error: suggestedError,
  } = useSuggestedDevices(open);

  const suggestedEmpty =
    !suggestedLoading && !suggestedError && suggested.length === 0;

  const [userTab, setUserTab] = useState<TabId | null>(null);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<NormalizedDevice[]>([]);
  const [error, setError] = useState<string | null>(null);
  const inFlightRef = useRef(false);

  // Force the All tab whenever Suggested is empty so a refetch that returns
  // [] doesn't strand the user on a tab that would submit zero choices.
  const tab: TabId = suggestedEmpty ? "all" : (userTab ?? "suggested");

  const debouncedSearch = useDebouncedValue(search, SEARCH_DEBOUNCE_MS);

  const {
    devices: allDevices,
    totalCount,
    isLoading: allLoading,
  } = useDevices({
    page,
    perPage: PER_PAGE,
    status: "accepted",
    search: debouncedSearch,
    enabled: tab === "all",
  });

  const totalPages = Math.max(1, Math.ceil(totalCount / PER_PAGE));

  const toggleSelected = (device: NormalizedDevice) => {
    if (userTab === null) setUserTab("all");
    setSelected((prev) => {
      const exists = prev.some((d) => d.uid === device.uid);
      if (exists) return prev.filter((d) => d.uid !== device.uid);
      if (prev.length >= FREE_TIER_DEVICE_LIMIT) return prev;
      return [...prev, device];
    });
  };

  const handleSearchChange = (value: string) => {
    setSearch(value);
    setPage(1);
    setError(null);
  };

  const handleTabChange = (next: TabId) => {
    if (next === "suggested" && suggestedEmpty) return;
    setError(null);
    if (next === "suggested") setSelected([]);
    startTransition(() => setUserTab(next));
  };

  const acceptDisabled =
    choice.isPending ||
    (tab === "suggested" && suggested.length === 0) ||
    (tab === "all" && selected.length === 0);

  const accept = async () => {
    if (inFlightRef.current) return;
    inFlightRef.current = true;
    setError(null);
    try {
      const choices = (tab === "suggested" ? suggested : selected).map(
        (d) => d.uid,
      );
      if (choices.length === 0) return;
      await choice.mutateAsync({ body: { choices } });
      onClose();
    } catch (err) {
      const status = isSdkError(err) ? err.status : undefined;
      setError(
        status === 403
          ? "You don't have permission to choose devices for this namespace."
          : "We couldn't save your selection. Please try again in a few moments.",
      );
    } finally {
      inFlightRef.current = false;
    }
  };

  const goSubscribe = () => {
    onClose();
    void navigate("/settings#billing");
  };

  // Block ESC/backdrop while the mutation is in flight so the user doesn't
  // dismiss the dialog mid-request.
  const canClose = () => !choice.isPending;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      canClose={canClose}
      size="xl"
      aria-labelledby={titleId}
      aria-describedby={descriptionId}
      className="sm:max-h-[85vh]"
    >
      <header className="px-6 pt-6 pb-2 shrink-0">
        <h2 id={titleId} className="text-base font-semibold text-text-primary">
          Update account or select three devices
        </h2>
        <p
          id={descriptionId}
          className="text-xs text-text-muted mt-1.5 leading-relaxed"
        >
          Your namespace has more than three accepted devices and no active
          subscription. Subscribe to ShellHub Cloud to keep them all, or pick
          three devices to remain accepted — the rest will be moved to pending.
        </p>
      </header>

      <TabBar
        tab={tab}
        onChange={handleTabChange}
        suggestedDisabled={suggestedEmpty}
        suggestedTabId={suggestedTabId}
        allTabId={allTabId}
        suggestedPanelId={suggestedPanelId}
        allPanelId={allPanelId}
      />

      <main className="flex-auto overflow-y-auto px-6 min-h-0">
        {tab === "suggested" && (
          <section
            id={suggestedPanelId}
            role="tabpanel"
            aria-labelledby={suggestedTabId}
            className="py-4 space-y-3"
          >
            {suggestedError ? (
              <div
                role="alert"
                className="flex items-start gap-2 rounded-lg bg-accent-red/[0.06] border border-accent-red/20 px-3 py-2.5 text-xs text-accent-red"
              >
                <ExclamationCircleIcon
                  className="w-4 h-4 shrink-0 mt-px"
                  strokeWidth={2}
                />
                <span>
                  We couldn't load the suggested devices. Switch to All or try
                  again.
                </span>
              </div>
            ) : null}
            <SuggestedTab devices={suggested} isLoading={suggestedLoading} />
          </section>
        )}

        {tab === "all" && (
          <section
            id={allPanelId}
            role="tabpanel"
            aria-labelledby={allTabId}
            className="py-4 space-y-4"
          >
            <SearchInput value={search} onChange={handleSearchChange} />
            <AllTab
              devices={allDevices}
              isLoading={allLoading}
              selected={selected}
              onToggle={toggleSelected}
              page={page}
              totalPages={totalPages}
              totalCount={totalCount}
              onPageChange={setPage}
            />
            <SelectedChips
              selected={selected}
              onRemove={(d) => toggleSelected(d)}
            />
            <SelectionStatus
              count={selected.length}
              max={FREE_TIER_DEVICE_LIMIT}
            />
          </section>
        )}
      </main>

      {error && (
        <div
          role="alert"
          className="mx-6 mt-3 flex items-start gap-2 rounded-lg bg-accent-red/[0.06] border border-accent-red/20 px-3 py-2.5 text-xs text-accent-red"
        >
          <ExclamationCircleIcon
            className="w-4 h-4 shrink-0 mt-px"
            strokeWidth={2}
          />
          <span>{error}</span>
        </div>
      )}

      <footer className="px-6 py-4 mt-4 border-t border-border shrink-0 flex items-center justify-end gap-2">
        <button
          type="button"
          onClick={onClose}
          disabled={choice.isPending}
          className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="button"
          onClick={goSubscribe}
          disabled={choice.isPending}
          className="px-4 py-2.5 rounded-lg text-sm font-semibold transition-all bg-card hover:bg-hover-medium border border-border hover:border-border-light text-text-primary disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Subscribe
        </button>
        <button
          type="button"
          onClick={() => void accept()}
          disabled={acceptDisabled}
          aria-disabled={acceptDisabled}
          className="inline-flex items-center gap-1.5 px-5 py-2.5 rounded-lg text-sm font-semibold transition-all bg-primary text-white hover:bg-primary-600 active:scale-[0.98] disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-primary disabled:active:scale-100"
        >
          {choice.isPending && (
            <span
              aria-hidden="true"
              className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"
            />
          )}
          {choice.isPending ? "Saving…" : "Accept"}
        </button>
      </footer>
    </BaseDialog>
  );
}

interface TabBarProps {
  tab: TabId;
  onChange: (next: TabId) => void;
  suggestedDisabled: boolean;
  suggestedTabId: string;
  allTabId: string;
  suggestedPanelId: string;
  allPanelId: string;
}

function TabBar({
  tab,
  onChange,
  suggestedDisabled,
  suggestedTabId,
  allTabId,
  suggestedPanelId,
  allPanelId,
}: TabBarProps) {
  const suggestedRef = useRef<HTMLButtonElement>(null);
  const allRef = useRef<HTMLButtonElement>(null);

  const focusTab = (id: TabId) => {
    if (id === "suggested") suggestedRef.current?.focus();
    if (id === "all") allRef.current?.focus();
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    const order: TabId[] = ["suggested", "all"];
    const enabled = order.filter(
      (id) => id !== "suggested" || !suggestedDisabled,
    );
    const idx = enabled.indexOf(tab);
    let next: TabId | null = null;
    if (event.key === "ArrowRight") next = enabled[(idx + 1) % enabled.length];
    else if (event.key === "ArrowLeft")
      next = enabled[(idx - 1 + enabled.length) % enabled.length];
    else if (event.key === "Home") next = enabled[0];
    else if (event.key === "End") next = enabled[enabled.length - 1];
    if (!next) return;
    event.preventDefault();
    onChange(next);
    focusTab(next);
  };

  return (
    <div
      role="tablist"
      aria-label="Device chooser tabs"
      className="px-6 mt-4 border-b border-border flex items-center gap-1"
    >
      <TabButton
        ref={suggestedRef}
        id={suggestedTabId}
        controls={suggestedPanelId}
        selected={tab === "suggested"}
        disabled={suggestedDisabled}
        onSelect={() => onChange("suggested")}
        onKeyDown={handleKeyDown}
      >
        Suggested
      </TabButton>
      <TabButton
        ref={allRef}
        id={allTabId}
        controls={allPanelId}
        selected={tab === "all"}
        disabled={false}
        onSelect={() => onChange("all")}
        onKeyDown={handleKeyDown}
      >
        All
      </TabButton>
    </div>
  );
}

interface TabButtonProps {
  id: string;
  controls: string;
  selected: boolean;
  disabled: boolean;
  onSelect: () => void;
  onKeyDown: (e: KeyboardEvent<HTMLButtonElement>) => void;
  children: React.ReactNode;
}

const TabButton = forwardRef<HTMLButtonElement, TabButtonProps>(
  function TabButton(
    { id, controls, selected, disabled, onSelect, onKeyDown, children },
    ref,
  ) {
    return (
      <button
        ref={ref}
        type="button"
        role="tab"
        id={id}
        aria-controls={controls}
        aria-selected={selected}
        aria-disabled={disabled || undefined}
        tabIndex={selected ? 0 : -1}
        disabled={disabled}
        onClick={onSelect}
        onKeyDown={onKeyDown}
        className={`relative px-4 py-2.5 text-xs font-semibold transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary/50 rounded-t-md
          ${selected ? "text-text-primary" : "text-text-muted hover:text-text-secondary"}
          ${disabled ? "opacity-40 cursor-not-allowed hover:text-text-muted" : ""}`}
      >
        {children}
        {selected && (
          <span
            aria-hidden="true"
            className="absolute left-2 right-2 -bottom-px h-0.5 bg-primary rounded-full"
          />
        )}
      </button>
    );
  },
);

function SearchInput({
  value,
  onChange,
}: {
  value: string;
  onChange: (next: string) => void;
}) {
  return (
    <label className="relative block">
      <span className="sr-only">Search by hostname</span>
      <MagnifyingGlassIcon
        aria-hidden="true"
        className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none"
      />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Search by hostname"
        className="w-full pl-9 pr-3 py-2 bg-surface border border-border rounded-lg text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/30 transition-colors"
      />
    </label>
  );
}

function SuggestedTab({
  devices,
  isLoading,
}: {
  devices: NormalizedDevice[];
  isLoading: boolean;
}) {
  const columns: Column<NormalizedDevice>[] = [
    {
      key: "selected",
      header: "",
      headerClassName: "w-10",
      render: () => (
        <span
          aria-hidden="true"
          className="inline-flex items-center justify-center w-4 h-4 rounded border border-primary bg-primary text-white"
        >
          <svg
            viewBox="0 0 16 16"
            className="w-3 h-3"
            fill="none"
            stroke="currentColor"
            strokeWidth={3}
          >
            <path
              d="M3 8l3.5 3.5L13 5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </span>
      ),
    },
    {
      key: "online",
      header: "",
      headerClassName: "w-8",
      render: (d) => <OnlineDot online={d.online} />,
    },
    {
      key: "name",
      header: "Hostname",
      render: (d) => (
        <span className="text-sm font-medium text-text-primary">{d.name}</span>
      ),
    },
    {
      key: "os",
      header: "Operating System",
      render: (d) => <OsCell info={d.info} />,
    },
    {
      key: "last_seen",
      header: "Last Seen",
      render: (d) => <LastSeenCell value={d.last_seen} />,
    },
  ];

  return (
    <DataTable<NormalizedDevice>
      columns={columns}
      data={devices}
      rowKey={(d, i) => d.uid ?? `suggested-${i}`}
      isLoading={isLoading}
      loadingMessage="Loading suggested devices…"
      emptyMessage="No suggested devices to show."
      label="Suggested devices"
    />
  );
}

interface AllTabProps {
  devices: NormalizedDevice[];
  isLoading: boolean;
  selected: NormalizedDevice[];
  onToggle: (d: NormalizedDevice) => void;
  page: number;
  totalPages: number;
  totalCount: number;
  onPageChange: (page: number) => void;
}

function AllTab({
  devices,
  isLoading,
  selected,
  onToggle,
  page,
  totalPages,
  totalCount,
  onPageChange,
}: AllTabProps) {
  const isSelected = (d: NormalizedDevice) =>
    selected.some((s) => s.uid === d.uid);
  const atLimit = selected.length >= FREE_TIER_DEVICE_LIMIT;

  const columns: Column<NormalizedDevice>[] = [
    {
      key: "selected",
      header: "",
      headerClassName: "w-10",
      render: (d) => {
        const checked = isSelected(d);
        const disabled = !checked && atLimit;
        return (
          <input
            type="checkbox"
            checked={checked}
            disabled={disabled}
            onChange={() => onToggle(d)}
            aria-label={`Select ${d.name ?? "device"}`}
            className="w-4 h-4 rounded border border-border bg-surface accent-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary/50 disabled:opacity-40 disabled:cursor-not-allowed"
          />
        );
      },
    },
    {
      key: "online",
      header: "",
      headerClassName: "w-8",
      render: (d) => <OnlineDot online={d.online} />,
    },
    {
      key: "name",
      header: "Hostname",
      render: (d) => (
        <span className="text-sm font-medium text-text-primary">{d.name}</span>
      ),
    },
    {
      key: "os",
      header: "Operating System",
      render: (d) => <OsCell info={d.info} />,
    },
    {
      key: "last_seen",
      header: "Last Seen",
      render: (d) => <LastSeenCell value={d.last_seen} />,
    },
  ];

  return (
    <DataTable<NormalizedDevice>
      columns={columns}
      data={devices}
      rowKey={(d, i) => d.uid ?? `all-${i}`}
      isLoading={isLoading}
      loadingMessage="Loading devices…"
      emptyMessage="No devices match your search."
      label="All accepted devices"
      page={page}
      totalPages={totalPages}
      totalCount={totalCount}
      itemLabel="device"
      onPageChange={onPageChange}
    />
  );
}

function SelectedChips({
  selected,
  onRemove,
}: {
  selected: NormalizedDevice[];
  onRemove: (d: NormalizedDevice) => void;
}) {
  if (selected.length === 0) return null;
  return (
    <ul
      aria-label="Selected devices"
      className="flex flex-wrap gap-1.5 list-none p-0 m-0"
    >
      {selected.map((d) => (
        <li
          key={d.uid}
          className="inline-flex items-center gap-1 bg-card border border-border rounded px-1.5 py-0.5 text-xs text-text-secondary"
        >
          <span className="truncate max-w-[180px]">{d.name}</span>
          <button
            type="button"
            onClick={() => onRemove(d)}
            aria-label={`Remove ${d.name ?? "device"} from selection`}
            className="text-text-muted hover:text-text-primary transition-colors rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary/50"
          >
            <XMarkIcon className="w-3 h-3" strokeWidth={2.5} />
          </button>
        </li>
      ))}
    </ul>
  );
}

function SelectionStatus({ count, max }: { count: number; max: number }) {
  const tone =
    count === 0
      ? "text-accent-yellow"
      : count === max
        ? "text-accent-green"
        : "text-text-muted";
  return (
    <p
      role="status"
      aria-live="polite"
      className={`text-2xs font-mono tabular-nums ${tone}`}
    >
      {count} of {max} selected
    </p>
  );
}
