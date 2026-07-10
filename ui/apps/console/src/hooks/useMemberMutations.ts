import { useMutation } from "@tanstack/react-query";
import {
  addNamespaceMemberMutation,
  approveUserMutation,
  removeNamespaceMemberMutation,
  updateNamespaceMemberMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useAddMember() {
  const invalidate = useInvalidateByIds(
    "getNamespaces",
    "getNamespace",
    "listNamespaceMembers",
  );
  return useMutation({
    ...addNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateMemberRole() {
  const invalidate = useInvalidateByIds(
    "getNamespaces",
    "getNamespace",
    "listNamespaceMembers",
  );
  return useMutation({
    ...updateNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useRemoveMember() {
  const invalidate = useInvalidateByIds(
    "getNamespaces",
    "getNamespace",
    "listNamespaceMembers",
  );
  return useMutation({
    ...removeNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

// Approves an account provisioned by a non-superadmin (awaiting_approval).
// Instance-admin only; the API gates on it. Clearing the flag lets the account
// log in once activated, so refresh the member list.
export function useApproveMember() {
  const invalidate = useInvalidateByIds(
    "getNamespaces",
    "getNamespace",
    "listNamespaceMembers",
  );
  return useMutation({
    ...approveUserMutation(),
    onSuccess: invalidate,
  });
}
