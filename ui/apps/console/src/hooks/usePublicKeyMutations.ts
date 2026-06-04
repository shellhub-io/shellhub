import { useMutation } from "@tanstack/react-query";
import {
  createPublicKeyMutation,
  updatePublicKeyMutation,
  deletePublicKeyMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreatePublicKey() {
  const invalidate = useInvalidateByIds("getPublicKeys");
  return useMutation({
    ...createPublicKeyMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdatePublicKey() {
  const invalidate = useInvalidateByIds("getPublicKeys");
  return useMutation({
    ...updatePublicKeyMutation(),
    onSuccess: invalidate,
  });
}

export function useDeletePublicKey() {
  const invalidate = useInvalidateByIds("getPublicKeys");
  return useMutation({
    ...deletePublicKeyMutation(),
    onSuccess: invalidate,
  });
}
