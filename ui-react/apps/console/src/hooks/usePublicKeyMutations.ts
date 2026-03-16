import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createPublicKeyMutation,
  updatePublicKeyMutation,
  deletePublicKeyMutation,
} from "../client/@tanstack/react-query.gen";

function useInvalidatePublicKeys() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      return (key as { _id: string })._id === "getPublicKeys";
    }
    return false;
  } });
}

export function useCreatePublicKey() {
  const invalidate = useInvalidatePublicKeys();
  return useMutation({
    ...createPublicKeyMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdatePublicKey() {
  const invalidate = useInvalidatePublicKeys();
  return useMutation({
    ...updatePublicKeyMutation(),
    onSuccess: invalidate,
  });
}

export function useDeletePublicKey() {
  const invalidate = useInvalidatePublicKeys();
  return useMutation({
    ...deletePublicKeyMutation(),
    onSuccess: invalidate,
  });
}
