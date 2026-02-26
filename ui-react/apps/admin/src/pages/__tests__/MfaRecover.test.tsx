import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import MfaRecover from "../MfaRecover";

beforeEach(() => {
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

  it("displays error message on invalid code", () => {
    useAuthStore.setState({
      error: "Invalid recovery code",
    });

    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    expect(screen.getByText("Invalid recovery code")).toBeInTheDocument();
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

  it("does not show timeout modal initially", () => {
    render(
      <MemoryRouter>
        <MfaRecover />
      </MemoryRouter>
    );

    // Countdown format "X minutes Y seconds" should not appear in the main body
    expect(screen.queryByText(/\d+ minutes? \d+ seconds?/)).not.toBeInTheDocument();
  });
});
