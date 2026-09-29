import { useAuthStore } from "../stores/authStore";
import { hasPermission, type Action } from "../utils/permission";

/**
 * Whether the current role may take an action, always true when there is no action to check.
 * Gates what the UI offers; the server checks again, so this hides a control rather than
 * enforcing anything.
 */
export function useHasPermission(action?: Action): boolean {
  const role = useAuthStore((s) => s.role);
  return !action || hasPermission(role, action);
}
