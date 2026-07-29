import { useMutation } from "@tanstack/react-query";
import {
  createServiceAccountMutation,
  deleteServiceAccountMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateServiceAccount() {
  // Creating a service account also enrolls an SSH identity, so refresh both lists.
  const invalidate = useInvalidateByIds(
    "listServiceAccounts",
    "listSshIdentities",
  );
  return useMutation({
    ...createServiceAccountMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteServiceAccount() {
  // Deleting a service account cascades to its SSH identities, so refresh both lists.
  const invalidate = useInvalidateByIds(
    "listServiceAccounts",
    "listSshIdentities",
  );
  return useMutation({
    ...deleteServiceAccountMutation(),
    onSuccess: invalidate,
  });
}
