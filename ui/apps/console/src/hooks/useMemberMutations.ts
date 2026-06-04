import { useMutation } from "@tanstack/react-query";
import {
  addNamespaceMemberMutation,
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
