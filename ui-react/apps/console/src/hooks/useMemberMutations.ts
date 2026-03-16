import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  addNamespaceMemberMutation,
  removeNamespaceMemberMutation,
  updateNamespaceMemberMutation,
} from "../client/@tanstack/react-query.gen";

function useInvalidateNamespaces() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      const id = (key as { _id: string })._id;
      return id === "getNamespaces" || id === "getNamespace";
    }
    return false;
  } });
}

export function useAddMember() {
  const invalidate = useInvalidateNamespaces();
  return useMutation({
    ...addNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateMemberRole() {
  const invalidate = useInvalidateNamespaces();
  return useMutation({
    ...updateNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useRemoveMember() {
  const invalidate = useInvalidateNamespaces();
  return useMutation({
    ...removeNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}
