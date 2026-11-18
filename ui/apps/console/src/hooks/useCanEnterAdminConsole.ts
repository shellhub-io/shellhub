import { features } from "@/edition";
import { useAuthStore } from "@/stores/authStore";

/**
 * Whether the signed-in user can open the admin console: an instance administrator on an edition
 * with the paid features. Community has no admin API.
 */
export function useCanEnterAdminConsole(): boolean {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  return isAdmin && features().paid;
}
