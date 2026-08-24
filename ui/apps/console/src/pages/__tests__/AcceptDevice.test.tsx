import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import {
  PENDING_DEVICE_CODE_KEY,
  hasPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import AcceptDevice from "../AcceptDevice";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    resolveDeviceLoginCode: vi.fn(),
    acceptDevice: vi.fn(),
    acceptDevicePairing: vi.fn(),
    getNamespaces: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

function respondWith<T>(data: T) {
  return { data } as never;
}

function mockDevice(overrides = {}) {
  return {
    kind: "device",
    status: "pending",
    name: "dev1",
    uid: "uid-1",
    tenant_id: "tenant1",
    namespace: "ns1",
    identity: { mac: "00:11:22:33:44:55" },
    info: { pretty_name: "Ubuntu 22.04" },
    ...overrides,
  };
}

beforeEach(() => {
  vi.resetAllMocks();
  localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
  useAuthStore.setState({ token: "token", tenant: "tenant1" });
});

function renderPage(path: string) {
  return render(<AcceptDevice />, {
    wrapper: createTestWrapper({ initialEntries: [path] }),
  });
}

function renderFlow({
  initialCode = "CODE1234",
  inDialog,
}: { initialCode?: string; inDialog?: boolean } = {}) {
  return render(
    <AcceptDeviceFlow initialCode={initialCode} inDialog={inDialog} />,
    { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
  );
}

describe("AcceptDevice page", () => {
  it("persists the code from the URL to localStorage", () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(respondWith(mockDevice()));
    renderPage("/accept-device?code=WXYZ2K7Q");
    expect(hasPendingDeviceCode()).toBe(true);
  });

  it("shows the pairing-code form when opened without a code", async () => {
    renderPage("/accept-device");

    expect(await screen.findByText("Claim a device")).toBeInTheDocument();
    expect(screen.getAllByRole("textbox")).toHaveLength(8);
    expect(
      screen.getByRole("button", { name: /claim device/i }),
    ).toBeDisabled();
  });
});

describe("AcceptDeviceFlow standalone", () => {
  it("shows loading while resolving a code", () => {
    sdk.resolveDeviceLoginCode.mockReturnValue(new Promise<never>(() => {}));
    renderFlow();

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.getByText("Checking code...")).toBeInTheDocument();
  });

  it("shows device preview for a login code", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(respondWith(mockDevice()));
    renderFlow();

    expect(
      await screen.findByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("dev1")).toBeInTheDocument();
    expect(screen.getByText("Ubuntu 22.04")).toBeInTheDocument();
    expect(screen.getByText("00:11:22:33:44:55")).toBeInTheDocument();
  });

  it("shows namespace picker for a pairing code", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(
      respondWith(mockDevice({ kind: "pairing", tenant_id: null })),
    );
    sdk.getNamespaces.mockResolvedValue(
      respondWith([{ name: "my-ns", tenant_id: "t1" }]),
    );
    renderFlow();

    expect(await screen.findByText("my-ns")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/choose where it belongs/i)).toBeInTheDocument();
  });

  it("shows error state on invalid/expired code", async () => {
    sdk.resolveDeviceLoginCode.mockRejectedValue(new Error("bad code"));
    renderFlow({ initialCode: "BADCODE1" });

    expect(
      await screen.findByRole("heading", { name: /invalid or expired code/i }),
    ).toBeInTheDocument();
  });

  it("shows already-accepted state", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(
      respondWith(mockDevice({ status: "accepted" })),
    );
    renderFlow();

    expect(
      await screen.findByRole("heading", { name: /already accepted/i }),
    ).toBeInTheDocument();
  });

  it("transitions to success after accepting a device", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(respondWith(mockDevice()));
    sdk.acceptDevice.mockResolvedValue(respondWith({}));
    renderFlow();

    fireEvent.click(
      await screen.findByRole("button", { name: /accept device/i }),
    );

    await screen.findByRole("heading", { name: /device accepted/i });
    expect(sdk.acceptDevice).toHaveBeenCalledWith(
      expect.objectContaining({ path: { uid: "uid-1" } }),
    );
  });

  it("shows accept error without leaving ready state", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(respondWith(mockDevice()));
    sdk.acceptDevice.mockRejectedValue(new Error("limit reached"));
    renderFlow();

    fireEvent.click(
      await screen.findByRole("button", { name: /accept device/i }),
    );

    await screen.findByRole("alert");
    expect(
      screen.getByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
  });

  it("shows dashboard link in missing-code state", () => {
    renderFlow({ initialCode: "" });

    expect(
      screen.getByRole("link", { name: /go to dashboard/i }),
    ).toHaveAttribute("href", "/dashboard");
  });

  it("shows dashboard link in error state", async () => {
    sdk.resolveDeviceLoginCode.mockRejectedValue(new Error("bad code"));
    renderFlow({ initialCode: "BADCODE1" });

    await screen.findByRole("heading", { name: /invalid or expired code/i });
    expect(
      screen.getByRole("link", { name: /go to dashboard/i }),
    ).toHaveAttribute("href", "/dashboard");
  });

  it("resets to code form via 'Enter another code' on error", async () => {
    sdk.resolveDeviceLoginCode.mockRejectedValue(new Error("bad code"));
    renderFlow({ initialCode: "BADCODE1" });

    await screen.findByRole("heading", { name: /invalid or expired code/i });
    fireEvent.click(
      screen.getByRole("button", { name: /enter another code/i }),
    );

    expect(await screen.findByText("Claim a device")).toBeInTheDocument();
  });

  it("resolves code entered from the manual form", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(respondWith(mockDevice()));
    renderFlow({ initialCode: "" });

    await screen.findByText("Claim a device");
    const cells = screen.getAllByRole("textbox");
    "VS3AMKME".split("").forEach((ch, i) => {
      fireEvent.change(cells[i], { target: { value: ch } });
    });

    fireEvent.click(screen.getByRole("button", { name: /claim device/i }));

    await screen.findByRole("heading", { name: /accept this device/i });
  });

  it("transitions to pairing-success after accepting with a namespace", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(
      respondWith(mockDevice({ kind: "pairing", tenant_id: null })),
    );
    sdk.getNamespaces.mockResolvedValue(
      respondWith([{ name: "my-ns", tenant_id: "t1" }]),
    );
    sdk.acceptDevicePairing.mockResolvedValue(
      respondWith({ uid: "new-uid", tenant_id: "t1", namespace: "my-ns" }),
    );
    renderFlow();

    await screen.findByText("my-ns");
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /accept device/i }),
      ).toBeEnabled();
    });
    fireEvent.click(screen.getByRole("button", { name: /accept device/i }));

    await screen.findByRole("heading", { name: /device accepted/i });
    expect(sdk.acceptDevicePairing).toHaveBeenCalledWith(
      expect.objectContaining({
        path: { code: "CODE1234" },
        body: { tenant_id: "t1" },
      }),
    );
  });

  it("clears pending device code on error", async () => {
    setPendingDeviceCode("STALE");
    sdk.resolveDeviceLoginCode.mockRejectedValue(new Error("bad code"));
    renderFlow({ initialCode: "BADCODE1" });

    await screen.findByText(/invalid or expired code/i);
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });

  it("clears pending device code when already accepted", async () => {
    setPendingDeviceCode("STALE");
    sdk.resolveDeviceLoginCode.mockResolvedValue(
      respondWith(mockDevice({ status: "accepted" })),
    );
    renderFlow();

    await screen.findByRole("heading", { name: /already accepted/i });
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });
});

