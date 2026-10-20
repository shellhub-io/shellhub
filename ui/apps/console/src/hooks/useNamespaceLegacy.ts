import { useAuthStore } from "@/stores/authStore";
import { useNamespace, type Namespace } from "@/hooks/useNamespaceRequests";

/**
 * Whether a namespace is legacy. A namespace with no settings is new.
 */
export function isLegacyNamespace(namespace?: Namespace | null): boolean {
  return namespace?.settings?.legacy === true;
}

/**
 * Whether the active namespace is legacy: `true` or `false` once the namespace
 * has loaded, `undefined` while it loads. A guard acts only on a decided value, so nobody is
 * bounced out of a page while the namespace resolves; a display treats `undefined` as new.
 */
export function useNamespaceLegacy(): boolean | undefined {
  const { tenant } = useAuthStore();
  const { namespace } = useNamespace(tenant ?? "");
  if (!namespace) return undefined;
  return isLegacyNamespace(namespace);
}
