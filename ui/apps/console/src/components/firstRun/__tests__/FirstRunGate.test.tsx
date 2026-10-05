import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockDevice, mockNamespace, mockStats } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import FirstRunGate from "../FirstRunGate";

vi.mock("@/components/layout/SessionMenu", () => ({
  default: () => <div data-testid="session-menu" />,
}));

vi.mock("@/components/ConnectModal", () => ({
  default: ({ open, sshid }: { open: boolean; sshid: string }) =>
    open ? <div role="dialog">connect to {sshid}</div> : null,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

const TENANT = "tenant-1";

function serveDashboard({
  accepted = 0,
  pending = 0,
  namespaces = [mockNamespace({ tenant_id: TENANT, name: "dev" })],
}: {
  accepted?: number;
  pending?: number;
  namespaces?: ReturnType<typeof mockNamespace>[];
} = {}) {
  server.use(
    http.get("*/api/stats", () =>
      HttpResponse.json(
        mockStats({ registered_devices: accepted, pending_devices: pending }),
      ),
    ),
    http.get("*/api/namespaces", () => jsonWithTotal(namespaces)),
    http.get(`*/api/namespaces/${TENANT}`, () =>
      HttpResponse.json(mockNamespace({ tenant_id: TENANT, name: "dev" })),
    ),
    http.get("*/api/devices", () => jsonWithTotal([])),
    http.get("*/api/devices/:uid", ({ params }) =>
      HttpResponse.json(
        mockDevice({ uid: String(params.uid), name: "box", online: true }),
      ),
    ),
    http.get("*/info", () =>
      HttpResponse.json({ version: "test", endpoints: null, setup: true }),
    ),
  );
}

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route element={<FirstRunGate />}>
          <Route path="/dashboard" element={<div>normal dashboard</div>} />
          <Route path="/devices" element={<div>device list</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore({ tenant: TENANT, role: "owner" });
});

afterEach(() => {
  vi.useRealTimers();
});

describe("FirstRunGate", () => {
  it("shows the first-run trail on the dashboard of a namespace with no accepted device", async () => {
    serveDashboard();
    renderAt("/dashboard");

    expect(
      await screen.findByRole("heading", { name: /get your first shell/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/install the agent/i)).toBeInTheDocument();
    expect(screen.queryByText("normal dashboard")).not.toBeInTheDocument();
  });

  it("counts only accepted devices, so pending ones still get the trail", async () => {
    serveDashboard({ pending: 2 });
    renderAt("/dashboard");

    expect(
      await screen.findByRole("heading", { name: /get your first shell/i }),
    ).toBeInTheDocument();
  });

  it("shows the normal dashboard once the namespace has an accepted device", async () => {
    serveDashboard({ accepted: 1 });
    renderAt("/dashboard");

    expect(await screen.findByText("normal dashboard")).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /get your first shell/i }),
    ).not.toBeInTheDocument();
  });

  it("shows the normal dashboard to a member who cannot accept devices", async () => {
    seedAuthStore({ tenant: TENANT, role: "observer" });
    serveDashboard();
    renderAt("/dashboard");

    expect(await screen.findByText("normal dashboard")).toBeInTheDocument();
  });

  it("leaves every other route alone", async () => {
    serveDashboard();
    renderAt("/devices");

    expect(await screen.findByText("device list")).toBeInTheDocument();
  });

  it("names the namespace the device will join as a finished step", async () => {
    serveDashboard();
    renderAt("/dashboard");

    expect(
      await screen.findByRole("heading", { name: /^namespace$/i }),
    ).toBeInTheDocument();
    expect(await screen.findByText("dev")).toBeInTheDocument();
  });

  it("leaves the dashboard to the app when the user has several namespaces", async () => {
    serveDashboard({
      namespaces: [
        mockNamespace({ tenant_id: TENANT, name: "dev" }),
        mockNamespace({ tenant_id: "tenant-2", name: "prod" }),
      ],
    });
    renderAt("/dashboard");

    expect(await screen.findByText("normal dashboard")).toBeInTheDocument();
  });
});

