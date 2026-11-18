import { useAuthStore } from "../stores/authStore";
import { hasPermission, type Permission } from "../utils/permission";

/**
 * Whether the current role holds every permission given, always true when given none.
 * Gates what the UI offers; the server checks again, so this hides a control rather than
 * enforcing anything.
 */
export function useHasPermission(
  required?: Permission | readonly Permission[],
): boolean {
  const role = useAuthStore((s) => s.role);
  return !required || hasPermission(role, required);
}
