import { describe, it, expect } from "vitest";
import { mockSdkResponse, makeSdkError } from "../sdk";

describe("mockSdkResponse", () => {
  it("wraps data with request and response", () => {
    const res = mockSdkResponse({ id: 1 });
    expect(res.data).toEqual({ id: 1 });
    expect(res.request).toBeInstanceOf(Request);
    expect(res.response).toBeInstanceOf(Response);
  });

  it("forwards headers to the response", () => {
    const res = mockSdkResponse("ok", { "X-Total-Count": "42" });
    expect(res.response.headers.get("X-Total-Count")).toBe("42");
  });
});

describe("makeSdkError", () => {
  it("returns an Error with status and headers", () => {
    const err = makeSdkError(409);
    expect(err).toBeInstanceOf(Error);
    expect(err.status).toBe(409);
    expect(err.headers).toBeInstanceOf(Headers);
  });

  it("forwards headers", () => {
    const err = makeSdkError(429, { "Retry-After": "60" });
    expect(err.headers.get("Retry-After")).toBe("60");
  });
});
