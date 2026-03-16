import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  apiKeyCreateMutation,
  apiKeyUpdateMutation,
  apiKeyDeleteMutation,
} from "../client/@tanstack/react-query.gen";

function useInvalidateApiKeys() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      return (key as { _id: string })._id === "apiKeyList";
    }
    return false;
  } });
}

export function useCreateApiKey() {
  const invalidate = useInvalidateApiKeys();
  return useMutation({
    ...apiKeyCreateMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateApiKey() {
  const invalidate = useInvalidateApiKeys();
  return useMutation({
    ...apiKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteApiKey() {
  const invalidate = useInvalidateApiKeys();
  return useMutation({
    ...apiKeyDeleteMutation(),
    onSuccess: invalidate,
  });
}
