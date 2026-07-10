import { useMutation } from "@tanstack/react-query";
import {
  acceptInviteMutation,
  generateInvitationLinkMutation,
  cancelMembershipInvitationMutation,
} from "@/client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

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

export function useGenerateInvitationLink() {
  // Enterprise adds an existing account directly (no invitation), so refresh the members list
  // and namespace too — not just the pending invitations.
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

export function useCancelMembershipInvitation() {
  const invalidate = useInvalidateByIds("getNamespaceMembershipInvitationList");
  return useMutation({
    ...cancelMembershipInvitationMutation(),
    onSuccess: invalidate,
  });
}
