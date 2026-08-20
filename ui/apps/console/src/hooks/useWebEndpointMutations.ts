import { useMutation } from "@tanstack/react-query";
import {
  createWebEndpointMutation,
  deleteWebEndpointMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateWebEndpoint() {
  const invalidate = useInvalidateByIds("listWebEndpoints");
  return useMutation({
    ...createWebEndpointMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteWebEndpoint() {
  const invalidate = useInvalidateByIds("listWebEndpoints");
  return useMutation({
    ...deleteWebEndpointMutation(),
    onSuccess: invalidate,
  });
}
