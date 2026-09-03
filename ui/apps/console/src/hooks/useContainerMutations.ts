import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import {
  createTag,
  pushTagToContainer,
  pullTagFromContainer,
} from "@/client/api";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 *
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
 *
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
