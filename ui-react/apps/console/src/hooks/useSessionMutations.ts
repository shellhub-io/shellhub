import { useMutation } from "@tanstack/react-query";
import { clsoeSessionMutation } from "../client/@tanstack/react-query.gen";
import apiClient from "../api/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCloseSession() {
  const invalidate = useInvalidateByIds("getSessions", "getSession", "getStatusDevices");
  return useMutation({
    ...clsoeSessionMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteSessionRecording() {
  const invalidate = useInvalidateByIds("getSessions", "getSession");
  return useMutation({
    mutationFn: async (uid: string) => {
      await apiClient.delete(`/api/sessions/${uid}/records/0`);
    },
    onSuccess: invalidate,
  });
}
