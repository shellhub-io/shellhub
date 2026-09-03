import { useGetNamespaces, useGetNamespace } from "@/client/api";
import type {
  Namespace as GeneratedNamespace,
  NamespaceMemberRole,
} from "@/client/model";

/**
 * A namespace as the console uses it: the generated model plus the type the cloud API adds and
 * the spec does not describe.
 */
export type Namespace = GeneratedNamespace & { type?: string };

/**
 * A member of a namespace. status tells an accepted member from an invitation not yet taken up;
 * both appear in the same list, and only the first has an id that means anything.
 */
export interface NamespaceMember {
  id: string;
  role: NamespaceMemberRole;
  email: string;
  added_at?: string;
  status?: "accepted" | "pending";
  account_status?: "confirmed" | "not-confirmed";
  awaiting_approval?: boolean;
}

/**
 * Every namespace the user belongs to. Fetched as one page of a hundred — the switcher shows
 * them all, and nobody is in more.
 */
export function useNamespaces() {
  const {
    data: namespaces = [],
    isLoading,
    error,
    refetch,
  } = useGetNamespaces({ page: 1, per_page: 100 });

  return { namespaces, isLoading, error, refetch };
}

/**
 * One namespace by tenant id. Idle until an id is given.
 */
export function useNamespace(tenantId: string) {
  const {
    data: namespace = null,
    isLoading,
    error,
    refetch,
  } = useGetNamespace(tenantId, { query: { enabled: !!tenantId } });

  return { namespace, isLoading, error, refetch };
}
