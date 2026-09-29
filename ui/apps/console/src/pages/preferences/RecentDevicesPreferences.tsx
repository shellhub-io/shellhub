import { ClockIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import SettingsSection from "@/components/settings/SettingsSection";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useRecentDevicesStore } from "@/stores/recentDevicesStore";
import { formatRelative } from "@/utils/date";

/**
 * The devices this browser connected to lately, per namespace, which the command palette offers
 * first. A namespace the user no longer belongs to is labelled as such rather than by its tenant
 * ID, and nothing is listed until the namespaces are known, so none is mislabelled meanwhile. Forgetting a device, or clearing them all, touches this browser's list only.
 */
export default function RecentDevicesPreferences() {
  const byTenant = useRecentDevicesStore((s) => s.byTenant);
  const forget = useRecentDevicesStore((s) => s.forget);
  const clear = useRecentDevicesStore((s) => s.clear);
  const { namespaces, isLoading } = useNamespaces();
  const groups = Object.entries(byTenant).filter(
    ([, devices]) => devices.length > 0,
  );
  const namespaceName = (tenant: string) =>
    namespaces.find((ns) => ns.tenant_id === tenant)?.name;

  return (
    <SettingsSection
      title="Recent devices"
      description="Devices you connected to from this browser. The command palette lists them first."
      action={
        groups.length > 0 && (
          <Button size="sm" variant="secondary" onClick={clear}>
            Clear history
          </Button>
        )
      }
    >
      {isLoading ? null : groups.length === 0 ? (
        <div className="flex flex-col items-center gap-2 px-5 py-10 rounded-xl border border-dashed border-border text-center">
          <ClockIcon className="w-6 h-6 text-text-muted/60" />
          <p className="text-sm text-text-muted">
            Devices you connect to will appear here.
          </p>
        </div>
      ) : (
        groups.map(([tenant, devices]) => {
          const name = namespaceName(tenant);
          return (
            <div
              key={tenant}
              className="rounded-xl border border-border bg-card overflow-hidden"
            >
              <div className="flex items-center justify-between gap-3 px-5 py-2.5 border-b border-border">
                <span
                  className={
                    name
                      ? "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted"
                      : "text-2xs font-mono uppercase tracking-label text-text-muted/60"
                  }
                >
                  {name ?? "A namespace you left"}
                </span>
                <span className="text-2xs font-mono text-text-muted/60">
                  {devices.length}
                </span>
              </div>
              <ul className="divide-y divide-border">
                {devices.map((device) => (
                  <li
                    key={device.uid}
                    className="flex items-center gap-3 pl-5 pr-3 py-2.5"
                  >
                    <span className="text-sm text-text-primary min-w-0 flex-1 truncate">
                      {device.name}
                    </span>
                    <span className="text-xs text-text-muted shrink-0">
                      {formatRelative(device.connectedAt)}
                    </span>
                    <IconButton
                      aria-label={`Forget ${device.name}`}
                      title="Forget"
                      onClick={() => forget(tenant, device.uid)}
                    >
                      <XMarkIcon className="w-4 h-4" />
                    </IconButton>
                  </li>
                ))}
              </ul>
            </div>
          );
        })
      )}
    </SettingsSection>
  );
}