const PREVIEW = {
  kind: "pairing" as const,
  ...mockDevice({
    name: "raspberrypi",
    identity: { mac: "dc:a6:32:4e:1f:08" },
    info: {
      id: "debian",
      pretty_name: "Debian GNU/Linux 12 (bookworm)",
      version: "v0.27.0",
      arch: "arm64",
      platform: "native",
    },
  }),
};

function serveCode(code: string) {
  server.use(
    http.get(`*/api/devices/login-code/${code}`, () =>
      HttpResponse.json(PREVIEW),
    ),
  );
}

function codeCells() {
  return screen
    .getAllByLabelText(/character \d of 8/i)
    .map((cell) => (cell as HTMLInputElement).value)
    .join("");
}

async function activeStepFind(text: string | RegExp) {
  await screen.findByRole("button", { name: /open terminal/i });
  return within(screen.getByRole("listitem", { current: "step" })).findByText(
    text,
  );
}

function activeStep() {
  return within(screen.getByRole("listitem", { current: "step" }));
}

async function enterCode(
  user: ReturnType<typeof userEvent.setup>,
  value: string,
) {
  await user.click(await screen.findByLabelText(/character 1 of 8/i));
  await user.paste(value);
  await user.click(screen.getByRole("button", { name: /^continue$/i }));
}

