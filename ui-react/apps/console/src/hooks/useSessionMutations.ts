import { useMutation, useQueryClient } from "@tanstack/react-query";
import { clsoeSessionMutation } from "../client/@tanstack/react-query.gen";

function useInvalidateSessions() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      const id = (key as { _id: string })._id;
      return id === "getSessions" || id === "getSession" || id === "getStats";
    }
    return false;
  } });
}

export function useCloseSession() {
  const invalidate = useInvalidateSessions();
  return useMutation({
    ...clsoeSessionMutation(),
    onSuccess: invalidate,
  });
}
