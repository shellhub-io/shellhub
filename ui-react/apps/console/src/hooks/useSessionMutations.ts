import { useMutation } from "@tanstack/react-query";
import { clsoeSessionMutation } from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCloseSession() {
  const invalidate = useInvalidateByIds("getSessions", "getSession", "getStatusDevices");
  return useMutation({
    ...clsoeSessionMutation(),
    onSuccess: invalidate,
  });
}
