import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import {
  PENDING_DEVICE_CODE_KEY,
  hasPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import AcceptDevice from "../AcceptDevice";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

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
import { resolveDeviceLoginCode } from "@/client";
const mockedResolve = vi.mocked(resolveDeviceLoginCode);

beforeEach(() => {
  mockNavigate.mockClear();
  mockedResolve.mockReset();
  localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
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
    expect(
      screen.getByRole("button", { name: /claim device/i }),
    ).toBeDisabled();
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

    expect(
      screen.getByRole("button", { name: /claim device/i }),
    ).toBeDisabled();
  });
});

describe("AcceptDevice page", () => {
  it("persists the code from the URL to localStorage", () => {
    renderAt("/accept-device?code=WXYZ2K7Q");
    expect(hasPendingDeviceCode()).toBe(true);
  });
});

describe("AcceptDeviceFlow pending device code", () => {
  it("navigates to /login without redirect param when unauthenticated", async () => {
    useAuthStore.setState({ token: null, tenant: "tenant1" });
    render(
      <MemoryRouter>
        <AcceptDeviceFlow code="WXYZ2K7Q" />
      </MemoryRouter>,
    );

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith("/login"));
  });

  it("navigates to /login on a 401 resolve error", async () => {
    useAuthStore.setState({ token: "stale-token", tenant: "tenant1" });
    const err = Object.assign(new Error(), { status: 401 });
    mockedResolve.mockRejectedValue(err);

    render(
      <MemoryRouter>
        <AcceptDeviceFlow code="ABC12345" />
      </MemoryRouter>,
    );

    await waitFor(() => expect(mockNavigate).toHaveBeenCalledWith("/login"));
  });

  it("clears the key when the flow reaches the error branch", async () => {
    setPendingDeviceCode("STALE");
    useAuthStore.setState({ token: "valid", tenant: "tenant1" });
    mockedResolve.mockRejectedValue(new Error("bad code"));

    render(
      <MemoryRouter>
        <AcceptDeviceFlow code="BADCODE1" />
      </MemoryRouter>,
    );

    await screen.findByText(/invalid or expired code/i);
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });

  it("clears the key when the device is already accepted", async () => {
    setPendingDeviceCode("STALE");
    useAuthStore.setState({ token: "valid", tenant: "tenant1" });
    mockedResolve.mockResolvedValue({
      data: {
        kind: "device",
        status: "accepted",
        name: "dev1",
        tenant_id: "tenant1",
      },
      request: new Request("http://localhost"),
      response: new Response(),
    });

    render(
      <MemoryRouter>
        <AcceptDeviceFlow code="CODE1234" />
      </MemoryRouter>,
    );

    await screen.findByRole("heading", { name: /already accepted/i });
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });
});
