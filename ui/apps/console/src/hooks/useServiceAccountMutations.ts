import { useMutation } from "@tanstack/react-query";
import {
  createServiceAccountMutation,
  deleteServiceAccountMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a service account. It brings an SSH identity with it, so both lists are refreshed.
 */
export function useCreateServiceAccount() {
  const invalidate = useInvalidateByIds(
    "listServiceAccounts",
    "listSshIdentities",
  );
  return useMutation({
    ...createServiceAccountMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a service account and the identity that belongs to it.
 */
export function useDeleteServiceAccount() {
  const invalidate = useInvalidateByIds(
    "listServiceAccounts",
    "listSshIdentities",
  );
  return useMutation({
    ...deleteServiceAccountMutation(),
    onSuccess: invalidate,
  });
}
