import { useMutation } from "@tanstack/react-query";
import {
  addNamespaceMemberMutation,
  approveUserMutation,
  removeNamespaceMemberMutation,
  updateNamespaceMemberMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Adds a member to the namespace, refreshing the member list and the namespace itself.
 */
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

/**
 * Changes a member's role. The namespace queries are refreshed because the caller's own
 * permissions may be what changed.
 */
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

/**
 * Removes a member from the namespace.
 */
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

/**
 * Approves an account that a non-superadmin provisioned, clearing awaiting_approval.
 * Instance-admin only — the API gates on it. The account still has to be activated before it
 * can sign in; this only removes the block, which is why the member list is refreshed rather
 * than the account being treated as live.
 */
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
