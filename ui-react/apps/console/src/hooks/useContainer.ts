import { useQuery } from "@tanstack/react-query";
import { getContainerOptions } from "../client/@tanstack/react-query.gen";

export function useContainer(uid: string) {
  const result = useQuery({
    ...getContainerOptions({ path: { uid } }),
    enabled: !!uid,
  });

  return {
    container: (result.data ?? null),
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
