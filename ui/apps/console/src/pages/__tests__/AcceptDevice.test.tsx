import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLocation } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import {
  mockNamespace,
  mockPairingAccepted,
  mockUserAuth,
} from "@/tests/factories";
import { setEdition } from "@/tests/edition";
import {
  choiceElement,
  mockSurveyPayload,
  surveyEnvironment,
} from "@/tests/onboardingSurvey";
import { defaultConfig, getConfig } from "@/env";
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
  seedAuthStore({ tenant: "tenant1" });
  server.use(
    http.get("*/api/devices/login-code/:code", () =>
      HttpResponse.json(mockDevice()),
    ),
    http.patch(
      "*/api/devices/:uid/accept",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.post("*/api/devices/pairing/:code/accept", () =>
      HttpResponse.json(mockPairingAccepted()),
    ),
    http.get("*/api/namespaces", () =>
      jsonWithTotal([{ name: "my-ns", tenant_id: "t1" }]),
    ),
    http.get("*/api/namespaces/:tenant", ({ params }) =>
      HttpResponse.json(mockNamespace({ tenant_id: String(params.tenant) })),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: VALID_JWT, role: "owner" }),
    ),
    http.get("*/api/vault", () => new HttpResponse(null, { status: 404 })),
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
        return HttpResponse.json({
          uid: "u",
          tenant_id: "t2",
          namespace: "other-ns",
          owner_id: "user-1",
        });
      }),
    );
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();
    const user = userEvent.setup();

    await user.click(
      await screen.findByRole("button", { name: "Namespace: my-ns" }),
    );
    await user.click(
      await screen.findByRole("menuitemradio", { name: /other-ns/ }),
    );
    await user.click(screen.getByRole("button", { name: /accept device/i }));

    await screen.findByRole("heading", { name: /device accepted/i });
    expect(tenant).toBe("t2");
  });

  it("leaves legacy namespaces out of the namespace picker", async () => {
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([
          { name: "my-ns", tenant_id: "t1", settings: { legacy: false } },
          { name: "old-ns", tenant_id: "t2", settings: { legacy: true } },
          { name: "other-ns", tenant_id: "t3", settings: { legacy: false } },
        ]),
      ),
    );
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();
    const user = userEvent.setup();

    await user.click(
      await screen.findByRole("button", { name: "Namespace: my-ns" }),
    );

    expect(
      await screen.findByRole("menuitemradio", { name: /other-ns/ }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("menuitemradio", { name: /old-ns/ }),
    ).not.toBeInTheDocument();
  });

  it("explains when every namespace is legacy", async () => {
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([
          { name: "old-ns", tenant_id: "t2", settings: { legacy: true } },
        ]),
      ),
    );
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    renderFlow();

    expect(
      await screen.findByText(/every namespace you belong to is legacy/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^Namespace:/ }),
    ).not.toBeInTheDocument();
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
        HttpResponse.json(mockPairingAccepted({ owner_id: ownerId })),
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

