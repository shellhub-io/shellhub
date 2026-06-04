import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import {
  deleteContainerMutation,
  updateContainerMutation,
  updateContainerStatusMutation,
} from "../client/@tanstack/react-query.gen";
import { createTag, pushTagToContainer, pullTagFromContainer } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useUpdateContainerStatus() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
  return useMutation({
    ...updateContainerStatusMutation(),
    onSuccess: invalidate,
  });
}

export function useRemoveContainer() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
  return useMutation({
    ...deleteContainerMutation(),
    onSuccess: invalidate,
  });
}

export function useRenameContainer() {
  const invalidate = useInvalidateByIds("getContainers", "getContainer");
  return useMutation({
    ...updateContainerMutation(),
    onSuccess: invalidate,
  });
}

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
      // If this fails, the tag exists globally but is not attached to this container.
      return pushTagToContainer({
        path: { uid: options.path.uid, name: options.path.name },
        throwOnError: true,
      });
    },
    onSuccess: invalidate,
  });
}

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
