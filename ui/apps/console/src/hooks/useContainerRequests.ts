import { useMutation } from "@tanstack/react-query";
import {
  updateContainerMutation,
  pushTagToContainer,
  pullTagFromContainer,
} from "@/client";
import { ensureTag } from "@/hooks/useTagRequests";
import {
  useInvalidateByIds,
  DEVICE_ROW_QUERIES,
} from "@/hooks/useInvalidateQueries";

/**
 * Renames a container and refreshes every query that shows it (DEVICE_ROW_QUERIES).
 */
export function useRenameContainer() {
  const invalidate = useInvalidateByIds(...DEVICE_ROW_QUERIES);
  return useMutation({
    ...updateContainerMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Tags a container, creating the tag first when it does not exist yet.
 */
export function useAddContainerTag() {
  const invalidate = useInvalidateByIds(...DEVICE_ROW_QUERIES, "getTags");
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      await ensureTag(options.path.name);
      return pushTagToContainer({
        path: { uid: options.path.uid, name: options.path.name },
        throwOnError: true,
      });
    },
    onSuccess: invalidate,
  });
}

/**
 * Removes a tag from a container. The tag survives on anything else carrying it.
 */
export function useRemoveContainerTag() {
  const invalidate = useInvalidateByIds(...DEVICE_ROW_QUERIES);
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      return pullTagFromContainer({
        path: { uid: options.path.uid, name: options.path.name },
        throwOnError: true,
      });
    },
    onSuccess: invalidate,
  });
}
