import { useState, type KeyboardEvent, type ReactNode } from "react";
import { createPortal } from "react-dom";
import {
  ArrowsPointingInIcon,
  ArrowsPointingOutIcon,
  Cog6ToothIcon,
  CommandLineIcon,
  PlusIcon,
  ShieldCheckIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { IconButton } from "@shellhub/design-system/primitives";
import {
  useTerminalStore,
  type ConnectionStatus,
} from "@/stores/terminalStore";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import { accountDisplayName, useAuthStore } from "@/stores/authStore";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useWorkspaceTabs } from "@/hooks/useWorkspaceTabs";
import { useNamespaces } from "@/hooks/useNamespaces";
import {
  useWorkspaceTabsStore,
  type WorkspaceTab,
} from "@/stores/workspaceTabsStore";
import { getInitials } from "@/utils/string";
import { useTabReorder, type TabReorder } from "./useTabReorder";
import TerminalSettingsDrawer from "../terminal/TerminalSettingsDrawer";

const STATUS_DOT: Record<ConnectionStatus, string> = {
  connected: "bg-accent-green",
  connecting: "bg-accent-yellow animate-pulse",
  disconnected: "bg-accent-red",
};

interface TabProps {
  id: string;
  active: boolean;
  surfaceClassName: string;
  surfaceColors?: { background: string; foreground: string };
  label: string;
  tooltip?: string;
  sublabel?: string;
  icon: ReactNode;
  onSelect: () => void;
  onClose?: () => void;
  reorder: TabReorder;
}

function Tab({
  id,
  active,
  surfaceClassName,
  surfaceColors,
  label,
  tooltip,
  sublabel,
  icon,
  onSelect,
  onClose,
  reorder,
}: TabProps) {
  const onKeyDown = (e: KeyboardEvent) => {
    const arrow =
      e.key === "ArrowLeft" ? -1 : e.key === "ArrowRight" ? 1 : null;
    if (arrow !== null && (e.ctrlKey || e.metaKey) && e.shiftKey) {
      e.preventDefault();
      e.stopPropagation();
      reorder.onMove(arrow);
    } else if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onSelect();
    } else if (e.key === "Delete" && onClose) {
      e.preventDefault();
      onClose();
    }
  };

  return (
    <div
      role="tab"
      data-tab-id={id}
      aria-selected={active}
      tabIndex={active ? 0 : -1}
      title={tooltip ?? label}
      aria-keyshortcuts="Control+Shift+ArrowLeft Control+Shift+ArrowRight"
      onPointerDown={reorder.onPointerDown}
      onPointerMove={reorder.onPointerMove}
      onPointerUp={reorder.onPointerUp}
      onPointerCancel={reorder.onPointerCancel}
      style={{
        ...(active && surfaceColors
          ? {
              backgroundColor: surfaceColors.background,
              color: surfaceColors.foreground,
            }
          : {}),
        ...(reorder.offset !== 0
          ? { transform: `translateX(${reorder.offset}px)` }
          : {}),
      }}
      onMouseDown={(e) => {
        if (e.button === 1 && onClose) {
          e.preventDefault();
          onClose();
        } else if (e.button === 0) {
          onSelect();
        }
      }}
      onKeyDown={onKeyDown}
      className={cn(
        "group flex items-center gap-2 min-w-0 max-w-[220px] px-3 rounded-t-lg border border-b-0 text-[13px] cursor-default select-none transition-colors duration-200",
        active
          ? cn(
              "relative z-raised h-[39px] -mb-px border-border text-text-primary",
              surfaceClassName,
            )
          : "relative h-[38px] border-transparent text-text-secondary hover:text-text-primary",
        active && sublabel && "h-[44px] pt-[5px]",
        reorder.dragging && "z-terminal-bar",
        reorder.settling &&
          !reorder.dragging &&
          "transition-[transform,color,background-color] duration-150",
      )}
    >
      {icon}
      <span className="truncate">{label}</span>
      {active && sublabel && (
        <span className="absolute left-3 top-0.5 max-w-[calc(100%-24px)] truncate text-[10px] leading-none opacity-50">
          {sublabel}
        </span>
      )}
      {onClose && (
        <button
          type="button"
          tabIndex={-1}
          aria-label={`Close ${label}`}
          onMouseDown={(e) => e.stopPropagation()}
          onPointerDown={(e) => e.stopPropagation()}
          onClick={onClose}
          className={cn(
            "-mr-1 w-[18px] h-[18px] rounded flex items-center justify-center text-text-muted hover:text-text-primary hover:bg-hover-medium shrink-0",
            !active && "invisible group-hover:visible",
          )}
        >
          <XMarkIcon className="w-3 h-3" strokeWidth={2.5} />
        </button>
      )}
    </div>
  );
}

