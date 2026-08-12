import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
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

describe("MfaLogin", () => {
  beforeEach(() => {
    localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
    useAuthStore.setState({
      token: null,
      mfaToken: "temp-mfa-token",
      loading: false,
      error: null,
      loginWithMfa: vi.fn(),
    });
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

  it("has link to recovery page", () => {
    renderMfaLogin();

    const recoveryLink = screen.getByText(/Lost your TOTP password/i);
    expect(recoveryLink).toHaveAttribute("href", "/mfa-recover");
  });

  it("disables submit button when code is incomplete", () => {
    renderMfaLogin();
    fillCode(3);

    expect(screen.getByRole("button", { name: /verify/i })).toBeDisabled();
  });

  it("enables submit button when code is complete", () => {
    renderMfaLogin();
    fillCode();

    expect(screen.getByRole("button", { name: /verify/i })).not.toBeDisabled();
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
