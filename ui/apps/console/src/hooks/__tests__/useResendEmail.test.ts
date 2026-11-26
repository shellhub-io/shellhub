// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, afterEach, vi } from "vitest";
import { act, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { renderHookWithClient } from "@/tests/wrapper";
import { useResendEmail, RESEND_COOLDOWN_S } from "../useResendEmail";

afterEach(() => {
  vi.useRealTimers();
});

function serveResend(status: number) {
  const calls: unknown[] = [];
  server.use(
    http.post("*/api/user/resend_email", async ({ request }) => {
      calls.push(await request.json());
      return status === 200
        ? HttpResponse.json({})
        : HttpResponse.json({}, { status });
    }),
  );
  return calls;
}

describe("useResendEmail", () => {
  it("starts the cooldown and keeps the error when a resend the user asked for is throttled", async () => {
    serveResend(429);
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());

    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S);
    expect(result.current.resendSuccess).toBe(false);
    expect(result.current.resendError).toBe(
      "Wait a minute before asking for another email.",
    );
  });

  it("starts the cooldown and stays quiet when an unprompted resend is throttled", async () => {
    serveResend(429);
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend({ unprompted: true }));

    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S);
    expect(result.current.resendError).toBeNull();
  });

  it("asks for the email by the identifier, then reports success and starts the cooldown", async () => {
    const calls = serveResend(200);
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());

    expect(calls).toEqual([{ identifier: "testuser" }]);
    expect(result.current.resendSuccess).toBe(true);
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S);
  });

  it("says the resend failed, without a cooldown, when the server errors", async () => {
    serveResend(500);
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());

    expect(result.current.resendSuccess).toBe(false);
    expect(result.current.resendCooldown).toBe(0);
    expect(result.current.resendError).toBe(
      "Failed to resend email. Please try again.",
    );
  });

  it("clears the last success before a new resend", async () => {
    serveResend(200);
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));
    await act(() => result.current.handleResend());
    expect(result.current.resendSuccess).toBe(true);

    serveResend(500);
    await act(() => result.current.handleResend());

    expect(result.current.resendSuccess).toBe(false);
  });

  it("is busy while the resend is in flight", async () => {
    server.use(
      http.post("*/api/user/resend_email", () => new Promise(() => {})),
    );
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));

    act(() => {
      void result.current.handleResend();
    });

    await waitFor(() => expect(result.current.resendLoading).toBe(true));
  });

  it("decrements cooldown by 1 each second", async () => {
    vi.useFakeTimers();
    serveResend(200);
    const { result } = renderHookWithClient(() => useResendEmail("testuser"));
    await act(() => result.current.handleResend());
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S);

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S - 1);

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S - 2);
  });

  it("sends nothing when the identifier is empty", async () => {
    const calls = serveResend(200);
    const { result } = renderHookWithClient(() => useResendEmail(""));

    await act(() => result.current.handleResend());

    expect(calls).toEqual([]);
  });
});
