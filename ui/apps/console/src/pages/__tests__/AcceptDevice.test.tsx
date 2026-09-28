import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLocation } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import {
  PENDING_DEVICE_CODE_KEY,
  hasPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import AcceptDevice from "../AcceptDevice";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";

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

function setResolveCode(device: ReturnType<typeof mockDevice>) {
  server.use(
    http.get("*/api/devices/login-code/:code", () => HttpResponse.json(device)),
  );
}

beforeEach(() => {
  vi.resetAllMocks();
  localStorage.removeItem(PENDING_DEVICE_CODE_KEY);
  useAuthStore.setState({ tenant: "tenant1" });
  server.use(
    http.get("*/api/devices/login-code/:code", () =>
      HttpResponse.json(mockDevice()),
    ),
    http.patch(
      "*/api/devices/:uid/accept",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.post("*/api/devices/pairing/:code/accept", () =>
      HttpResponse.json({
        uid: "new-uid",
        tenant_id: "t1",
        namespace: "my-ns",
        owner_id: "user-1",
      }),
    ),
    http.get("*/api/namespaces", () =>
      jsonWithTotal([{ name: "my-ns", tenant_id: "t1" }]),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
  );
});

function CurrentLocation() {
  const location = useLocation();
  return (
    <output data-testid="location">
      {location.pathname + location.search}
    </output>
  );
}

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
    server.use(
      http.get("*/api/devices/login-code/:code", () => new Promise(() => {})),
    );
    renderFlow();

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.getByText("Checking code...")).toBeInTheDocument();
  });

  it("shows device preview for a login code", async () => {
    renderFlow();

    expect(
      await screen.findByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("dev1")).toBeInTheDocument();
    expect(screen.getByText("Ubuntu 22.04")).toBeInTheDocument();
    expect(screen.getByText("00:11:22:33:44:55")).toBeInTheDocument();
  });

  it("shows namespace picker for a pairing code", async () => {
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();

    expect(await screen.findByText("my-ns")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/choose where it belongs/i)).toBeInTheDocument();
  });

  it("shows already-accepted state", async () => {
    setResolveCode(mockDevice({ status: "accepted" }));
    renderFlow();

    expect(
      await screen.findByRole("heading", { name: /already accepted/i }),
    ).toBeInTheDocument();
  });

  it("transitions to success after accepting a device", async () => {
    renderFlow();

    fireEvent.click(
      await screen.findByRole("button", { name: /accept device/i }),
    );

    await screen.findByRole("heading", { name: /device accepted/i });
  });

  it("shows accept error without leaving ready state", async () => {
    server.use(
      http.patch("*/api/devices/:uid/accept", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    renderFlow();

    fireEvent.click(
      await screen.findByRole("button", { name: /accept device/i }),
    );

    await screen.findByRole("alert");
    expect(
      screen.getByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
  });

  it("resets to code form via 'Enter another code' on error", async () => {
    server.use(
      http.get("*/api/devices/login-code/:code", () =>
        HttpResponse.json({}, { status: 404 }),
      ),
    );
    renderFlow({ initialCode: "BADCODE1" });

    await screen.findByRole("heading", { name: /invalid or expired code/i });
    fireEvent.click(
      screen.getByRole("button", { name: /enter another code/i }),
    );

    expect(await screen.findByText("Claim a device")).toBeInTheDocument();
  });

  it("resolves code entered from the manual form", async () => {
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
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();

    await screen.findByText("my-ns");
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /accept device/i }),
      ).toBeEnabled();
    });
    fireEvent.click(screen.getByRole("button", { name: /accept device/i }));

    await screen.findByRole("heading", { name: /device accepted/i });
  });

  it("says a paired device is tied to the member who accepts it", async () => {
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();

    await screen.findByText("my-ns");
    expect(screen.getByText(/you.ll own this device/i)).toBeInTheDocument();
  });

  it("shows the pairing code to check against the terminal", async () => {
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow({ initialCode: "wxyz2k7q" });

    await screen.findByText("my-ns");
    expect(screen.getByText("WXYZ-2K7Q")).toBeInTheDocument();
  });

  it("shows the account the device will be tied to", async () => {
    useAuthStore.setState({ name: "Ada", email: "ada@example.com" });
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();

    const owner = await screen.findByRole("region", { name: "Accepting as" });
    expect(owner).toHaveTextContent("ada@example.com");
    expect(owner).toHaveTextContent(/you.ll own this device/i);
  });

  it("switches account by signing out and coming back to the same code", async () => {
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    render(
      <>
        <AcceptDeviceFlow initialCode="CODE1234" />
        <CurrentLocation />
      </>,
      {
        wrapper: createTestWrapper({
          initialEntries: ["/accept-device?code=CODE1234"],
        }),
      },
    );
    const user = userEvent.setup();

    await user.click(
      await screen.findByRole("button", { name: /not you\? switch account/i }),
    );

    expect(useAuthStore.getState().token).toBeFalsy();
    const location = new URL(
      screen.getByTestId("location").textContent ?? "",
      "http://x",
    );
    expect(location.pathname).toBe("/login");
    expect(location.searchParams.get("redirect")).toBe(
      "/accept-device?code=CODE1234",
    );
  });

  it("accepts into the namespace chosen from the list", async () => {
    let tenant: unknown;
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([
          { name: "my-ns", tenant_id: "t1" },
          { name: "other-ns", tenant_id: "t2" },
        ]),
      ),
      http.post("*/api/devices/pairing/:code/accept", async ({ request }) => {
        tenant = ((await request.json()) as { tenant_id: string }).tenant_id;
        return HttpResponse.json({ uid: "u", tenant_id: "t2", namespace: "other-ns", owner_id: "user-1" });
      }),
    );
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();
    const user = userEvent.setup();

    await user.click(await screen.findByRole("button", { name: "Namespace: my-ns" }));
    await user.click(await screen.findByRole("menuitemradio", { name: /other-ns/ }));
    await user.click(screen.getByRole("button", { name: /accept device/i }));

    await screen.findByRole("heading", { name: /device accepted/i });
    expect(tenant).toBe("t2");
  });

  it("offers no account switch inside the add-device dialog", async () => {
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow({ inDialog: true });

    await screen.findByRole("region", { name: "Accepting as" });
    expect(
      screen.queryByRole("button", { name: /switch account/i }),
    ).not.toBeInTheDocument();
  });

  async function acceptPairingAnsweredWith(ownerId?: string) {
    server.use(
      http.post("*/api/devices/pairing/:code/accept", () =>
        HttpResponse.json({
          uid: "new-uid",
          tenant_id: "t1",
          namespace: "my-ns",
          owner_id: ownerId,
        }),
      ),
    );
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();
    const user = userEvent.setup();

    await screen.findByText("my-ns");
    const accept = screen.getByRole("button", { name: /accept device/i });
    await waitFor(() => expect(accept).toBeEnabled());
    await user.click(accept);

    await screen.findByRole("heading", { name: /device accepted/i });
  }

  it("says nothing about the team when the device is tied to the member", async () => {
    await acceptPairingAnsweredWith("user-1");

    expect(screen.queryByText(/stays the team's/i)).not.toBeInTheDocument();
  });

  it("says the device stayed the team's when it merged into a team device", async () => {
    await acceptPairingAnsweredWith(undefined);

    expect(screen.getByText(/stays the team's/i)).toBeInTheDocument();
  });

  it("clears pending device code on error", async () => {
    setPendingDeviceCode("STALE");
    server.use(
      http.get("*/api/devices/login-code/:code", () =>
        HttpResponse.json({}, { status: 404 }),
      ),
    );
    renderFlow({ initialCode: "BADCODE1" });

    await screen.findByText(/invalid or expired code/i);
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });

  it("clears pending device code when already accepted", async () => {
    setPendingDeviceCode("STALE");
    setResolveCode(mockDevice({ status: "accepted" }));
    renderFlow();

    await screen.findByRole("heading", { name: /already accepted/i });
    expect(localStorage.getItem(PENDING_DEVICE_CODE_KEY)).toBeNull();
  });
});

describe("AcceptDeviceFlow dashboard link", () => {
  async function renderDeadEnd(
    state: "missing-code" | "error",
    inDialog: boolean,
  ) {
    if (state === "missing-code") {
      renderFlow({ initialCode: "", inDialog });
      return;
    }
    server.use(
      http.get("*/api/devices/login-code/:code", () =>
        HttpResponse.json({}, { status: 404 }),
      ),
    );
    renderFlow({ initialCode: "BADCODE1", inDialog });
    await screen.findByRole("heading", { name: /invalid or expired code/i });
  }

  it.each([
    ["missing-code", false, true],
    ["error", false, true],
    ["missing-code", true, false],
    ["error", true, false],
  ] as const)(
    "%s state with inDialog=%s offers a dashboard link: %s",
    async (state, inDialog, offered) => {
      await renderDeadEnd(state, inDialog);
      const link = screen.queryByRole("link", { name: /go to dashboard/i });
      if (offered) expect(link).toHaveAttribute("href", "/dashboard");
      else expect(link).not.toBeInTheDocument();
    },
  );
});

describe("AcceptDeviceFlow dialog mode", () => {
  it("resets to form via 'Use a different code' on ready state", async () => {
    renderFlow({ inDialog: true });

    fireEvent.click(
      await screen.findByRole("button", { name: /use a different code/i }),
    );

    expect(await screen.findByText("Claim a device")).toBeInTheDocument();
  });
});
