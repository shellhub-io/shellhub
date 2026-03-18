import { useAuthStore } from "../stores/authStore";
import { hasPermission, type Action } from "../utils/permission";

export function useHasPermission(action: Action): boolean {
  const role = useAuthStore((s) => s.role);
  return hasPermission(role, action);
}
