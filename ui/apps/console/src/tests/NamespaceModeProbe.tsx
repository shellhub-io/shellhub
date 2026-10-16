import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";

/**
 * Shows what {@link useNamespaceLegacy} has decided for the active namespace: `mode: pending` while
 * it loads, then `mode: legacy` or `mode: new`. Rendered beside the component under test, so a
 * case can wait for the decided mode before asserting what the component shows for it.
 */
export function NamespaceModeProbe() {
  const legacy = useNamespaceLegacy();
  const mode = legacy === undefined ? "pending" : legacy ? "legacy" : "new";

  return <div>mode: {mode}</div>;
}
