import { useMutation } from "@tanstack/react-query";
import {
  acceptInviteMutation,
  addNamespaceMemberMutation,
  declineInviteMutation,
  generateInvitationLinkMutation,
  cancelMembershipInvitationMutation,
  updateMembershipInvitationMutation,
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

export function useDeclineInvite() {
  const invalidate = useInvalidateByIds("getMembershipInvitationList");
  return useMutation({
    ...declineInviteMutation(),
    onSuccess: invalidate,
  });
}

export function useSendInvitationEmail() {
  const invalidate = useInvalidateByIds("getNamespaceMembershipInvitationList");
  return useMutation({
    ...addNamespaceMemberMutation(),
    onSuccess: invalidate,
  });
}

export function useGenerateInvitationLink() {
  const invalidate = useInvalidateByIds("getNamespaceMembershipInvitationList");
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

export function useUpdateMembershipInvitation() {
  const invalidate = useInvalidateByIds("getNamespaceMembershipInvitationList");
  return useMutation({
    ...updateMembershipInvitationMutation(),
    onSuccess: invalidate,
  });
}
