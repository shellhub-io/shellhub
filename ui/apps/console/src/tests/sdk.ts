import type { SdkHttpError } from "@/api/errors";

export type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

export function mockSdkResponse<T>(
  data: T,
  headers?: HeadersInit,
): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(null, headers ? { headers } : undefined),
  };
}

export function makeSdkError(
  status: number,
  headers?: HeadersInit,
): SdkHttpError {
  return Object.assign(new Error("Request failed"), {
    status,
    headers: new Headers(headers),
  });
}
