import { useMutation } from "@tanstack/react-query";
import {
  addNamespaceMemberMutation,
  createUserActivationTokenMutation,
  removeNamespaceMemberMutation,
  updateNamespaceMemberMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useAddMember() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...addNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateMemberRole() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...updateNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useRemoveMember() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...removeNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

// Mints a one-time activation link token for a provisioned (not-confirmed)
// account. No cache invalidation: it doesn't change the member list.
export function useCreateActivationToken() {
  return useMutation(createUserActivationTokenMutation());
}