function WorkspaceIcon({
  tab,
  active,
  failed,
  accountName,
}: {
  tab: WorkspaceTab;
  active: boolean;
  failed: boolean;
  accountName: string;
}) {
  if (tab.kind === "account") {
    return (
      <span
        className={cn(
          "w-5 h-5 rounded-full border flex items-center justify-center text-[9px] font-bold font-mono shrink-0",
          active
            ? "bg-primary/15 border-primary/20 text-primary"
            : "bg-card border-border text-text-muted",
        )}
      >
        {getInitials(accountName)}
      </span>
    );
  }
  if (tab.kind === "admin") {
    return (
      <ShieldCheckIcon
        className={cn(
          "w-4 h-4 shrink-0",
          active ? "text-primary" : "text-text-muted",
        )}
      />
    );
  }
  return (
    <span
      className={cn(
        "w-5 h-5 rounded border flex items-center justify-center text-[9px] font-bold font-mono shrink-0",
        failed
          ? "bg-accent-red/10 border-accent-red/30 text-accent-red"
          : active
            ? "bg-primary/15 border-primary/20 text-primary"
            : "bg-card border-border text-text-muted",
      )}
    >
      {getInitials(tab.name)}
    </span>
  );
}

function TerminalIcon({ status }: { status: ConnectionStatus }) {
  return (
    <span className="relative flex items-center shrink-0">
      <CommandLineIcon className="w-4 h-4 opacity-70" />
      <span
        className={cn(
          "absolute -right-0.5 -bottom-0.5 w-1.5 h-1.5 rounded-full ring-2 ring-background",
          STATUS_DOT[status],
        )}
      />
    </span>
  );
}

function moveFocus(e: KeyboardEvent<HTMLDivElement>) {
  if (e.key !== "ArrowLeft" && e.key !== "ArrowRight") return;
  if (e.ctrlKey || e.metaKey) return;
  const tabs = Array.from(
    e.currentTarget.querySelectorAll<HTMLElement>('[role="tab"]'),
  );
  const current = tabs.indexOf(document.activeElement as HTMLElement);
  if (current < 0) return;
  e.preventDefault();
  const step = e.key === "ArrowRight" ? 1 : -1;
  tabs[(current + step + tabs.length) % tabs.length].focus();
}

/**
 * The tab strip above the framed content. The open contexts come first, namespaces and the admin
 * console, then every terminal session. The terminal in view carries its namespace atop the tab
 * once more than one namespace is open, and selecting a terminal from another namespace enters
 * that namespace first. Selecting a context minimizes the terminals, the same state navigating
 * away leaves them in. Tabs are reordered by dragging, or with Ctrl+Shift+Left/Right, each within
 * its own group: contexts always come first, terminals after. The strip is the framed panel's
 * peer: its root carries data-first-tab-active, which the panel right after it reads to square its
 * top-left corner, so the two must stay siblings.
 */
