import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import AcceptDevice from "../AcceptDevice";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// The query hooks are irrelevant to the manual-entry (missing-code) path; mock
// them so the page renders without a QueryClient provider.
vi.mock("@/hooks/useDeviceMutations", () => ({
  useAcceptDevice: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("@/hooks/useNamespaceMutations", () => ({
  useSwitchNamespace: () => ({ mutateAsync: vi.fn() }),
}));
vi.mock("@/hooks/useNamespaces", () => ({
  useNamespaces: () => ({ data: [], isLoading: false }),
}));
vi.mock("@/client", () => ({
  resolveDeviceLoginCode: vi.fn(),
  acceptDevicePairing: vi.fn(),
}));

beforeEach(() => {
  mockNavigate.mockClear();
  useAuthStore.setState({ token: "token", tenant: "tenant1" });
});

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AcceptDevice />
    </MemoryRouter>,
  );
}

describe("AcceptDevice manual entry", () => {
  it("shows the pairing-code form when opened without a code", async () => {
    renderAt("/accept-device");

    expect(await screen.findByText("Claim a device")).toBeInTheDocument();
    expect(screen.getAllByRole("textbox")).toHaveLength(8);
    expect(screen.getByRole("button", { name: /claim device/i })).toBeDisabled();
  });

  it("navigates with the canonical code once every cell is filled", async () => {
    renderAt("/accept-device");

    await screen.findByText("Claim a device");
    const cells = screen.getAllByRole("textbox");
    "VS3AMKME".split("").forEach((ch, i) => {
      fireEvent.change(cells[i], { target: { value: ch } });
    });

    const submit = screen.getByRole("button", { name: /claim device/i });
    expect(submit).not.toBeDisabled();
    fireEvent.click(submit);

    expect(mockNavigate).toHaveBeenCalledWith("/accept-device?code=VS3AMKME");
  });

  it("accepts a pasted code with its display hyphen", async () => {
    renderAt("/accept-device");

    await screen.findByText("Claim a device");
    const cells = screen.getAllByRole("textbox");
    fireEvent.paste(cells[0], {
      clipboardData: { getData: () => "vs3a-mkme" },
    });

    fireEvent.click(screen.getByRole("button", { name: /claim device/i }));
    expect(mockNavigate).toHaveBeenCalledWith("/accept-device?code=VS3AMKME");
  });

  it("keeps submit disabled until the code is complete", async () => {
    renderAt("/accept-device");

    await screen.findByText("Claim a device");
    const cells = screen.getAllByRole("textbox");
    "VS3A".split("").forEach((ch, i) => {
      fireEvent.change(cells[i], { target: { value: ch } });
    });

    expect(screen.getByRole("button", { name: /claim device/i })).toBeDisabled();
  });
});
