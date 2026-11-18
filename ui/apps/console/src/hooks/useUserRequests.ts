import { useQuery } from "@tanstack/react-query";
import { getUserInfoOptions } from "@/client";
import { features } from "@/edition";

/**
 * The signed-in account's info, namespace allowance included. Community never requests it, so
 * `data` stays undefined there.
 */
export function useUserInfo() {
  return useQuery({
    ...getUserInfoOptions(),
    enabled: features().paid,
  });
}
