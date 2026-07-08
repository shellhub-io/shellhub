import { useMutation } from "@tanstack/react-query";
import {
  createSshIdentityMutation,
  renameSshIdentityMutation,
  deleteSshIdentityMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

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
