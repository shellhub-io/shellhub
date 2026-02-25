import { AxiosError } from "axios";

/**
 * Creates a mock AxiosError with the specified status code.
 * Useful for testing error handling scenarios in stores and components.
 *
 * @param status - The HTTP status code for the error
 * @param data - Optional response data (e.g., validation errors array)
 * @returns A mock AxiosError instance
 *
 * @example
 * ```ts
 * const error = createAxiosError(404);
 * mockedFetch.mockRejectedValue(error);
 * ```
 *
 * @example
 * ```ts
 * const error = createAxiosError(400, ["username", "email"]);
 * mockedSignUp.mockRejectedValue(error);
 * ```
 */
export const createAxiosError = (status: number, data?: unknown): AxiosError =>
  new AxiosError(
    undefined,
    String(status),
    undefined,
    undefined,
    { status, statusText: "", data: data ?? {}, headers: {}, config: {} as never },
  );
