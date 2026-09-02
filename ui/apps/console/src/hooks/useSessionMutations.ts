import { useMutation } from "@tanstack/react-query";
import { deleteSessionRecord } from "@/client/api";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 *
 */
export function useDeleteSessionRecording() {
  const invalidate = useInvalidateByIds("getSessions", "getSession");
  return useMutation({
    mutationFn: async (uid: string) => {
      await deleteSessionRecord({ path: { uid, seat: 0 }, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}
