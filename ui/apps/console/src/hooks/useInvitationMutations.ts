import { useMutation } from "@tanstack/react-query";
import {
  acceptInviteMutation,
  generateInvitationLinkMutation,
  cancelMembershipInvitationMutation,
} from "@/client";
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
