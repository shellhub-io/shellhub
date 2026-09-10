import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import { createTag, pushTagToDevice } from "@/client/api";
import { useInvalidateByIds } from "./useInvalidateQueries";

/** Creates the tag if it doesn't exist (swallows 409), then pushes it to the device. */
export function useAddDeviceTag() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
    "getTags",
  );
  return useMutation({
    mutationFn: async ({ uid, name }: { uid: string; name: string }) => {
      try {
        await createTag({ name });
      } catch (e) {
        if (!isSdkError(e) || e.status !== 409) throw e;
      }
      return pushTagToDevice(uid, name);
    },
    onSuccess: invalidate,
  });
}
