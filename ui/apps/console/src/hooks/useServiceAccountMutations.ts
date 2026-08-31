import { useMutation } from "@tanstack/react-query";
import {
  createServiceAccountMutation,
  deleteServiceAccountMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

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
