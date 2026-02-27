import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDevicesStore } from "../stores/devicesStore";
import { useTerminalStore } from "../stores/terminalStore";
import { useAuthStore } from "../stores/authStore";
import {
  MagnifyingGlassIcon,
  HomeIcon,
  CubeIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
  KeyIcon,
  Cog6ToothIcon,
  ArrowRightStartOnRectangleIcon,
  UsersIcon,
  CpuChipIcon,
  CommandLineIcon,
  PlusIcon,
  ChevronDoubleRightIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";

/* ─── Types ─── */

type BadgeVariant = "green" | "yellow" | "red" | "muted";

interface CommandItem {
  id: string;
  label: string;
  sublabel?: string;
  section: string;
  icon: JSX.Element;
  badge?: { text: string; variant: BadgeVariant };
  onSelect: () => void;
}

const badgeStyles: Record<BadgeVariant, string> = {
  green: "text-accent-green bg-accent-green/10 border-accent-green/20",
  yellow: "text-accent-yellow bg-accent-yellow/10 border-accent-yellow/20",
  red: "text-accent-red bg-accent-red/10 border-accent-red/20",
  muted: "text-text-muted bg-hover-medium border-border",
};

/* ─── Icons ─── */

const icons = {
  search: <MagnifyingGlassIcon className="w-5 h-5" />,
  dashboard: <HomeIcon className="w-4 h-4" />,
  devices: <CpuChipIcon className="w-4 h-4" />,
  sessions: <CommandLineIcon className="w-4 h-4" />,
  keys: <KeyIcon className="w-4 h-4" />,
  firewall: <ShieldCheckIcon className="w-4 h-4" />,
  settings: <Cog6ToothIcon className="w-4 h-4" />,
  add: <PlusIcon className="w-4 h-4" />,
  terminal: <ChevronDoubleRightIcon className="w-4 h-4" />,
  logout: <ArrowRightStartOnRectangleIcon className="w-4 h-4" />,
  team: <UsersIcon className="w-4 h-4" />,
  containers: <CubeIcon className="w-4 h-4" />,
  web: <GlobeAltIcon className="w-4 h-4" />,
  vault: <LockClosedIcon className="w-4 h-4" />,
};

/* ─── Fuzzy match ─── */

function fuzzyMatch(query: string, text: string): boolean {
  const q = query.toLowerCase();
  const t = text.toLowerCase();
  if (t.includes(q)) return true;
  let qi = 0;
  for (let ti = 0; ti < t.length && qi < q.length; ti++) {
    if (t[ti] === q[qi]) qi++;
  }
  return qi === q.length;
}

/* ─── Component ─── */

export default function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const backdropRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const devices = useDevicesStore((s) => s.devices);
  const terminalSessions = useTerminalStore((s) => s.sessions);
  const restoreTerminal = useTerminalStore((s) => s.restore);
  const logout = useAuthStore((s) => s.logout);

  const close = useCallback(() => {
    setOpen(false);
    setQuery("");
    setActiveIndex(0);
  }, []);

  const go = useCallback(
    (path: string) => {
      close();
      navigate(path);
    },
    [close, navigate],
  );

  /* Keyboard shortcut to open */
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        setOpen((prev) => {
          if (prev) {
            setQuery("");
            setActiveIndex(0);
            return false;
          }
          return true;
        });
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  /* Focus input when opened */
  useEffect(() => {
    if (open) {
      requestAnimationFrame(() => inputRef.current?.focus());
    }
  }, [open]);

  /* Build items */
  const items = useMemo<CommandItem[]>(() => {
    const list: CommandItem[] = [];

    /* Navigation */
    const nav: Array<{ label: string; path: string; icon: JSX.Element }> = [
      { label: "Dashboard", path: "/dashboard", icon: icons.dashboard },
      { label: "Devices", path: "/devices", icon: icons.devices },
      { label: "Containers", path: "/containers", icon: icons.containers },
      { label: "Web Endpoints", path: "/webendpoints", icon: icons.web },
      { label: "Sessions", path: "/sessions", icon: icons.sessions },
      { label: "Public Keys", path: "/sshkeys/public-keys", icon: icons.keys },
      { label: "Secure Vault", path: "/secure-vault", icon: icons.vault },
      {
        label: "Firewall Rules",
        path: "/firewall/rules",
        icon: icons.firewall,
      },
      { label: "Team", path: "/team", icon: icons.team },
      { label: "Settings", path: "/settings", icon: icons.settings },
      { label: "Add Device", path: "/devices/add", icon: icons.add },
    ];

    nav.forEach((n) => {
      list.push({
        id: `nav-${n.path}`,
        label: n.label,
        sublabel: n.path,
        section: "Navigation",
        icon: n.icon,
        onSelect: () => go(n.path),
      });
    });

    /* Devices */
    devices
      .filter((d) => d.status === "accepted")
      .forEach((d) => {
        list.push({
          id: `device-${d.uid}`,
          label: d.name,
          sublabel: d.identity?.mac ?? d.uid.slice(0, 12),
          section: "Devices",
          icon: icons.devices,
          badge: d.online
            ? { text: "Online", variant: "green" as BadgeVariant }
            : { text: "Offline", variant: "muted" as BadgeVariant },
          onSelect: () => go(`/devices/${d.uid}`),
        });
      });

    /* Terminal Sessions */
    terminalSessions.forEach((s) => {
      const statusLabel =
        s.connectionStatus === "connected"
          ? "Connected"
          : s.connectionStatus === "connecting"
            ? "Connecting"
            : "Disconnected";
      const statusVariant: BadgeVariant =
        s.connectionStatus === "connected"
          ? "green"
          : s.connectionStatus === "connecting"
            ? "yellow"
            : "red";
      list.push({
        id: `term-${s.id}`,
        label: `${s.username}@${s.deviceName}`,
        sublabel:
          s.state === "minimized"
            ? "Minimized"
            : s.state === "fullscreen"
              ? "Fullscreen"
              : "Docked",
        section: "Terminal Sessions",
        icon: icons.terminal,
        badge: { text: statusLabel, variant: statusVariant },
        onSelect: () => {
          close();
          restoreTerminal(s.id);
        },
      });
    });

    /* Actions */
    list.push({
      id: "action-logout",
      label: "Logout",
      section: "Actions",
      icon: icons.logout,
      onSelect: () => {
        close();
        logout();
        navigate("/login");
      },
    });

    return list;
  }, [devices, terminalSessions, go, close, restoreTerminal, logout, navigate]);

  /* Filtered */
  const filtered = useMemo(() => {
    if (!query.trim()) return items;
    return items.filter(
      (item) =>
        fuzzyMatch(query, item.label) ||
        (item.sublabel && fuzzyMatch(query, item.sublabel)) ||
        fuzzyMatch(query, item.section),
    );
  }, [items, query]);

  /* Group by section */
  const sections = useMemo(() => {
    const map = new Map<string, CommandItem[]>();
    filtered.forEach((item) => {
      const existing = map.get(item.section);
      if (existing) existing.push(item);
      else map.set(item.section, [item]);
    });
    return map;
  }, [filtered]);

  /* Flat list for keyboard nav */
  const flatList = useMemo(() => {
    const flat: CommandItem[] = [];
    sections.forEach((items) => flat.push(...items));
    return flat;
  }, [sections]);

  /* Clamp active index */
  useEffect(() => {
    setActiveIndex((prev) => Math.min(prev, Math.max(flatList.length - 1, 0)));
  }, [flatList.length]);

  /* Scroll active into view */
  useEffect(() => {
    if (!listRef.current) return;
    const active = listRef.current.querySelector("[data-active=true]");
    active?.scrollIntoView({ block: "nearest" });
  }, [activeIndex]);

  /* Keyboard navigation */
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          setActiveIndex((prev) => (prev + 1) % flatList.length);
          break;
        case "ArrowUp":
          e.preventDefault();
          setActiveIndex(
            (prev) => (prev - 1 + flatList.length) % flatList.length,
          );
          break;
        case "Enter":
          e.preventDefault();
          flatList[activeIndex]?.onSelect();
          break;
        case "Escape":
          e.preventDefault();
          close();
          break;
      }
    },
    [flatList, activeIndex, close],
  );

  if (!open) return null;

  let globalIdx = -1;

  return (
    <div className="fixed inset-0 z-[80] flex items-start justify-center pt-[min(20vh,140px)]">
      {/* Backdrop */}
      <div
        ref={backdropRef}
        className="absolute inset-0 bg-black/60 backdrop-blur-sm animate-fade-in"
        onClick={close}
      />

      {/* Palette */}
      <div
        className="relative w-full max-w-[540px] mx-4 bg-surface border border-border rounded-xl shadow-2xl shadow-black/50 overflow-hidden animate-slide-up"
        onKeyDown={handleKeyDown}
      >
        {/* Search input */}
        <div className="flex items-center gap-3 px-4 border-b border-border">
          <span className="text-text-muted shrink-0">{icons.search}</span>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setActiveIndex(0);
            }}
            placeholder="Search pages, devices, sessions..."
            className="flex-1 h-12 bg-transparent text-sm text-text-primary placeholder:text-text-secondary focus:outline-none"
          />
          <kbd className="shrink-0 px-1.5 py-0.5 text-2xs font-mono font-semibold text-text-muted/50 bg-hover-medium border border-border rounded">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div
          ref={listRef}
          className="max-h-[min(50vh,400px)] overflow-y-auto overscroll-contain"
        >
          {flatList.length === 0 ? (
            <div className="px-4 py-10 text-center">
              <p className="text-sm text-text-muted">
                No results for "{query}"
              </p>
              <p className="text-2xs text-text-muted/50 mt-1">
                Try a different search term
              </p>
            </div>
          ) : (
            Array.from(sections.entries()).map(([section, sectionItems]) => (
              <div key={section}>
                <div className="px-4 pt-3 pb-1.5">
                  <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted/50">
                    {section}
                  </p>
                </div>
                {sectionItems.map((item) => {
                  globalIdx++;
                  const isActive = globalIdx === activeIndex;
                  const idx = globalIdx;
                  return (
                    <button
                      key={item.id}
                      data-active={isActive}
                      onClick={item.onSelect}
                      onMouseEnter={() => setActiveIndex(idx)}
                      className={`w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors duration-75 ${
                        isActive ? "bg-primary/[0.08]" : "hover:bg-hover-subtle"
                      }`}
                    >
                      <span
                        className={`shrink-0 ${isActive ? "text-primary" : "text-text-muted"} transition-colors duration-75`}
                      >
                        {item.icon}
                      </span>
                      <div className="flex-1 min-w-0">
                        <span
                          className={`text-sm ${isActive ? "text-text-primary" : "text-text-secondary"} transition-colors duration-75`}
                        >
                          {item.label}
                        </span>
                        {item.sublabel && (
                          <span className="text-2xs text-text-muted/50 ml-2 font-mono">
                            {item.sublabel}
                          </span>
                        )}
                      </div>
                      {item.badge && (
                        <span
                          className={`shrink-0 text-2xs font-mono font-semibold px-1.5 py-0.5 rounded border ${badgeStyles[item.badge.variant]}`}
                        >
                          {item.badge.text}
                        </span>
                      )}
                      {isActive && (
                        <kbd className="shrink-0 px-1.5 py-0.5 text-2xs font-mono text-text-muted/40 bg-hover-subtle border border-border/50 rounded">
                          ↵
                        </kbd>
                      )}
                    </button>
                  );
                })}
              </div>
            ))
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center gap-4 px-4 py-2.5 border-t border-border bg-card/30">
          <div className="flex items-center gap-1.5">
            <kbd className="px-1 py-0.5 text-2xs font-mono text-text-muted/40 bg-hover-subtle border border-border/50 rounded">
              ↑
            </kbd>
            <kbd className="px-1 py-0.5 text-2xs font-mono text-text-muted/40 bg-hover-subtle border border-border/50 rounded">
              ↓
            </kbd>
            <span className="text-2xs text-text-muted/40">navigate</span>
          </div>
          <div className="flex items-center gap-1.5">
            <kbd className="px-1 py-0.5 text-2xs font-mono text-text-muted/40 bg-hover-subtle border border-border/50 rounded">
              ↵
            </kbd>
            <span className="text-2xs text-text-muted/40">select</span>
          </div>
          <div className="flex items-center gap-1.5 ml-auto">
            <kbd className="px-1.5 py-0.5 text-2xs font-mono text-text-muted/40 bg-hover-subtle border border-border/50 rounded">
              ⌘K
            </kbd>
            <span className="text-2xs text-text-muted/40">toggle</span>
          </div>
        </div>
      </div>
    </div>
  );
}
