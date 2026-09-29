import {
  ChevronRightIcon,
  GlobeAltIcon,
  KeyIcon,
} from "@heroicons/react/24/outline";
import SettingsSection from "@/components/settings/SettingsSection";
import { useNamespaces, type Namespace } from "@/hooks/useNamespaces";
import { useWorkspaceTabs } from "@/hooks/useWorkspaceTabs";
import { useBrowserKeyFingerprints } from "@/hooks/useBrowserKey";
import { namespaceTabId } from "@/stores/workspaceTabsStore";
import { browserLabel } from "@/utils/browserKey";

function KeyState({
  ns,
  fingerprint,
}: {
  ns: Namespace;
  fingerprint: string | null | undefined;
}) {
  if ((ns.settings?.ssh_access_mode ?? "identity") !== "identity")
    return (
      <span className="text-xs text-text-muted">Not used in legacy mode</span>
    );
  if (fingerprint)
    return (
      <code className="text-2xs font-mono text-text-secondary truncate">
        {fingerprint}
      </code>
    );
  return (
    <span className="text-xs text-text-muted">
      No key, made on your first connection
    </span>
  );
}

/**
 * The SSH keys this browser signs in with, one per namespace the user belongs to. The private
 * half never leaves the browser; a namespace in identity mode gets one the first time the user
 * connects there, and one in legacy mode never uses it. Picking a namespace opens its tab on the
 * SSH identities page, where the key is enrolled, and a namespace that cannot be entered says why
 * in its row.
 */
export default function BrowserIdentityPreferences() {
  const { namespaces } = useNamespaces();
  const workspace = useWorkspaceTabs();
  const fingerprints = useBrowserKeyFingerprints(
    namespaces.map((ns) => ns.tenant_id),
  );

  return (
    <SettingsSection
      title="Browser identity"
      description="This browser signs you in to devices with an SSH key of its own, one per namespace. The private key stays in this browser."
    >
      <div className="rounded-xl border border-border bg-card overflow-hidden">
        <div className="flex items-center gap-4 px-5 py-4 border-b border-border">
          <span className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
            <GlobeAltIcon className="w-5 h-5" />
          </span>
          <div className="min-w-0">
            <p className="text-sm font-medium text-text-primary">
              {browserLabel()}
            </p>
            <p className="mt-0.5 text-xs text-text-muted">This browser</p>
          </div>
        </div>
        {namespaces.length === 0 ? (
          <p className="px-5 py-4 text-sm text-text-muted">
            Keys appear here once you belong to a namespace.
          </p>
        ) : (
          <ul className="divide-y divide-border">
            {namespaces.map((ns) => {
              const identity =
                (ns.settings?.ssh_access_mode ?? "identity") === "identity";
              const failure = workspace.failures[namespaceTabId(ns.tenant_id)];
              const row = (
                <>
                  <KeyIcon
                    className={
                      fingerprints[ns.tenant_id]
                        ? "w-4 h-4 text-accent-green shrink-0"
                        : "w-4 h-4 text-text-muted/50 shrink-0"
                    }
                  />
                  <span className="text-sm text-text-primary min-w-0 flex-1 truncate">
                    {ns.name}
                  </span>
                  <KeyState ns={ns} fingerprint={fingerprints[ns.tenant_id]} />
                </>
              );
              return (
                <li key={ns.tenant_id}>
                  {identity ? (
                    <button
                      type="button"
                      onClick={() =>
                        void workspace.openNamespace(ns.tenant_id, ns.name, {
                          landOn: "/ssh-identities",
                        })
                      }
                      className="group w-full flex items-center gap-3 px-5 py-3 text-left hover:bg-hover-subtle transition-colors"
                    >
                      {row}
                      <ChevronRightIcon className="w-4 h-4 text-text-muted shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
                    </button>
                  ) : (
                    <div className="flex items-center gap-3 px-5 py-3">
                      {row}
                    </div>
                  )}
                  {failure && (
                    <p role="alert" className="px-5 pb-3 text-xs text-accent-red">
                      Couldn't open {ns.name}: {failure}
                    </p>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </SettingsSection>
  );
}
