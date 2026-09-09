import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import MfaRecover from "../MfaRecover";

function renderRecover() {
  return render(
    <MemoryRouter>
      <MfaRecover />
    </MemoryRouter>,
  );
}

async function typeRecoveryCode(code: string, user = userEvent.setup()) {
  await user.type(screen.getByPlaceholderText(/recovery code/i), code);
  return user;
}

beforeEach(() => {
  vi.clearAllMocks();
  server.use(
    http.put(
      "*/api/user/mfa/recovery/disable",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
  useAuthStore.setState({
    user: "admin",
    loading: false,
    error: null,
    recoverWithCode: vi.fn(),
    mfaRecoveryExpiry: null,
  });
});

describe("MfaRecover", () => {
  it("submits recovery code successfully", async () => {
    const mockRecover = vi.fn().mockResolvedValue(undefined);
    useAuthStore.setState({ recoverWithCode: mockRecover });

    renderRecover();

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "ABC-123-XYZ",
    );
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(mockRecover).toHaveBeenCalledWith("ABC-123-XYZ", "admin");
    });
  });

  it("displays error message on invalid code", async () => {
    const mockRecover = vi.fn().mockImplementation(async () => {
      useAuthStore.setState({ error: "Invalid recovery code or username" });
      throw new Error("Invalid recovery code or username");
    });
    useAuthStore.setState({ recoverWithCode: mockRecover });

    renderRecover();

    const user = userEvent.setup();
    await user.type(screen.getByPlaceholderText(/recovery code/i), "BAD-CODE");
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(
        screen.getByText("Invalid recovery code or username"),
      ).toBeInTheDocument();
    });
  });

  it("disables submit button when input is empty", () => {
    renderRecover();

    expect(screen.getByRole("button", { name: /recover/i })).toBeDisabled();
  });

  it("enables submit button after typing a valid code", async () => {
    renderRecover();

    await typeRecoveryCode("ABC-123-XYZ");

    expect(screen.getByRole("button", { name: /recover/i })).not.toBeDisabled();
  });

  it("shows loading state during submission", () => {
    useAuthStore.setState({ loading: true });

    renderRecover();

    expect(screen.getByText(/Recovering.../i)).toBeInTheDocument();
  });

  it("has link back to MFA login", () => {
    renderRecover();

    expect(screen.getByText(/Back to verification/i)).toHaveAttribute(
      "href",
      "/mfa-login",
    );
  });

  it("shows timeout modal after successful recovery", async () => {
    const futureTime = Math.floor(Date.now() / 1000) + 300;
    const mockRecover = vi.fn().mockResolvedValue(undefined);
    useAuthStore.setState({
      recoverWithCode: mockRecover,
      mfaRecoveryExpiry: futureTime,
    });

    renderRecover();

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "ABC-123-XYZ",
    );
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(mockRecover).toHaveBeenCalled();
    });
    expect(
      await screen.findByText(/Recovery Window Active/i),
    ).toBeInTheDocument();
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
      </MemoryRouter>,
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
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.queryByText("Login Page")).not.toBeInTheDocument();
    });
  });

  it("clears the recovery code field after a failed submission", async () => {
    const mockRecover = vi.fn().mockRejectedValue(new Error("Invalid"));
    useAuthStore.setState({ recoverWithCode: mockRecover });

    renderRecover();

    const user = userEvent.setup();
    const input = screen.getByPlaceholderText(/recovery code/i);
    await user.type(input, "BAD-CODE");
    await user.click(screen.getByRole("button", { name: /recover/i }));

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
  });

  it("clears the recovery code field after a successful submission", async () => {
    const mockRecover = vi.fn().mockResolvedValue(undefined);
    useAuthStore.setState({ recoverWithCode: mockRecover });

    renderRecover();

    const user = userEvent.setup();
    const input = screen.getByPlaceholderText(/recovery code/i);
    await user.type(input, "ABC-123-XYZ");
    await user.click(screen.getByRole("button", { name: /recover account/i }));

    await waitFor(() => {
      expect(input).toHaveValue("");
    });
  });

  it("calls the disable MFA endpoint during recovery window", async () => {
    const disableCalled = vi.fn();
    server.use(
      http.put("*/api/user/mfa/recovery/disable", () => {
        disableCalled();
        return new HttpResponse(null, { status: 204 });
      }),
    );
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
      </MemoryRouter>,
    );

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "MY-RECOVERY-CODE",
    );
    await user.click(screen.getByRole("button", { name: /recover account/i }));

    await waitFor(() => {
      expect(screen.getByText(/Recovery Window Active/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /disable mfa/i }));

    await waitFor(() => {
      expect(disableCalled).toHaveBeenCalled();
    });
  });

  it("navigates to dashboard and updates MFA status after successful disable", async () => {
    const futureExpiry = Math.floor(Date.now() / 1000) + 600;
    const mockRecover = vi.fn().mockImplementation(async () => {
      useAuthStore.setState({ mfaRecoveryExpiry: futureExpiry });
    });
    const mockUpdateMfaStatus = vi.fn();
    useAuthStore.setState({
      recoverWithCode: mockRecover,
      updateMfaStatus: mockUpdateMfaStatus,
    });

    render(
      <MemoryRouter initialEntries={["/recover-mfa"]}>
        <Routes>
          <Route path="/recover-mfa" element={<MfaRecover />} />
          <Route path="/dashboard" element={<div>Dashboard</div>} />
        </Routes>
      </MemoryRouter>,
    );

    const user = userEvent.setup();
    await user.type(
      screen.getByPlaceholderText(/recovery code/i),
      "MY-RECOVERY-CODE",
    );
    await user.click(screen.getByRole("button", { name: /recover account/i }));

    await waitFor(() => {
      expect(screen.getByText(/Recovery Window Active/i)).toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: /disable mfa/i }));

    await waitFor(() => {
      expect(mockUpdateMfaStatus).toHaveBeenCalledWith(false);
    });
    await waitFor(() => {
      expect(screen.getByText("Dashboard")).toBeInTheDocument();
    });
  });
});
