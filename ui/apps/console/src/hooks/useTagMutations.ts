import { useMutation } from "@tanstack/react-query";
import {
  createTagMutation,
  deleteTagMutation,
  updateTagMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a tag. The device queries are refreshed too, because a device list may be filtered on
 * tags that now include it.
 */
export function useCreateTag() {
  const invalidate = useInvalidateByIds("getTags", "getDevices", "getDevice");
  return useMutation({
    ...createTagMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Renames a tag. Every device carrying it shows the new name, so the device queries go with it.
 */
export function useUpdateTag() {
  const invalidate = useInvalidateByIds("getTags", "getDevices", "getDevice");
  return useMutation({
    ...updateTagMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a tag, removing it from every device that carried it.
 */
export function useDeleteTag() {
  const invalidate = useInvalidateByIds("getTags", "getDevices", "getDevice");
  return useMutation({
    ...deleteTagMutation(),
    onSuccess: invalidate,
  });
}
