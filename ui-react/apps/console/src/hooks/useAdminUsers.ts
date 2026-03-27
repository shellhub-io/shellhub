import { useQuery } from "@tanstack/react-query";
import {
  getUsers as getUsersSdk,
  type GetUsersData,
  type UserAdminResponse,
} from "../client";
import {
  getUsersQueryKey,
  getUserOptions,
} from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

function buildUsernameFilter(search: string): string {
  const filter = [
    {
      type: "property",
      params: { name: "username", operator: "contains", value: search },
    },
  ];
  return btoa(JSON.stringify(filter));
}

interface UseAdminUsersParams {
  page?: number;
  perPage?: number;
  search?: string;
}

export function useAdminUsers({
  page = 1,
  perPage = 10,
  search = "",
}: UseAdminUsersParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const query: GetUsersData["query"] = { page, per_page: perPage };
  if (search) query.filter = buildUsernameFilter(search);
  const options = { query };

  const result = useQuery<PaginatedResult<UserAdminResponse>>({
    queryKey: getUsersQueryKey(options),
    queryFn: paginatedQueryFn(getUsersSdk, options),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    users: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

export function useAdminUser(id: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getUserOptions({ path: { id } }),
    enabled: isAdmin && !!id,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });
}
