import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import {
  deleteContainerMutation,
  updateContainerMutation,
  updateContainerStatusMutation,
  createTag,
  pushTagToContainer,
  pullTagFromContainer,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Accepts or rejects a pending container, refreshing the list and the container itself.
 */
export function useUpdateContainerStatus() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
  return useMutation({
    ...updateContainerStatusMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Removes a container from the namespace.
 */
export function useRemoveContainer() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
  return useMutation({
    ...deleteContainerMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Renames a container.
 */
export function useRenameContainer() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
  return useMutation({
    ...updateContainerMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Tags a container. The tag list is refreshed too, because a tag may not have existed before.
 */
export function useAddContainerTag() {
  const invalidate = useInvalidateByIds(
    "getContainers",
    "getContainer",
    "getTags",
  );
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      try {
        await createTag({
          body: { name: options.path.name },
          throwOnError: true,
        });
      } catch (e) {
        if (!isSdkError(e) || e.status !== 409) throw e;
      }
      return pushTagToContainer({
        path: { uid: options.path.uid, name: options.path.name },
        throwOnError: true,
      });
    },
    onSuccess: invalidate,
  });
}

/**
 * Removes a tag from a container. The tag itself survives on anything else carrying it, so the
 * tag list is not refreshed.
 */
export function useRemoveContainerTag() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
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
