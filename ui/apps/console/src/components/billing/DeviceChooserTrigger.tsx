import { lazy, Suspense, useState } from "react";
import { features } from "@/edition";
import { useCurrentNamespace, useStats } from "@/hooks/useNamespaceRequests";
import { hasActiveSubscription } from "@/utils/billing";
import { useHasPermission } from "@/hooks/useHasPermission";

const DeviceChooserDialog = lazy(() => import("./DeviceChooserDialog"));

/**
 * How many devices a free namespace may keep. Enforced by the server; this is what decides when
 * the chooser is offered.
 */
export const FREE_TIER_DEVICE_LIMIT = 3;

/**
 * Mounts the DeviceChooserDialog when a Cloud namespace owner has more accepted devices than the
 * free tier keeps and no active subscription. Dismissing it is not remembered, so it reopens on
 * every mount until the owner subscribes or picks the devices to keep.
 */
export default function DeviceChooserTrigger() {
  if (!features().billing) return null;
  return <DeviceChooserTriggerInner />;
}

function DeviceChooserTriggerInner() {
  const canChoose = useHasPermission("BillingChooseDevices");
  const { namespace, isLoading: nsLoading } = useCurrentNamespace();
  const { stats, isLoading: statsLoading } = useStats();
  const [dismissed, setDismissed] = useState(false);

  if (nsLoading || statsLoading || !namespace || !stats) return null;

  const overLimit = (stats?.registered_devices ?? 0) > FREE_TIER_DEVICE_LIMIT;
  const noActiveSubscription = !hasActiveSubscription(namespace.billing);
  const shouldShow = canChoose && noActiveSubscription && overLimit;

  if (!shouldShow || dismissed) return null;

  return (
    <Suspense fallback={null}>
      <DeviceChooserDialog open onClose={() => setDismissed(true)} />
    </Suspense>
  );
}
