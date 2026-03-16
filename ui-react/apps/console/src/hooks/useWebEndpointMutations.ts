import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createWebEndpointMutation,
  deleteWebEndpointMutation,
} from "../client/@tanstack/react-query.gen";

function useInvalidateWebEndpoints() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      return (key as { _id: string })._id === "listWebEndpoints";
    }
    return false;
  } });
}

export function useCreateWebEndpoint() {
  const invalidate = useInvalidateWebEndpoints();
  return useMutation({
    ...createWebEndpointMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteWebEndpoint() {
  const invalidate = useInvalidateWebEndpoints();
  return useMutation({
    ...deleteWebEndpointMutation(),
    onSuccess: invalidate,
  });
}
