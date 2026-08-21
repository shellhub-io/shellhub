import type { SdkHttpError } from "@/api/errors";

export type SdkResponse<T = unknown> = {
  data: T;
  error: undefined;
  request: Request;
  response: Response;
};

export function mockSdkResponse<T>(
  data: T,
  headers?: HeadersInit,
): SdkResponse<T> {
  return {
    data,
    error: undefined,
    request: new Request("http://localhost"),
    response: new Response(null, headers ? { headers } : undefined),
  };
}

export function paginatedResponse<T>(data: T[], total = data.length) {
  return mockSdkResponse(data, { "X-Total-Count": String(total) });
}

export function makeSdkError(
  status: number,
  headers?: HeadersInit,
): Error & SdkHttpError {
  return Object.assign(new Error("Request failed"), {
    status,
    headers: new Headers(headers),
  });
}
