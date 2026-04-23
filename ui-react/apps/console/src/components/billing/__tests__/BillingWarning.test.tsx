import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import "@/components/common/__tests__/helpers/setup-dialog";
import BillingWarning from "../BillingWarning";

const mockNavigate = vi.fn();
const mockHasPermission = vi.fn();

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

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: (perm: string) => Boolean(mockHasPermission(perm)),
}));

function renderWithRouter(ui: React.ReactNode) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("BillingWarning", () => {
  it("navigates to /settings#billing when the owner confirms", () => {
    mockHasPermission.mockReturnValue(true);
    const onClose = vi.fn();
    renderWithRouter(<BillingWarning open onClose={onClose} />);

    fireEvent.click(screen.getByRole("button", { name: /go to billing/i }));

    expect(mockNavigate).toHaveBeenCalledWith("/settings#billing");
    expect(onClose).toHaveBeenCalled();
  });

  it("does not navigate for non-owners and just closes", () => {
    mockHasPermission.mockReturnValue(false);
    const onClose = vi.fn();
    renderWithRouter(<BillingWarning open onClose={onClose} />);

    // For non-owners the confirm button reads "Close".
    fireEvent.click(screen.getByRole("button", { name: /^close$/i }));

    expect(mockNavigate).not.toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });

  it("shows an owner-scoped description for owners", () => {
    mockHasPermission.mockReturnValue(true);
    renderWithRouter(<BillingWarning open onClose={() => {}} />);
    expect(
      screen.getByText(/subscribe to shellhub cloud/i),
    ).toBeInTheDocument();
  });

  it("shows a non-owner description for members", () => {
    mockHasPermission.mockReturnValue(false);
    renderWithRouter(<BillingWarning open onClose={() => {}} />);
    expect(screen.getByText(/ask the namespace owner/i)).toBeInTheDocument();
  });
});
