import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import MfaLogin from "../MfaLogin";

beforeEach(() => {
  useAuthStore.setState({
    token: null,
    mfaToken: "temp-mfa-token",
    loading: false,
    error: null,
    loginWithMfa: vi.fn(),
  });
});

function renderMfaLogin() {
  return render(
    <MemoryRouter initialEntries={["/mfa-login"]}>
      <Routes>
        <Route path="/mfa-login" element={<MfaLogin />} />
        <Route path="/login" element={<div>Login Page</div>} />
        <Route path="/dashboard" element={<div>Dashboard</div>} />
      </Routes>
    </MemoryRouter>
  );
}

describe("MfaLogin", () => {
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

    const inputs = screen.getAllByRole("textbox");

    // Enter 6-digit code
    inputs.forEach((input, i) => {
      fireEvent.change(input, { target: { value: String(i + 1) } });
    });

    const submitBtn = screen.getByRole("button", { name: /verify/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockLoginWithMfa).toHaveBeenCalledWith("123456");
    });
  });

  it("displays error message on invalid code", async () => {
    const mockLoginWithMfa = vi.fn().mockRejectedValue(new Error("Invalid"));
    useAuthStore.setState({
      loginWithMfa: mockLoginWithMfa,
      error: "Invalid verification code",
    });

    renderMfaLogin();

    const inputs = screen.getAllByRole("textbox");
    inputs.forEach((input) => {
      fireEvent.change(input, { target: { value: "9" } });
    });

    const submitBtn = screen.getByRole("button", { name: /verify/i });
    fireEvent.click(submitBtn);

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

    const submitBtn = screen.getByRole("button", { name: /verify/i });
    expect(submitBtn).toBeDisabled();

    const inputs = screen.getAllByRole("textbox");
    // Only fill 3 digits
    inputs.slice(0, 3).forEach((input, i) => {
      fireEvent.change(input, { target: { value: String(i + 1) } });
    });

    expect(submitBtn).toBeDisabled();
  });

  it("enables submit button when code is complete", () => {
    renderMfaLogin();

    const inputs = screen.getAllByRole("textbox");
    inputs.forEach((input, i) => {
      fireEvent.change(input, { target: { value: String(i + 1) } });
    });

    const submitBtn = screen.getByRole("button", { name: /verify/i });
    expect(submitBtn).not.toBeDisabled();
  });

  it("shows loading state during submission", () => {
    useAuthStore.setState({ loading: true });
    renderMfaLogin();

    const inputs = screen.getAllByRole("textbox");
    inputs.forEach((input, i) => {
      fireEvent.change(input, { target: { value: String(i + 1) } });
    });

    expect(screen.getByText(/Verifying.../i)).toBeInTheDocument();
  });
});
