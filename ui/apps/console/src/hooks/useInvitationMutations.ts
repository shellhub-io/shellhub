import { useMutation } from "@tanstack/react-query";
import {
  acceptInviteMutation,
  generateInvitationLinkMutation,
  cancelMembershipInvitationMutation,
} from "@/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Accepts an invitation. The user gains a namespace, so the namespace queries are refreshed
 * along with the invitation list.
 */
export function useAcceptInvite() {
  const invalidate = useInvalidateByIds(
    "getMembershipInvitationList",
    "getNamespace",
    "getNamespaces",
  );
  return useMutation({
    ...acceptInviteMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Mints an invitation link for a namespace. Refreshes the member list too, since a pending
 * invitation is shown there alongside the members.
 */
export function useGenerateInvitationLink() {
  const invalidate = useInvalidateByIds(
    "getNamespaceMembershipInvitationList",
    "listNamespaceMembers",
    "getNamespace",
    "getNamespaces",
  );
  return useMutation({
    ...generateInvitationLinkMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Cancels a pending invitation, which makes its link stop working.
 */
export function useCancelMembershipInvitation() {
  const invalidate = useInvalidateByIds("getNamespaceMembershipInvitationList");
  return useMutation({
    ...cancelMembershipInvitationMutation(),
    onSuccess: invalidate,
  });
}
