import { useMutation } from "@tanstack/react-query";
import {
  createWebEndpointMutation,
  deleteWebEndpointMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a web endpoint, refreshing the list.
 */
export function useCreateWebEndpoint() {
  const invalidate = useInvalidateByIds("listWebEndpoints");
  return useMutation({
    ...createWebEndpointMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a web endpoint, refreshing the list. The address stops resolving at once.
 */
export function useDeleteWebEndpoint() {
  const invalidate = useInvalidateByIds("listWebEndpoints");
  return useMutation({
    ...deleteWebEndpointMutation(),
    onSuccess: invalidate,
  });
}