describe("AcceptDeviceFlow signed out", () => {
  beforeEach(() => {
    useAuthStore.setState({ token: null, tenant: null });
    server.use(
      http.post("*/api/login", () =>
        HttpResponse.json(mockUserAuth({ tenant: "tenant1" })),
      ),
    );
  });

  it("offers to create an account next to signing in on cloud", () => {
    setEdition("cloud");
    renderFlow({ initialCode: "WXYZ2K7Q" });

    expect(
      screen.getByRole("link", { name: /^create account$/i }),
    ).toHaveAttribute("href", "/sign-up");
  });

  it.each(["community", "enterprise"] as const)(
    "offers no account creation on %s",
    (edition) => {
      setEdition(edition);
      renderFlow({ initialCode: "WXYZ2K7Q" });

      expect(
        screen.queryByRole("link", { name: /create account|sign up/i }),
      ).not.toBeInTheDocument();
    },
  );

  it("asks to sign in while showing the code to check", () => {
    renderFlow({ initialCode: "WXYZ2K7Q" });

    expect(
      screen.getByRole("heading", { name: /sign in to accept this device/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("WXYZ-2K7Q")).toBeInTheDocument();
  });

  it("continues to the device review in place once signed in", async () => {
    const user = userEvent.setup();
    setPendingDeviceCode("WXYZ2K7Q");
    renderFlow({ initialCode: "WXYZ2K7Q" });

    await user.type(screen.getByLabelText(/username/i), "admin");
    await user.type(screen.getByLabelText(/^password$/i), "secret");
    await user.click(
      screen.getByRole("button", { name: /sign in and continue/i }),
    );

    expect(
      await screen.findByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("dev1")).toBeInTheDocument();
    expect(hasPendingDeviceCode()).toBe(false);
  });

  it("asks for the second factor in place and continues to the device", async () => {
    server.use(
      http.post("*/api/login", () =>
        HttpResponse.json(mockUserAuth(), {
          status: 401,
          headers: { "x-mfa-token": "mfa-temp" },
        }),
      ),
      http.post("*/api/user/mfa/auth", () =>
        HttpResponse.json(mockUserAuth({ tenant: "tenant1" })),
      ),
    );
    const user = userEvent.setup();
    renderFlow({ initialCode: "WXYZ2K7Q" });

    await user.type(screen.getByLabelText(/username/i), "admin");
    await user.type(screen.getByLabelText(/^password$/i), "secret");
    await user.click(
      screen.getByRole("button", { name: /sign in and continue/i }),
    );
    for (const [i, cell] of (await screen.findAllByRole("textbox")).entries()) {
      await user.type(cell, String(i + 1));
    }
    await user.click(
      screen.getByRole("button", { name: /sign in and continue/i }),
    );

    expect(
      await screen.findByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
  });
});

describe("AcceptDeviceFlow first namespace", () => {
  const SURVEY_API = "https://forms.example.test/api/v1/client/ws1";
  const acceptedIntoNewNamespace = mockPairingAccepted({
    tenant_id: "t-new",
    namespace: "joaocel",
    owner_id: "user-123",
  });
  let surveySent: unknown[];

  async function createAndAccept(user: ReturnType<typeof userEvent.setup>) {
    await user.click(
      await screen.findByRole("button", { name: /^create and accept$/i }),
    );
    await screen.findByRole("heading", { name: /device accepted/i });
  }

  beforeEach(() => {
    surveySent = [];
    vi.mocked(getConfig).mockReturnValue({
      ...defaultConfig,
      edition: "cloud",
      onboardingUrl: SURVEY_API,
    });
    seedAuthStore({ tenant: null, userId: "user-123" });
    setResolveCode(mockDevice({ kind: "pairing", tenant_id: null }));
    server.use(
      http.get("*/api/namespaces", () => jsonWithTotal([])),
      http.get("*/api/auth/user", () => HttpResponse.json(mockUserAuth())),
      http.get(`${SURVEY_API}/environment`, () =>
        HttpResponse.json(
          surveyEnvironment(
            mockSurveyPayload({
              elements: [
                choiceElement("role", "What's your role?", [
                  ["dev", "Developer"],
                ]),
                choiceElement("use", "What will you use ShellHub for?", [
                  ["iot", "A fleet of IoT devices"],
                ]),
              ],
            }),
          ),
        ),
      ),
      http.post(`${SURVEY_API}/responses`, async ({ request }) => {
        surveySent.push(await request.json());
        return HttpResponse.json({ data: { id: "response-1" } });
      }),
      http.post("*/api/namespaces", () =>
        HttpResponse.json(
          mockNamespace({ name: "joaocel", tenant_id: "t-new" }),
        ),
      ),
      http.post("*/api/devices/pairing/:code/accept", () =>
        HttpResponse.json(acceptedIntoNewNamespace),
      ),
    );
    useAuthStore.setState({ user: "joaocel", username: null });
  });

  it("offers the way into the device after accepting into a first namespace, with no survey", async () => {
    const user = userEvent.setup();
    renderFlow();

    await createAndAccept(user);

    expect(
      await screen.findByRole("button", { name: /view device/i }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("radio")).not.toBeInTheDocument();
    expect(surveySent).toEqual([]);
  });

  it("offers no namespace to create when the namespaces cannot be read", async () => {
    server.use(
      http.get("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    renderFlow();

    expect(
      await screen.findByText(/your namespaces could not be loaded/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^create and accept$/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^accept device$/i }),
    ).not.toBeInTheDocument();
  });

  it("offers no Accept while the namespaces are loading", async () => {
    server.use(http.get("*/api/namespaces", () => new Promise(() => {})));
    renderFlow();

    expect(await screen.findByText(/loading namespaces/i)).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^accept device$/i }),
    ).not.toBeInTheDocument();
  });

  it("keeps the device on screen while creating the namespace and accepting into it", async () => {
    let resolves = 0;
    let acceptSent = false;
    server.use(
      http.get("*/api/devices/login-code/:code", () => {
        resolves += 1;
        return resolves === 1
          ? HttpResponse.json(mockDevice({ kind: "pairing", tenant_id: null }))
          : new Promise(() => {});
      }),
      http.post("*/api/devices/pairing/:code/accept", () => {
        acceptSent = true;
        return new Promise(() => {});
      }),
    );
    const user = userEvent.setup();
    renderFlow();

    await user.click(
      await screen.findByRole("button", { name: /^create and accept$/i }),
    );
    await waitFor(() => expect(acceptSent).toBe(true));

    expect(screen.queryByText("Checking code...")).not.toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /accept this device/i }),
    ).toBeInTheDocument();
  });

  it("shows the new namespace when it was created but could not be entered", async () => {
    let namespaces: { name: string; tenant_id: string }[] = [];
    server.use(
      http.get("*/api/namespaces", () => jsonWithTotal(namespaces)),
      http.post("*/api/namespaces", () => {
        namespaces = [{ name: "joaocel", tenant_id: "t-new" }];
        return HttpResponse.json(
          mockNamespace({ name: "joaocel", tenant_id: "t-new" }),
        );
      }),
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderFlow();

    await user.click(
      await screen.findByRole("button", { name: /^create and accept$/i }),
    );

    expect(
      await screen.findByRole("button", { name: "Namespace: joaocel" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^create and accept$/i }),
    ).not.toBeInTheDocument();
  });

  it("keeps the new namespace and offers Accept again when accepting into it fails", async () => {
    let namespaces: { name: string; tenant_id: string }[] = [];
    server.use(
      http.get("*/api/namespaces", () => jsonWithTotal(namespaces)),
      http.post("*/api/namespaces", () => {
        namespaces = [{ name: "joaocel", tenant_id: "t-new" }];
        return HttpResponse.json(
          mockNamespace({ name: "joaocel", tenant_id: "t-new" }),
        );
      }),
      http.post("*/api/devices/pairing/:code/accept", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderFlow();

    await user.click(
      await screen.findByRole("button", { name: /^create and accept$/i }),
    );

    expect(await screen.findByRole("alert")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Namespace: joaocel" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /^accept device$/i }),
    ).toBeEnabled();
  });

  it("offers a namespace named after the user, with no survey first", async () => {
    useAuthStore.setState({ user: "Joao.Cel_01", username: null });
    renderFlow();

    expect(await screen.findByLabelText(/namespace name/i)).toHaveValue(
      "joao-cel-01",
    );
    expect(screen.queryByRole("radio")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^accept device$/i }),
    ).not.toBeInTheDocument();
  });

  it("creates the namespace and accepts the device into it in one step", async () => {
    let created: unknown;
    let accepted: unknown;
    server.use(
      http.post("*/api/namespaces", async ({ request }) => {
        created = await request.json();
        return HttpResponse.json(
          mockNamespace({ name: "joaocel", tenant_id: "t-new" }),
        );
      }),
      http.post("*/api/devices/pairing/:code/accept", async ({ request }) => {
        accepted = await request.json();
        return HttpResponse.json(acceptedIntoNewNamespace);
      }),
    );
    useAuthStore.setState({ user: "joaocel", username: null });
    const user = userEvent.setup();
    renderFlow();

    await user.click(
      await screen.findByRole("button", { name: /^create and accept$/i }),
    );

    expect(
      await screen.findByRole("heading", { name: /device accepted/i }),
    ).toBeInTheDocument();
    expect(created).toEqual({ name: "joaocel" });
    expect(accepted).toEqual({ tenant_id: "t-new" });
    expect(useAuthStore.getState().tenant).toBe("t-new");
  });

  it("never asks the survey of a user who already had a namespace", async () => {
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([{ name: "my-ns", tenant_id: "t1" }]),
      ),
    );
    const user = userEvent.setup();
    renderFlow();
    await screen.findByText("my-ns");

    await user.click(screen.getByRole("button", { name: /^accept device$/i }));

    expect(
      await screen.findByRole("button", { name: /view device/i }),
    ).toBeInTheDocument();
    expect(screen.queryByRole("radio")).not.toBeInTheDocument();
  });
});
