import { useMutation } from "@tanstack/react-query";
import {
  createTagMutation,
  deleteTagMutation,
  updateTagMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateTag() {
  const invalidate = useInvalidateByIds("getTags", "getDevices", "getDevice");
  return useMutation({
    ...createTagMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateTag() {
  const invalidate = useInvalidateByIds("getTags", "getDevices", "getDevice");
  return useMutation({
    ...updateTagMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteTag() {
  const invalidate = useInvalidateByIds("getTags", "getDevices", "getDevice");
  return useMutation({
    ...deleteTagMutation(),
    onSuccess: invalidate,
  });
}