export default function TabStrip({
  leading,
  trailing,
}: {
  leading?: ReactNode;
  trailing?: ReactNode;
}) {
  const workspace = useWorkspaceTabs();
  const accountName = useAuthStore(accountDisplayName);
  const { namespaces } = useNamespaces();
  const sessions = useTerminalStore((s) => s.sessions);
  const closeSession = useTerminalStore((s) => s.close);
  const toggleFullscreen = useTerminalStore((s) => s.toggleFullscreen);
  const openPalette = useCommandPaletteStore((s) => s.openPalette);
  const terminalColors = useTerminalThemeStore((s) => s.theme.colors);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const moveContext = useWorkspaceTabsStore((s) => s.move);
  const moveSession = useTerminalStore((s) => s.move);
  const reorder = useTabReorder();
  const contextIds = workspace.tabs.map((t) => t.id);
  const contextOrder = reorder.orderOf(contextIds);
  const sessionIds = sessions.map((s) => s.id);

  const active = sessions.find((s) => s.state !== "minimized");
  const firstTabActive =
    !leading && !active && contextOrder[0] === workspace.activeId;
  const openNamespaceTabs = workspace.tabs.filter(
    (t) => t.kind === "namespace",
  ).length;
  const namespaceName = (tenant?: string) =>
    namespaces.find((ns) => ns.tenant_id === tenant)?.name;

  return (
    // eslint-disable-next-line jsx-a11y/interactive-supports-focus -- focus sits on the tabs; the list only relays the arrow keys between them
    <div
      role="tablist"
      aria-label="Open views"
      data-tauri-drag-region
      onKeyDown={moveFocus}
      onKeyDownCapture={(e) => {
        if (e.key === "Escape") reorder.cancel();
      }}
      data-first-tab-active={firstTabActive}
      className="peer h-12 shrink-0 flex items-end gap-0.5 min-w-0"
    >
      {leading}
      {workspace.tabs.map((tab) => {
        const isActive = !active && tab.id === workspace.activeId;
        const failure = workspace.failures[tab.id];
        return (
          <Tab
            key={tab.id}
            id={tab.id}
            active={isActive}
            surfaceClassName="bg-surface"
            label={tab.name}
            tooltip={
              failure ? `Couldn't open ${tab.name}: ${failure}` : undefined
            }
            icon={
              <WorkspaceIcon
                tab={tab}
                active={isActive}
                failed={!!failure}
                accountName={accountName}
              />
            }
            onSelect={() => void workspace.activate(tab)}
            onClose={
              workspace.tabs.length > 1 ? () => workspace.close(tab) : undefined
            }
            reorder={reorder.tab(contextIds, moveContext, tab.id)}
          />
        );
      })}

      {sessions.length > 0 && workspace.tabs.length > 0 && (
        <span
          aria-hidden="true"
          className="mx-1.5 mb-3 h-4 w-px shrink-0 bg-border"
        />
      )}

      {sessions.map((s) => {
        const owner = namespaceName(s.tenant);
        return (
          <Tab
            key={s.id}
            id={s.id}
            active={s.id === active?.id}
            surfaceClassName=""
            surfaceColors={terminalColors}
            label={s.deviceName}
            tooltip={owner ? `${s.deviceName} · ${owner}` : undefined}
            sublabel={openNamespaceTabs > 1 ? owner : undefined}
            icon={<TerminalIcon status={s.connectionStatus} />}
            onSelect={() => void workspace.showTerminal(s)}
            onClose={() => closeSession(s.id)}
            reorder={reorder.tab(sessionIds, moveSession, s.id)}
          />
        );
      })}

      <button
        type="button"
        aria-label="Open a device, session or namespace"
        title="Open… (⌘K)"
        onClick={openPalette}
        className="mb-[5px] ml-1 w-7 h-7 rounded-md flex items-center justify-center text-text-muted hover:text-text-primary hover:bg-hover-subtle shrink-0"
      >
        <PlusIcon className="w-4 h-4" strokeWidth={2} />
      </button>

      <div className="ml-auto mb-1 flex items-center gap-0.5 shrink-0">
        {active && (
          <>
            <IconButton
              aria-label={
                active.state === "fullscreen" ? "Exit fullscreen" : "Fullscreen"
              }
              title={
                active.state === "fullscreen" ? "Exit fullscreen" : "Fullscreen"
              }
              onClick={() => toggleFullscreen(active.id)}
            >
              {active.state === "fullscreen" ? (
                <ArrowsPointingInIcon className="w-4 h-4" />
              ) : (
                <ArrowsPointingOutIcon className="w-4 h-4" />
              )}
            </IconButton>
            <IconButton
              aria-label="Terminal settings"
              title="Terminal settings"
              onClick={() => setSettingsOpen(true)}
            >
              <Cog6ToothIcon className="w-4 h-4" />
            </IconButton>
            <span className="mx-1.5 h-4 w-px bg-border" aria-hidden="true" />
          </>
        )}
        {trailing}
      </div>

      {createPortal(
        <TerminalSettingsDrawer
          open={settingsOpen}
          onClose={() => setSettingsOpen(false)}
        />,
        document.body,
      )}
    </div>
  );
}
