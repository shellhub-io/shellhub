import { useMutation } from "@tanstack/react-query";
import {
  confirmSshApprovalMutation,
  rejectSshApprovalMutation,
  createSshIdentityMutation,
  renameSshIdentityMutation,
  deleteSshIdentityMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useConfirmSSHApproval() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...confirmSshApprovalMutation(),
    onSuccess: invalidate,
  });
}

export function useRejectSSHApproval() {
  return useMutation({
    ...rejectSshApprovalMutation(),
  });
}

export function useCreateSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...createSshIdentityMutation(),
    onSuccess: invalidate,
  });
}

export function useRenameSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...renameSshIdentityMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteSSHIdentity() {
  const invalidate = useInvalidateByIds("listSshIdentities");
  return useMutation({
    ...deleteSshIdentityMutation(),
    onSuccess: invalidate,
  });
}
