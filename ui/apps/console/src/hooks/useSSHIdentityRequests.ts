import { useMutation, useQuery } from "@tanstack/react-query";
import {
  createSshIdentityMutation,
  deleteSshIdentityMutation,
  listApiKeySshIdentitiesOptions,
  listSshIdentitiesOptions,
  type SshIdentity,
} from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

/**
 * The SSH identities the caller may see: their own, or every member's when they hold the
 * permission to manage them. The server decides which, so there is nothing to ask for here.
 * enabled holds the request back until the caller needs the list.
 */
export function useSSHIdentities({
  enabled = true,
}: { enabled?: boolean } = {}) {
  const result = useQuery({ ...listSshIdentitiesOptions(), enabled });

  const identities: SshIdentity[] = result.data ?? [];

  return {
    identities,
    isLoading: result.isLoading,
    error: result.error,
  };
}

/**
 * The SSH credentials an API key owns. Asked for by key name, which is what its routes are
 * addressed by, and answered only to a caller who may manage identities. Idle until a name is
 * given and enabled is true.
 */
export function useApiKeySshIdentities(name: string, enabled = true) {
  const result = useQuery({
    ...listApiKeySshIdentitiesOptions({ path: { name } }),
    enabled: enabled && !!name,
  });

  const identities: SshIdentity[] = result.data ?? [];

  return {
    identities,
    isLoading: result.isLoading,
    error: result.error,
  };
}

/**
 * Enrols an SSH identity, refreshing the list.
 */
export function useCreateSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...createSshIdentityMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Revokes an SSH identity. Anything signing with it stops being able to connect.
 */
export function useDeleteSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...deleteSshIdentityMutation(),
    onSuccess: invalidate,
  });
}
