import { useMutation } from "@tanstack/react-query";
import {
  createPublicKeyMutation,
  updatePublicKeyMutation,
  deletePublicKeyMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a public key, refreshing the list.
 */
export function useCreatePublicKey() {
  const invalidate = useInvalidateByIds("getPublicKeys");
  return useMutation({
    ...createPublicKeyMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates a public key, refreshing the list.
 */
export function useUpdatePublicKey() {
  const invalidate = useInvalidateByIds("getPublicKeys");
  return useMutation({
    ...updatePublicKeyMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a public key, refreshing the list. Anything authenticating with it stops working.
 */
export function useDeletePublicKey() {
  const invalidate = useInvalidateByIds("getPublicKeys");
  return useMutation({
    ...deletePublicKeyMutation(),
    onSuccess: invalidate,
  });
}