describe("AcceptDeviceFlow dialog mode", () => {
  it("shows code entry form when no code provided", () => {
    renderFlow({ initialCode: "", inDialog: true });

    expect(screen.getByText("Claim a device")).toBeInTheDocument();
    expect(screen.getAllByRole("textbox")).toHaveLength(8);
  });

  it("does not show dashboard link in missing-code state", () => {
    renderFlow({ initialCode: "", inDialog: true });

    expect(
      screen.queryByRole("link", { name: /go to dashboard/i }),
    ).not.toBeInTheDocument();
  });

  it("does not show dashboard link in error state", async () => {
    sdk.resolveDeviceLoginCode.mockRejectedValue(new Error("bad code"));
    renderFlow({ initialCode: "BADCODE1", inDialog: true });

    await screen.findByRole("heading", { name: /invalid or expired code/i });
    expect(
      screen.queryByRole("link", { name: /go to dashboard/i }),
    ).not.toBeInTheDocument();
  });

  it("resets to form via 'Use a different code' on ready state", async () => {
    sdk.resolveDeviceLoginCode.mockResolvedValue(respondWith(mockDevice()));
    renderFlow({ inDialog: true });

    fireEvent.click(
      await screen.findByRole("button", { name: /use a different code/i }),
    );

    expect(await screen.findByText("Claim a device")).toBeInTheDocument();
  });
});
