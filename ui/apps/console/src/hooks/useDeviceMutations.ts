import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import { createTag, pushTagToDevice } from "@/client/api";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 *
 */
export function useAddDeviceTag() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
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
      return pushTagToDevice({ ...options, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}
