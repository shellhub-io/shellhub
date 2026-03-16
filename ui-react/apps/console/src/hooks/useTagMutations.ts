import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createTagMutation,
  deleteTagMutation,
  updateTagMutation,
} from "../client/@tanstack/react-query.gen";

function useInvalidateTags() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      const id = (key as { _id: string })._id;
      return id === "getTags" || id === "getDevices" || id === "getDevice";
    }
    return false;
  } });
}

export function useCreateTag() {
  const invalidate = useInvalidateTags();
  return useMutation({
    ...createTagMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateTag() {
  const invalidate = useInvalidateTags();
  return useMutation({
    ...updateTagMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteTag() {
  const invalidate = useInvalidateTags();
  return useMutation({
    ...deleteTagMutation(),
    onSuccess: invalidate,
  });
}
