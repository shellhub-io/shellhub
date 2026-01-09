import { vi } from "vitest";
import { AxiosError } from "axios";
import * as handleErrorModule from "@/utils/handleError";

/**
 * Creates a mock AxiosError with the specified status code.
 * Useful for testing error handling scenarios in components and stores.
 * Mocks handleError to prevent expected console output during tests.
 *
 * @param status - The HTTP status code for the error
 * @param message - The error message
 * @returns A mock AxiosError instance
 *
 * @example
 * ```ts
 * const error = createAxiosError(404, "Not Found");
 * vi.spyOn(store, "fetchData").mockRejectedValueOnce(error);
 * ```
 */
export const createAxiosError = (status: number, message: string): AxiosError => {
  vi.spyOn(handleErrorModule, "default").mockImplementation(() => {});
  return new AxiosError(
    message,
    String(status),
    undefined,
    undefined,
    { status, statusText: message, data: {}, headers: {}, config: {} as never },
  );
};