describe("FirstRunGate pairing", () => {
  it("takes a typed code one character per cell, in upper case", async () => {
    serveDashboard();
    const user = userEvent.setup();
    renderAt("/dashboard");

    await user.type(
      await screen.findByLabelText(/character 1 of 8/i),
      "7k2m9qxr",
    );

    expect(codeCells()).toBe("7K2M9QXR");
  });

  it("takes the code out of a pasted link", async () => {
    serveDashboard();
    const user = userEvent.setup();
    renderAt("/dashboard");

    await user.click(await screen.findByLabelText(/character 1 of 8/i));
    await user.paste("http://shellhub.local/accept-device?code=7K2M9QXR");

    expect(codeCells()).toBe("7K2M9QXR");
  });

  it("previews the device behind the code before pairing", async () => {
    serveDashboard();
    serveCode("7K2M9QXR");
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "7K2M-9QXR");

    await screen.findByRole("button", { name: /pair into dev/i });
    const step = activeStep();
    expect(step.getByText("raspberrypi")).toBeInTheDocument();
    expect(
      step.getByText(/debian gnu\/linux 12 \(bookworm\)/i),
    ).toBeInTheDocument();
    expect(step.getByText("dc:a6:32:4e:1f:08")).toBeInTheDocument();
  });

  it("pairs the device into the current namespace and moves to the shell step", async () => {
    serveDashboard();
    serveCode("7K2M9QXR");
    let body: unknown;
    server.use(
      http.post(
        "*/api/devices/pairing/7K2M9QXR/accept",
        async ({ request }) => {
          body = await request.json();
          return HttpResponse.json({
            uid: "paired-uid",
            tenant_id: TENANT,
            namespace: "dev",
          });
        },
      ),
    );
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "7K2M9QXR");
    await user.click(
      await screen.findByRole("button", { name: /pair into dev/i }),
    );

    expect(
      await screen.findByRole("button", { name: /open terminal/i }),
    ).toBeInTheDocument();
    expect(body).toEqual({ tenant_id: TENANT });
  });

  it("says an unknown code is invalid or expired and how to get a new one", async () => {
    serveDashboard();
    server.use(
      http.get("*/api/devices/login-code/:code", () =>
        HttpResponse.json({}, { status: 404 }),
      ),
    );
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "AAAABBBB");

    expect(
      await screen.findByText(/invalid or has expired/i),
    ).toBeInTheDocument();
    expect(screen.getByText(/shellhub-agent login/i)).toBeInTheDocument();
  });

  it("shows why an accept was refused", async () => {
    serveDashboard();
    serveCode("7K2M9QXR");
    server.use(
      http.post("*/api/devices/pairing/7K2M9QXR/accept", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "7K2M9QXR");
    await user.click(
      await screen.findByRole("button", { name: /pair into dev/i }),
    );

    expect(
      await screen.findByText(/do not have permission to accept devices/i),
    ).toBeInTheDocument();
  });

  it("tells a lookup failure apart from a bad code", async () => {
    serveDashboard();
    server.use(
      http.get("*/api/devices/login-code/:code", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "AAAABBBB");

    expect(
      await screen.findByText(/couldn't look that code up/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(/invalid or has expired/i),
    ).not.toBeInTheDocument();
  });

  it("looks the code up again when Continue is pressed after a failure", async () => {
    serveDashboard();
    let lookups = 0;
    server.use(
      http.get("*/api/devices/login-code/7K2M9QXR", () => {
        lookups += 1;
        return lookups === 1
          ? HttpResponse.json({}, { status: 500 })
          : HttpResponse.json(PREVIEW);
      }),
    );
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "7K2M9QXR");
    await screen.findByText(/couldn't look that code up/i);
    await user.click(screen.getByRole("button", { name: /^continue$/i }));

    expect(
      await screen.findByRole("button", { name: /pair into dev/i }),
    ).toBeInTheDocument();
  });

  it("ignores a paste that is not a code", async () => {
    serveDashboard();
    const user = userEvent.setup();
    renderAt("/dashboard");

    await user.click(await screen.findByLabelText(/character 1 of 8/i));
    await user.paste("ABCD12345");

    expect(codeCells()).toBe("");
  });

  it("explains a refusal for being over the device limit", async () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    serveDashboard();
    serveCode("7K2M9QXR");
    server.use(
      http.post("*/api/devices/pairing/7K2M9QXR/accept", () =>
        HttpResponse.json({}, { status: 402 }),
      ),
    );
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "7K2M9QXR");
    await user.click(
      await screen.findByRole("button", { name: /pair into dev/i }),
    );

    expect(
      await screen.findByText(/reached its device limit/i),
    ).toBeInTheDocument();
  });

  it("goes back to the code field when the preview is the wrong device", async () => {
    serveDashboard();
    serveCode("7K2M9QXR");
    const user = userEvent.setup();
    renderAt("/dashboard");

    await enterCode(user, "7K2M9QXR");
    await user.click(
      await screen.findByRole("button", { name: /not this one/i }),
    );

    await screen.findByLabelText(/character 1 of 8/i);
    expect(codeCells()).toBe("");
  });

  it("moves on when the device is accepted from the printed link", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    serveDashboard();
    let calls = 0;
    server.use(
      http.get("*/api/devices", () => {
        calls += 1;
        return jsonWithTotal(
          calls > 1
            ? [mockDevice({ uid: "linked-uid", name: "linked-box" })]
            : [],
        );
      }),
      http.get("*/api/devices/linked-uid", () =>
        HttpResponse.json(
          mockDevice({ uid: "linked-uid", name: "linked-box", online: true }),
        ),
      ),
    );
    renderAt("/dashboard");

    await screen.findByLabelText(/character 1 of 8/i);
    await vi.advanceTimersByTimeAsync(3000);

    await screen.findByRole("button", { name: /open terminal/i });
    expect(activeStep().getByText("linked-box")).toBeInTheDocument();
  });
});

function serveLinkedDevice(sshEndpoint: string) {
  server.use(
    http.get("*/api/devices", () =>
      jsonWithTotal([mockDevice({ uid: "box-uid", name: "box" })]),
    ),
    http.get("*/info", () =>
      HttpResponse.json({
        version: "test",
        endpoints: { api: "shellhub.local", ssh: sshEndpoint },
        setup: true,
      }),
    ),
  );
}

