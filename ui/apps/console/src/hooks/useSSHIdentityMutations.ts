import { useMutation } from "@tanstack/react-query";
import {
  confirmSshApprovalMutation,
  rejectSshApprovalMutation,
  createSshIdentityMutation,
  createApiKeySshIdentityMutation,
  renameSshIdentityMutation,
  deleteSshIdentityMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Approves a pending SSH login. The identity list is refreshed, since approving may enrol one.
 */
export function useConfirmSSHApproval() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...confirmSshApprovalMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Rejects a pending SSH login. Nothing is enrolled, so nothing is invalidated.
 */
export function useRejectSSHApproval() {
  return useMutation({
    ...rejectSshApprovalMutation(),
  });
}

/**
 * Enrols an SSH identity, refreshing the list.
 */
export function useCreateSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...createSshIdentityMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Enrols an SSH identity that an API key owns, which is how an automation is given a way to
 * reach a device. Needs the permission to manage identities, not just to add one's own.
 */
export function useCreateApiKeySSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...createApiKeySshIdentityMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Renames an SSH identity. The key is untouched — only its label changes.
 */
export function useRenameSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...renameSshIdentityMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Revokes an SSH identity. Anything signing with it stops being able to connect.
 */
export function useDeleteSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...deleteSshIdentityMutation(),
    onSuccess: invalidate,
  });
}
