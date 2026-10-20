import { useMutation } from "@tanstack/react-query";
import { generateInvitationLinkMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

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
