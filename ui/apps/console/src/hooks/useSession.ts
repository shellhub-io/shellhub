import { useQuery } from "@tanstack/react-query";
import { getSessionOptions } from "../client/@tanstack/react-query.gen";

export function useSession(uid: string) {
  const result = useQuery({
    ...getSessionOptions({ path: { uid } }),
    enabled: !!uid,
  });

  return {
    session: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
