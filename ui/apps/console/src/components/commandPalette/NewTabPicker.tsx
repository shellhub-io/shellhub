import { useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import { PlusIcon } from "@heroicons/react/24/outline";
import { Dropdown } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useCommandPaletteStore } from "@/stores/commandPaletteStore";
import { useCreateNamespaceStore } from "@/stores/createNamespaceStore";
import { useAuthStore } from "@/stores/authStore";
import { useNamespaces } from "@/hooks/useNamespaceRequests";
import { useUserInfo } from "@/hooks/useUserRequests";
import { useWorkspaceTabs } from "@/hooks/useWorkspaceTabs";
import { isEnterpriseOrCloud } from "@/env";
import { canCreateNamespace } from "@/utils/namespaceCreation";
import { isAdminPath } from "@/utils/adminRoute";
import {
  LISTBOX_ID,
  buildNewTabItems,
  fuzzyMatch,
  handleListboxKey,
  icons,
  useListboxHighlight,
  optionId,
} from "./items";
import CommandRow from "./CommandRow";

function NewTabList() {
  const show = useCommandPaletteStore((s) => s.show);
  const close = useCommandPaletteStore((s) => s.close);
  const openCreateNamespace = useCreateNamespaceStore((s) => s.openDialog);
  const tenant = useAuthStore((s) => s.tenant);
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const { namespaces } = useNamespaces();
  const { data: userInfo } = useUserInfo();
  const workspace = useWorkspaceTabs();
  const adminContext = isAdminPath(useLocation().pathname);
  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const listRef = useRef<HTMLDivElement>(null);

  const items = buildNewTabItems({
    namespaces,
    currentTenant: tenant,
    openTabTenants: workspace.tabs.flatMap((t) =>
      t.kind === "namespace" ? [t.tenant] : [],
    ),
    showConnect: !adminContext,
    showAdmin: isEnterpriseOrCloud() && isAdmin,
    canCreateNamespace: canCreateNamespace(userInfo),
    openNamespace: (namespaceTenant, name) => {
      close();
      void workspace.openNamespace(namespaceTenant, name);
    },
    connect: () => show("connect"),
    openAdmin: () => {
      close();
      void workspace.openAdmin();
    },
    createNamespace: () => {
      close();
      openCreateNamespace();
    },
  });

  const term = query.trim();
  const filtered = term
    ? items.filter((item) => fuzzyMatch(term, item.label))
    : items;
  const { safeIndex, activeItem } = useListboxHighlight(
    filtered,
    activeIndex,
    listRef,
  );
  const sections = new Map<string, typeof filtered>();
  filtered.forEach((item) =>
    sections.set(item.section, [...(sections.get(item.section) ?? []), item]),
  );
  const groups = [...sections.values()];

  return (
    <>
      <div className="flex items-center gap-2.5 h-10 px-3 border-b border-border">
        <span className="shrink-0 text-text-muted" aria-hidden="true">
          {icons.namespace}
        </span>
        <input
          type="text"
          role="combobox"
          aria-label="Search namespaces"
          aria-expanded
          aria-controls={filtered.length ? LISTBOX_ID : undefined}
          aria-activedescendant={
            activeItem ? optionId(activeItem.id) : undefined
          }
          aria-autocomplete="list"
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setActiveIndex(0);
          }}
          onKeyDown={(e) =>
            handleListboxKey(e, filtered, safeIndex, setActiveIndex)
          }
          placeholder="Open a namespace…"
          className="flex-1 min-w-0 bg-transparent text-sm text-text-primary placeholder:text-text-secondary focus:outline-none"
        />
      </div>
      <div ref={listRef} className="max-h-80 overflow-y-auto py-1">
        {filtered.length === 0 ? (
          <p
            role="status"
            className="px-4 py-6 text-center text-sm text-text-muted"
          >
            Nothing matches
          </p>
        ) : (
          <div id={LISTBOX_ID} role="listbox" aria-label="New tab options">
            {groups.map((group, g) => (
              <div
                key={group[0].section}
                role="group"
                aria-label={group[0].section}
                className={cn(g > 0 && "mt-1 pt-1 border-t border-border")}
              >
                {group.map((item) => {
                  const i = filtered.indexOf(item);
                  return (
                    <CommandRow
                      key={item.id}
                      item={item}
                      isActive={i === safeIndex}
                      shaking={false}
                      onActivate={() => setActiveIndex(i)}
                    />
                  );
                })}
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
}

/**
 * The tab strip's "+": a picker anchored under it for what opens as a tab. Namespaces lead, in
 * buildNewTabItems order, then the actions buildNewTabItems allows: "Connect to a device" outside
 * the admin console, which hands over to the device palette's connect view; "Create namespace"
 * where canCreateNamespace allows it, which is never on community; and the admin console for an
 * instance admin on enterprise or cloud. Escape or a click outside closes it, and every open starts from an empty filter.
 */
export default function NewTabPicker() {
  const view = useCommandPaletteStore((s) => s.view);

  return (
    <Dropdown
      mode="content"
      portal
      placement="bottom-start"
      open={view === "picker"}
      onOpenChange={(next) => {
        const store = useCommandPaletteStore.getState();
        if (next) store.show("picker");
        else if (store.view === "picker") store.close();
      }}
    >
      <Dropdown.Trigger>
        <button
          type="button"
          aria-label="New tab"
          title="New tab"
          className="mb-[5px] ml-1 w-7 h-7 rounded-md flex items-center justify-center text-text-muted hover:text-text-primary hover:bg-hover-subtle shrink-0"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
        </button>
      </Dropdown.Trigger>
      <Dropdown.Panel
        aria-label="New tab"
        className="w-72 bg-card border-border-light overflow-hidden"
      >
        <NewTabList />
      </Dropdown.Panel>
    </Dropdown>
  );
}
