import { useQueryClient } from "@tanstack/react-query";

export function useInvalidateByIds(...ids: string[]) {
  const queryClient = useQueryClient();
  const idSet = new Set(ids);
  return () => queryClient.invalidateQueries({
    predicate: (query) => {
      const key = query.queryKey[0];
      return (
        typeof key === "object"
        && key !== null
        && "_id" in key
        && typeof key._id === "string"
        && idSet.has(key._id)
      );
    },
  });
}
