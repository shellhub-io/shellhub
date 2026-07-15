import { lazy, Suspense, useState } from "react";
import { isCloud } from "@/env";
import { useStats } from "@/hooks/useStats";
import { useNamespace } from "@/hooks/useNamespaces";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useAuthStore } from "@/stores/authStore";

const DeviceChooserDialog = lazy(() => import("./DeviceChooserDialog"));

export const FREE_TIER_DEVICE_LIMIT = 3;

/**
 * Mounts the DeviceChooserDialog when a Cloud namespace owner has more than
 * three accepted devices and no active subscription. The dialog reopens on
 * every fresh mount: dismissing it doesn't persist anywhere — matches the Vue
 * behavior so the user keeps seeing it until they subscribe or pick three.
 */
export default function DeviceChooserTrigger() {
  if (!isCloud()) return null;
  return <DeviceChooserTriggerInner />;
}

function DeviceChooserTriggerInner() {
  const canChoose = useHasPermission("device:choose");
  const tenantId = useAuthStore((s) => s.tenant);
  const { namespace, isLoading: nsLoading } = useNamespace(tenantId ?? "");
  const { stats, isLoading: statsLoading } = useStats();
  const [dismissed, setDismissed] = useState(false);

  if (!tenantId || nsLoading || statsLoading || !namespace || !stats)
    return null;

  const overLimit = (stats?.registered_devices ?? 0) > FREE_TIER_DEVICE_LIMIT;
  const noActiveSubscription = !namespace.billing?.active;
  const shouldShow = canChoose && noActiveSubscription && overLimit;

  if (!shouldShow || dismissed) return null;

  return (
    <Suspense fallback={null}>
      <DeviceChooserDialog open onClose={() => setDismissed(true)} />
    </Suspense>
  );
}
