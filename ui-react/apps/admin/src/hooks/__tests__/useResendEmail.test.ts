// @vitest-environment jsdom
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { useResendEmail, RESEND_COOLDOWN_S } from "../useResendEmail";
import { useSignUpStore } from "../../stores/signUpStore";

const mockResendEmail = vi.fn();

beforeEach(() => {
  useSignUpStore.setState({
    resendEmail: mockResendEmail,
    resendLoading: false,
    resendError: null,
  });
  vi.clearAllMocks();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("useResendEmail", () => {
  it("sets resendSuccess and starts cooldown after successful resend", async () => {
    mockResendEmail.mockResolvedValue(true);
    const { result } = renderHook(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());

    expect(result.current.resendSuccess).toBe(true);
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S);
  });

  it("does not start cooldown or set resendSuccess on failed resend", async () => {
    mockResendEmail.mockResolvedValue(false);
    const { result } = renderHook(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());

    expect(result.current.resendSuccess).toBe(false);
    expect(result.current.resendCooldown).toBe(0);
  });

  it("clears resendSuccess before each new resend call", async () => {
    mockResendEmail.mockResolvedValueOnce(true).mockResolvedValueOnce(false);
    const { result } = renderHook(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());
    expect(result.current.resendSuccess).toBe(true);

    await act(() => result.current.handleResend());
    expect(result.current.resendSuccess).toBe(false);
  });

  it("decrements cooldown by 1 each second", async () => {
    vi.useFakeTimers();
    mockResendEmail.mockResolvedValue(true);
    const { result } = renderHook(() => useResendEmail("testuser"));

    await act(() => result.current.handleResend());
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S);

    await act(() => { vi.advanceTimersByTime(1000); });
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S - 1);

    await act(() => { vi.advanceTimersByTime(1000); });
    expect(result.current.resendCooldown).toBe(RESEND_COOLDOWN_S - 2);
  });

  it("does not call resendEmail when username is empty", async () => {
    const { result } = renderHook(() => useResendEmail(""));

    await act(() => result.current.handleResend());

    expect(mockResendEmail).not.toHaveBeenCalled();
  });

  it("reflects resendLoading from the store", () => {
    useSignUpStore.setState({ resendLoading: true });
    const { result } = renderHook(() => useResendEmail("testuser"));

    expect(result.current.resendLoading).toBe(true);
  });

  it("reflects resendError from the store", () => {
    useSignUpStore.setState({ resendError: "Failed to send" });
    const { result } = renderHook(() => useResendEmail("testuser"));

    expect(result.current.resendError).toBe("Failed to send");
  });
});
