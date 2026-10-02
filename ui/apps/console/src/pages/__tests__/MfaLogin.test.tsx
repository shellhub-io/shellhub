import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  act,
  render,
  screen,
  fireEvent,
  waitFor,
} from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import {
  PENDING_DEVICE_CODE_KEY,
  hasPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import MfaLogin from "../MfaLogin";

function renderMfaLogin(initialEntry = "/login-mfa") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/login-mfa" element={<MfaLogin />} />
        <Route path="/login" element={<div>Login Page</div>} />
        <Route path="/dashboard" element={<div>Dashboard</div>} />
        <Route path="/accept-device" element={<div>Accept Device</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

function fillCode(digits = 6, digitValue?: string) {
  const inputs = screen.getAllByRole("textbox");
  inputs.slice(0, digits).forEach((input, i) => {
    fireEvent.change(input, {
      target: { value: digitValue ?? String(i + 1) },
    });
  });
}

function submitCode() {
  fireEvent.click(screen.getByRole("button", { name: /verify/i }));
}

function respondToCode(status: number, headers: Record<string, string> = {}) {
  server.use(
    http.post("*/api/user/mfa/auth", () =>
      HttpResponse.json({}, { status, headers }),
    ),
  );
}

async function passSeconds(seconds: number) {
  for (let i = 0; i < seconds; i++) {
    await act(() => vi.advanceTimersByTimeAsync(1000));
  }
}

function lockOutMfa(secondsLeft: number) {
  const epoch = Math.floor(Date.now() / 1000) + secondsLeft;
  respondToCode(429, { "x-account-lockout": String(epoch) });
}

describe("MfaLogin", () => {
  beforeEach(() => {
    localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
    useAuthStore.setState(useAuthStore.getInitialState(), true);
    useAuthStore.setState({ mfaToken: "temp-mfa-token" });
  });

  it("renders MFA login form when mfaToken exists", () => {
    renderMfaLogin();

    expect(screen.getByText("Two-Factor Authentication")).toBeInTheDocument();
    expect(screen.getByText(/Verification Code/i)).toBeInTheDocument();
  });

  it("redirects to login when no mfaToken", async () => {
    useAuthStore.setState({ mfaToken: null });
    renderMfaLogin();

    await waitFor(() => {
      expect(screen.getByText("Login Page")).toBeInTheDocument();
    });
  });

  it("submits code and navigates on success", async () => {
    const mockLoginWithMfa = vi.fn().mockResolvedValue(undefined);
    useAuthStore.setState({ loginWithMfa: mockLoginWithMfa });

    renderMfaLogin();
    fillCode();
    submitCode();

    await waitFor(() => {
      expect(mockLoginWithMfa).toHaveBeenCalledWith("123456");
    });
  });

  it("displays error message on invalid code", async () => {
    const mockLoginWithMfa = vi.fn().mockImplementation(async () => {
      useAuthStore.setState({ error: "Invalid verification code" });
      throw new Error("Invalid verification code");
    });
    useAuthStore.setState({
      loginWithMfa: mockLoginWithMfa,
      error: null,
    });

    renderMfaLogin();
    fillCode(6, "9");
    submitCode();

    await waitFor(() => {
      expect(screen.getByText("Invalid verification code")).toBeInTheDocument();
    });
  });

  it("keeps the user on the code step, with the token, after a wrong code", async () => {
    respondToCode(401);

    renderMfaLogin();
    fillCode(6, "9");
    submitCode();

    await screen.findByText("Invalid verification code");
    expect(useAuthStore.getState().mfaToken).toBe("temp-mfa-token");
    expect(screen.queryByText("Login Page")).not.toBeInTheDocument();
  });

  describe("lockout", () => {
    beforeEach(() => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("shows the lockout and its remaining time instead of a wrong-code message", async () => {
      lockOutMfa(30);

      renderMfaLogin();
      fillCode();
      submitCode();

      await screen.findByText(/too many failed attempts/i);
      expect(
        screen.queryByText("Invalid verification code"),
      ).not.toBeInTheDocument();
      await passSeconds(1);
      expect(screen.getByText(/\(2[89] seconds\)/)).toBeInTheDocument();
      expect(useAuthStore.getState().mfaToken).toBe("temp-mfa-token");
    });

    it("tells the user to try again once the lockout ends", async () => {
      lockOutMfa(1);

      renderMfaLogin();
      fillCode();
      submitCode();

      await screen.findByText(/too many failed attempts/i);
      await passSeconds(3);
      expect(
        screen.getByText(/your timeout has finished/i),
      ).toBeInTheDocument();
    });

    it("shows the lockout without a countdown when the deadline header is missing", async () => {
      respondToCode(429);

      renderMfaLogin();
      fillCode();
      submitCode();

      await screen.findByText(/too many failed attempts/i);
      await passSeconds(3);
      expect(
        screen.queryByText(/your timeout has finished/i),
      ).not.toBeInTheDocument();
      expect(screen.queryByText(/seconds|minutes/i)).not.toBeInTheDocument();
    });
  });

  it("has link to recovery page", () => {
    renderMfaLogin();

    expect(
      screen.getByRole("link", { name: /use a recovery code/i }),
    ).toHaveAttribute("href", "/mfa-recover");
  });

  it.each([
    [3, true],
    [6, false],
  ])("a %i-digit code leaves submit disabled: %s", (digits, disabled) => {
    renderMfaLogin();
    fillCode(digits);

    const verify = screen.getByRole("button", { name: /verify/i });
    if (disabled) expect(verify).toBeDisabled();
    else expect(verify).not.toBeDisabled();
  });

  it("shows loading state during submission", () => {
    useAuthStore.setState({ loading: true });
    renderMfaLogin();
    fillCode();

    expect(screen.getByText(/Verifying.../i)).toBeInTheDocument();
  });

  describe("pending device code", () => {
    it("redirects to /accept-device when a pending code exists and no explicit redirect", async () => {
      const mockLoginWithMfa = vi.fn().mockResolvedValue(undefined);
      useAuthStore.setState({ loginWithMfa: mockLoginWithMfa });
      setPendingDeviceCode("WXYZ2K7Q");

      renderMfaLogin();
      fillCode();
      submitCode();

      await waitFor(() => {
        expect(screen.getByText("Accept Device")).toBeInTheDocument();
      });
      expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
    });

    it("prefers an explicit redirect over the pending code", async () => {
      const mockLoginWithMfa = vi.fn().mockResolvedValue(undefined);
      useAuthStore.setState({ loginWithMfa: mockLoginWithMfa });
      setPendingDeviceCode("WXYZ2K7Q");

      renderMfaLogin("/login-mfa?redirect=%2Fdevices");
      fillCode();
      submitCode();

      await waitFor(() => {
        expect(mockLoginWithMfa).toHaveBeenCalledWith("123456");
      });
      expect(hasPendingDeviceCode()).toBe(true);
    });
  });
});
