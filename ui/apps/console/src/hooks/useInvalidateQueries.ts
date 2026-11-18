import { useQueryClient } from "@tanstack/react-query";
import type * as Generated from "@/client";

/**
 * The id of a generated query, the name of the OpenAPI operation it reads. A family named by a
 * string that is not one fails to compile, so a renamed operation cannot silently invalidate
 * nothing.
 */
export type OperationId = {
  [K in keyof typeof Generated]: K extends `${string}InfiniteQueryKey`
    ? never
    : K extends `${infer Id}QueryKey`
      ? Id
      : never;
}[keyof typeof Generated];

/**
 * Every query a device edit moves: the device and container lists and details, and the
 * namespace's device counts by status. A container is a device row on the server, so the two
 * lists move together.
 */
export const DEVICE_ROW_QUERIES = [
  "getDevices",
  "getDevice",
  "getContainers",
  "getContainer",
  "getStatusDevices",
] as const satisfies readonly OperationId[];

/**
 * Every query a device status change moves: DEVICE_ROW_QUERIES, plus the admin instance stats,
 * each provisioning key's pending count and its registrations, which an edit leaves alone.
 */
export const DEVICE_STATUS_QUERIES = [
  ...DEVICE_ROW_QUERIES,
  "getStats",
  "provisioningKeyList",
  "listTenantRegistrations",
] as const satisfies readonly OperationId[];

/** The namespace list and the current namespace, which carry membership, settings and billing. */
export const NAMESPACE_QUERIES = [
  "getNamespaces",
  "getNamespace",
] as const satisfies readonly OperationId[];

/**
 * The operation id a generated query key was built from, or undefined for a key the generated
 * client did not make.
 */
export function queryOperationId(queryKey: readonly unknown[]): string | undefined {
  const head = queryKey[0];
  return typeof head === "object"
    && head !== null
    && "_id" in head
    && typeof head._id === "string"
    ? head._id
    : undefined;
}

/**
 * Builds an invalidator for whole query families, matched on the first element of the key rather
 * than the whole key. A mutation cannot know which page or filter is cached, so it invalidates
 * by operation id and lets React Query refetch whichever are mounted.
 */
export function useInvalidateByIds(...ids: OperationId[]) {
  const queryClient = useQueryClient();
  const idSet = new Set<string>(ids);
  return () => queryClient.invalidateQueries({
    predicate: (query) => idSet.has(queryOperationId(query.queryKey) ?? ""),
  });
}
