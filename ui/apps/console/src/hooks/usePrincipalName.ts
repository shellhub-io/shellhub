import type { SessionPrincipal } from "@/client";
import { useApiKeys } from "@/hooks/useApiKeys";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";

/**
 * A resolver from a session's principal to the name a person reads: a member's email, or an API
 * key's name. It looks among the namespace's members and its first 100 API keys, and answers
 * undefined for an account it does not find there, so the caller can fall back to the id.
 */
export function usePrincipalName(): (
  principal: SessionPrincipal,
) => string | undefined {
  const tenant = useAuthStore((state) => state.tenant);
  const { namespace } = useNamespace(tenant ?? "");
  const { apiKeys } = useApiKeys({ perPage: 100 });

  return (principal) =>
    principal.kind === "api-key"
      ? apiKeys.find((key) => key.id === principal.id)?.name
      : namespace?.members?.find((member) => member.id === principal.id)?.email;
}
