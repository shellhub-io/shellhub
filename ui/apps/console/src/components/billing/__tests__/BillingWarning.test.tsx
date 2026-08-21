import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import "@/components/common/__tests__/helpers/setup-dialog";
import BillingWarning from "../BillingWarning";

const mockNavigate = vi.fn();

vi.mock("react-router-dom", async () => {
  const actual =
    await vi.importActual<typeof import("react-router-dom")>(
      "react-router-dom",
    );
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

function renderWithRouter(ui: React.ReactNode) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe("BillingWarning", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ role: "owner" });
  });

  it("navigates to /settings#billing when the owner confirms", () => {
    const onClose = vi.fn();
    renderWithRouter(<BillingWarning open onClose={onClose} />);

    fireEvent.click(screen.getByRole("button", { name: /go to billing/i }));

    expect(mockNavigate).toHaveBeenCalledWith("/settings#billing");
    expect(onClose).toHaveBeenCalled();
  });

  it("does not navigate for non-owners and just closes", () => {
    useAuthStore.setState({ role: "observer" });
    const onClose = vi.fn();
    renderWithRouter(<BillingWarning open onClose={onClose} />);

    fireEvent.click(screen.getByRole("button", { name: /^close$/i }));

    expect(mockNavigate).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it("shows an owner-scoped description for owners", () => {
    renderWithRouter(<BillingWarning open onClose={() => {}} />);
    expect(
      screen.getByText(/subscribe to shellhub cloud/i),
    ).toBeInTheDocument();
  });

  it("shows a non-owner description for members", () => {
    useAuthStore.setState({ role: "observer" });
    renderWithRouter(<BillingWarning open onClose={() => {}} />);
    expect(screen.getByText(/ask the namespace owner/i)).toBeInTheDocument();
  });
});
