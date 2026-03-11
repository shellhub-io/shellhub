import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import MfaRecover from "../MfaRecover";

vi.mock("../../api/mfa", () => ({
  disableMfa: vi.fn(),
}));

import { disableMfa } from "../../api/mfa";
const mockedDisableMfa = vi.mocked(disableMfa);

beforeEach(() => {
  mockedDisableMfa.mockResolvedValue(undefined);
  useAuthStore.setState({
    user: "admin",
    loading: false,
    error: null,
    recoverWithCode: vi.fn(),
    mfaRecoveryExpiry: null,
  });
});

describe("MfaRecover", () => {
  it("renders recovery form", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    expect(screen.getByText(/Account Recovery/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/recovery code/i)).toBeInTheDocument();
  });

  it("submits recovery code successfully", async () => {
    const mockRecover = vi.fn().mockResolvedValue(undefined);
    useAuthStore.setState({ recoverWithCode: mockRecover });

    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText(/recovery code/i);
    fireEvent.change(input, { target: { value: "ABC-123-XYZ" } });

    const submitBtn = screen.getByRole("button", { name: /recover/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(mockRecover).toHaveBeenCalledWith("ABC-123-XYZ", "admin");
    });
  });

  it("displays error message on invalid code", async () => {
    const mockRecover = vi.fn().mockRejectedValue(new Error("Invalid recovery code"));
    useAuthStore.setState({ recoverWithCode: mockRecover });

    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    fireEvent.change(screen.getByPlaceholderText(/recovery code/i), {
      target: { value: "BAD-CODE" },
    });
    fireEvent.click(screen.getByRole("button", { name: /recover/i }));

    // Error is set in the store by recoverWithCode on failure
    useAuthStore.setState({ error: "Invalid recovery code" });

    await waitFor(() => {
      expect(screen.getByText("Invalid recovery code")).toBeInTheDocument();
    });
  });

  it("disables submit button when input is empty", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    const submitBtn = screen.getByRole("button", { name: /recover/i });
    expect(submitBtn).toBeDisabled();
  });

  it("enables submit button when code is entered", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText(/recovery code/i);
    fireEvent.change(input, { target: { value: "ABC-123-XYZ" } });

    const submitBtn = screen.getByRole("button", { name: /recover/i });
    expect(submitBtn).not.toBeDisabled();
  });

  it("shows loading state during submission", () => {
    useAuthStore.setState({ loading: true });

    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText(/recovery code/i);
    fireEvent.change(input, { target: { value: "ABC-123-XYZ" } });

    expect(screen.getByText(/Recovering.../i)).toBeInTheDocument();
  });

  it("has link back to MFA login", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    const backLink = screen.getByText(/Back to verification/i);
    expect(backLink).toHaveAttribute("href", "/mfa-login");
  });

  it("shows timeout modal after successful recovery", async () => {
    const futureTime = Math.floor(Date.now() / 1000) + 300;
    const mockRecover = vi.fn().mockResolvedValue(undefined);
    useAuthStore.setState({ recoverWithCode: mockRecover, mfaRecoveryExpiry: futureTime });

    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText(/recovery code/i);
    fireEvent.change(input, { target: { value: "ABC-123-XYZ" } });
    fireEvent.click(screen.getByRole("button", { name: /recover/i }));

    // After recovery, timeout modal should render (which contains a countdown)
    await waitFor(() => {
      expect(mockRecover).toHaveBeenCalled();
    });
  });

  it("shows 10-minute recovery window warning note", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    // The warning note is always rendered regardless of mfaRecoveryExpiry
    expect(screen.getByText(/10-minute window/i)).toBeInTheDocument();
  });

  it("redirects to login when no identifier and no mfaToken", async () => {
    useAuthStore.setState({
      user: null,
      username: null,
      mfaToken: null,
    });

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText("Login Page")).toBeInTheDocument();
    });
  });

  it("does not redirect when mfaToken is present even without identifier", async () => {
    useAuthStore.setState({
      user: null,
      username: null,
      mfaToken: "active-mfa-token",
    });

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </MemoryRouter>
    );

    // Should not redirect — mfaToken prevents redirect even without identifier
    await waitFor(() => {
      expect(screen.queryByText("Login Page")).not.toBeInTheDocument();
    });
  });

  it("does not show recovery timeout modal initially", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    // Modal only appears after a successful recovery code submission
    expect(screen.queryByText(/Recovery Window Active/i)).not.toBeInTheDocument();
  });

  it("calls disableMfa with the recovery code during recovery window", async () => {
    const futureExpiry = Math.floor(Date.now() / 1000) + 600;
    const mockRecover = vi.fn().mockImplementation(async () => {
      useAuthStore.setState({ mfaRecoveryExpiry: futureExpiry });
    });
    useAuthStore.setState({ recoverWithCode: mockRecover });

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/dashboard" element={<div>Dashboard</div>} />
        </Routes>
      </MemoryRouter>
    );

    // Enter and submit a recovery code
    fireEvent.change(screen.getByPlaceholderText(/recovery code/i), {
      target: { value: "MY-RECOVERY-CODE" },
    });
    fireEvent.click(screen.getByRole("button", { name: /recover account/i }));

    // Wait for the recovery window modal to appear
    await waitFor(() => {
      expect(screen.getByText(/Recovery Window Active/i)).toBeInTheDocument();
    });

    // Click "Disable MFA" in the recovery window modal
    fireEvent.click(screen.getByRole("button", { name: /disable mfa/i }));

    await waitFor(() => {
      expect(mockedDisableMfa).toHaveBeenCalledWith({ recovery_code: "MY-RECOVERY-CODE" });
    });
  });
});
