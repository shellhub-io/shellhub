/**
 * Shape attached to errors by the fetch error interceptor in fetchInterceptors.ts.
 * The interceptor monkey-patches `.status` and `.headers` onto the parsed
 * response body before it is thrown by the SDK with `throwOnError: true`.
 *
 * `message` and `fields` come from the body the API sends. `message` is for API clients — the
 * console renders its own copy, keyed by status; see `apiErrorMessage`.
 */
export interface SdkHttpError {
  status: number;
  headers: Headers;
  message?: unknown;
  fields?: unknown;
}

/**
 * Type guard analogous to `axios.isAxiosError()` — narrows an unknown
 * caught value to an object carrying the HTTP status and response headers
 * that the fetch error interceptor attaches.
 */
export function isSdkError(err: unknown): err is SdkHttpError {
  return (
    typeof err === "object"
    && err !== null
    && "status" in err
    && typeof err.status === "number"
  );
}

const GENERIC_MESSAGE = "Something went wrong. Please try again.";

const STATUS_MESSAGES: Record<number, string> = {
  400: "Some values are invalid. Review the form and try again.",
  401: "Your session has expired. Sign in again.",
  402: "This action needs an active subscription.",
  403: "You do not have permission to do this.",
  404: "We could not find what you asked for.",
  409: "That value is already in use.",
  423: "This account is not active yet.",
  429: "Too many attempts. Wait a moment and try again.",
  500: "Something went wrong on our side. Try again.",
  502: "The service is unavailable right now. Try again shortly.",
  503: "The service is unavailable right now. Try again shortly.",
  504: "The service is unavailable right now. Try again shortly.",
};

/**
 * Turns a caught value into text to show the user, keyed by HTTP status.
 *
 * The console keeps its own copy: the server's `message` member is for API clients, and reads as
 * internal phrasing here.
 */
export function apiErrorMessage(err: unknown): string {
  if (!isSdkError(err)) return GENERIC_MESSAGE;

  return STATUS_MESSAGES[err.status] ?? GENERIC_MESSAGE;
}

/**
 * Reads the per-field detail an API error carries, mapping a request field name to the reason it
 * was rejected. Returns an empty map when the error carries none.
 */
export function apiErrorFields(err: unknown): Record<string, string> {
  if (!isSdkError(err)) return {};

  const { fields } = err;
  if (typeof fields !== "object" || fields === null || Array.isArray(fields)) return {};

  return Object.fromEntries(
    Object.entries(fields as Record<string, unknown>)
      .filter((entry): entry is [string, string] => typeof entry[1] === "string"),
  );
}
