import { useMutation } from "@tanstack/react-query";
import {
  confirmSshApprovalMutation,
  rejectSshApprovalMutation,
  createSshIdentityMutation,
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
