import { useMutation } from "@tanstack/react-query";
import {
  apiKeyCreateMutation,
  apiKeyUpdateMutation,
  apiKeyDeleteMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateApiKey() {
  const invalidate = useInvalidateByIds("apiKeyList");
  return useMutation({
    ...apiKeyCreateMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateApiKey() {
  const invalidate = useInvalidateByIds("apiKeyList");
  return useMutation({
    ...apiKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteApiKey() {
  const invalidate = useInvalidateByIds("apiKeyList");
  return useMutation({
    ...apiKeyDeleteMutation(),
    onSuccess: invalidate,
  });
}