describe("FirstRunGate shell step", () => {
  it("shows the paired device online once its agent connects", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    serveDashboard();
    serveLinkedDevice("shellhub.local:22");
    let checks = 0;
    server.use(
      http.get("*/api/devices/:uid", ({ params }) => {
        checks += 1;
        return HttpResponse.json(
          mockDevice({
            uid: String(params.uid),
            name: "box",
            online: checks > 1,
          }),
        );
      }),
    );
    renderAt("/dashboard");

    expect(await activeStepFind("connecting...")).toBeInTheDocument();
    await vi.advanceTimersByTimeAsync(3000);

    expect(await activeStepFind("online")).toBeInTheDocument();
  });

  it("gives the ssh command for the paired device, with the SSH port", async () => {
    serveDashboard();
    serveLinkedDevice("shellhub.local:2202");
    renderAt("/dashboard");

    expect(
      await screen.findByRole("button", {
        name: "Copy ssh -p 2202 root@dev.box@shellhub.local",
      }),
    ).toBeInTheDocument();
  });

  it("addresses the device by its stored name, not the one the code previewed", async () => {
    serveDashboard();
    serveLinkedDevice("shellhub.local:22");
    server.use(
      http.get("*/api/devices/:uid", ({ params }) =>
        HttpResponse.json(
          mockDevice({
            uid: String(params.uid),
            name: "renamed",
            online: true,
          }),
        ),
      ),
    );
    renderAt("/dashboard");

    expect(
      await screen.findByRole("button", {
        name: "Copy ssh root@dev.renamed@shellhub.local",
      }),
    ).toBeInTheDocument();
  });

  it("leaves the port out when it is the default one", async () => {
    serveDashboard();
    serveLinkedDevice("shellhub.local:22");
    renderAt("/dashboard");

    expect(
      await screen.findByRole("button", {
        name: "Copy ssh root@dev.box@shellhub.local",
      }),
    ).toBeInTheDocument();
  });

  it("lets the user change the login in the command", async () => {
    serveDashboard();
    serveLinkedDevice("shellhub.local:22");
    const user = userEvent.setup();
    renderAt("/dashboard");

    const login = await screen.findByLabelText(/login on the device/i);
    await user.clear(login);
    await user.type(login, "pi");

    expect(
      screen.getByRole("button", {
        name: "Copy ssh pi@dev.box@shellhub.local",
      }),
    ).toBeInTheDocument();
  });

  it.each([
    {
      mode: "identity" as const,
      legacyAllowed: false,
      hint: /uses your ssh key\. the first time, it shows a link to approve it\./i,
    },
    {
      mode: "legacy" as const,
      legacyAllowed: true,
      hint: /the device's own users and passwords apply\./i,
    },
  ])(
    "tells how the ssh command signs in when the namespace is in $mode mode",
    async ({ mode, legacyAllowed, hint }) => {
      serveDashboard();
      serveLinkedDevice("shellhub.local:22");
      server.use(
        http.get(`*/api/namespaces/${TENANT}`, () =>
          HttpResponse.json(
            mockNamespace({
              tenant_id: TENANT,
              name: "dev",
              settings: {
                session_record: false,
                connection_announcement: "",
                ssh_access_mode: mode,
                ssh_legacy_allowed: legacyAllowed,
              },
            }),
          ),
        ),
      );
      renderAt("/dashboard");

      expect(await activeStepFind(hint)).toBeInTheDocument();
    },
  );

  it("opens the terminal dialog on the paired device", async () => {
    serveDashboard();
    serveLinkedDevice("shellhub.local:22");
    const user = userEvent.setup();
    renderAt("/dashboard");

    await user.click(
      await screen.findByRole("button", { name: /open terminal/i }),
    );

    expect(screen.getByRole("dialog")).toHaveTextContent(
      /connect to dev\.box@/,
    );
  });
});

describe("FirstRunGate when the counts fail", () => {
  it("leaves the dashboard to the app when the device counts cannot be loaded", async () => {
    serveDashboard();
    server.use(
      http.get("*/api/stats", () => HttpResponse.json({}, { status: 500 })),
    );
    renderAt("/dashboard");

    expect(await screen.findByText("normal dashboard")).toBeInTheDocument();
  });
});
