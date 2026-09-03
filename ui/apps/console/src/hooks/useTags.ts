import { useGetTags } from "@/client/api";

/** All tag names in the namespace, with loading state for callers that need it. */
export function useTagNames() {
  const { data = [], isLoading } = useGetTags({ page: 1, per_page: 100 });
  return { names: data.map((t) => t.name), isLoading };
}
