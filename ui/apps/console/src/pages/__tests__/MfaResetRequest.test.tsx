import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { useMfaResetStore } from "@/stores/mfaResetStore";
import MfaResetRequest from "../MfaResetRequest";

function renderPage() {
  return render(
    <MemoryRouter initialEntries={["/mfa-reset-request"]}>
      <Routes>
        <Route path="/mfa-reset-request" element={<MfaResetRequest />} />
        <Route path="/mfa-reset-verify" element={<div>Verify Page</div>} />
        <Route path="/login" element={<div>Login Page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  useAuthStore.setState({ user: "admin", username: null, mfaToken: "tok" });
  useMfaResetStore.setState({
    loading: false,
    error: null,
    requestMfaReset: vi.fn().mockResolvedValue(undefined),
  });
});

describe("MfaResetRequest", () => {
  it("calls requestMfaReset with the identifier when the form is submitted", async () => {
    const mockRequest = vi.fn().mockResolvedValue(undefined);
    useMfaResetStore.setState({ requestMfaReset: mockRequest });

    renderPage();

    const user = userEvent.setup();
    await user.click(
      screen.getByRole("button", { name: /send verification codes/i }),
    );

    await waitFor(() => {
      expect(mockRequest).toHaveBeenCalledWith("admin");
    });
  });

  it("navigates to /mfa-reset-verify after a successful submit", async () => {
    renderPage();

    const user = userEvent.setup();
    await user.click(
      screen.getByRole("button", { name: /send verification codes/i }),
    );

    await waitFor(() => {
      expect(screen.getByText("Verify Page")).toBeInTheDocument();
    });
  });

  it("shows the store error in a Callout when the request fails", async () => {
    const mockRequest = vi.fn().mockImplementation(async () => {
      useMfaResetStore.setState({
        error: "Unable to send reset emails. Please check your identifier.",
      });
      throw new Error("Reset request failed");
    });
    useMfaResetStore.setState({ requestMfaReset: mockRequest, error: null });

    renderPage();

    const user = userEvent.setup();
    await user.click(
      screen.getByRole("button", { name: /send verification codes/i }),
    );

    await waitFor(() => {
      expect(
        screen.getByText(/Unable to send reset emails/i),
      ).toBeInTheDocument();
    });
  });

  it("redirects to login when there is no identifier and no mfaToken", async () => {
    useAuthStore.setState({ user: null, username: null, mfaToken: null });

    renderPage();

    await waitFor(() => {
      expect(screen.getByText("Login Page")).toBeInTheDocument();
    });
  });

  it("renders nothing when there is no identifier but an active mfaToken", () => {
    useAuthStore.setState({
      user: null,
      username: null,
      mfaToken: "active-tok",
    });

    const { container } = renderPage();

    expect(container.firstChild).toBeNull();
  });
});
