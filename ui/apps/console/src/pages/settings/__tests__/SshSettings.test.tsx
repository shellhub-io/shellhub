import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { getConfig, defaultConfig } from "@/env";

import SshSettings from "../SshSettings";

const mockedGetConfig = vi.mocked(getConfig);

type Settings = Partial<{
  legacy: boolean;
  require_verified_device_key: boolean;
}>;

function defaultNs(settings: Settings = {}) {
  return mockNamespace({
    settings: {
      session_record: false,
      connection_announcement: "",
      legacy: false,
      require_verified_device_key: false,
      ...settings,
    },
  });
}

function setNamespace(settings: Settings = {}) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(defaultNs(settings)),
    ),
  );
}

async function openUpgrade(user: ReturnType<typeof userEvent.setup>) {
  await user.click(await screen.findByRole("button", { name: "Upgrade" }));

  return screen.getByRole("dialog", { name: "Upgrade this namespace?" });
}

async function confirmUpgrade(user: ReturnType<typeof userEvent.setup>) {
  const dialog = await openUpgrade(user);
  await user.type(
    within(dialog).getByLabelText(/type "my-namespace" to confirm/i),
    "my-namespace",
  );
  await user.click(
    within(dialog).getByRole("button", { name: "Upgrade namespace" }),
  );
}

async function submitBanner(
  user: ReturnType<typeof userEvent.setup>,
  text: string,
) {
  await user.click(await screen.findByRole("button", { name: "Edit" }));
  await user.type(screen.getByRole("textbox", { name: "Login banner" }), text);
  await user.click(screen.getByRole("button", { name: "Save banner" }));
}

function renderSsh() {
  return render(
    <MemoryRouter>
      <SshSettings />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockedGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces/:tenant", () => HttpResponse.json(defaultNs())),
    http.get("*/api/access-policies", () => jsonWithTotal([])),
    http.get("*/api/devices", () => jsonWithTotal([])),
    http.put(
      "*/api/namespaces/:tenant",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.post(
      "*/api/namespaces/:tenant/upgrade",
      () => new HttpResponse(null, { status: 200 }),
    ),
  );
});

describe("SshSettings", () => {
  describe("legacy or new", () => {
    it("offers the upgrade to a legacy namespace", async () => {
      setNamespace({ legacy: true });
      renderSsh();
      expect(await screen.findByText("Legacy")).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Upgrade" }),
      ).toBeInTheDocument();
    });

    it("shows a new namespace with nothing to upgrade", async () => {
      setNamespace({ legacy: false });
      renderSsh();
      expect(await screen.findByText("New")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /upgrade/i }),
      ).not.toBeInTheDocument();
    });

    it("spells out what the upgrade changes and that there is no way back", async () => {
      const user = userEvent.setup();
      setNamespace({ legacy: true });
      renderSsh();

      const dialog = await openUpgrade(user);

      expect(dialog).toHaveTextContent(/the owner keeps ssh access/i);
      expect(dialog).toHaveTextContent(
        /password and public-key logins stop authenticating/i,
      );
      expect(dialog).toHaveTextContent(
        /new devices enroll by pairing or a provisioning key/i,
      );
      expect(dialog).toHaveTextContent(/no way back/i);
    });

    it("upgrades only once the namespace's name is typed", async () => {
      const user = userEvent.setup();
      setNamespace({ legacy: true });
      renderSsh();

      const dialog = await openUpgrade(user);
      const confirm = within(dialog).getByRole("button", {
        name: "Upgrade namespace",
      });
      expect(confirm).toBeDisabled();

      await user.type(
        within(dialog).getByLabelText(/type "my-namespace" to confirm/i),
        "my-namespace",
      );
      expect(confirm).toBeEnabled();
    });

    it("offers no upgrade to a role that may not change the namespace", async () => {
      seedAuthStore({ role: "operator" });
      setNamespace({ legacy: true });
      renderSsh();

      expect(await screen.findByText("Legacy")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /upgrade/i }),
      ).not.toBeInTheDocument();
    });

    it("keeps the namespace legacy when the upgrade is cancelled", async () => {
      const user = userEvent.setup();
      setNamespace({ legacy: true });
      renderSsh();

      await openUpgrade(user);
      await user.click(screen.getByRole("button", { name: "Cancel" }));

      expect(screen.getByText("Legacy")).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Upgrade" }),
      ).toBeInTheDocument();
    });

    it("posts the upgrade and says it saved once confirmed", async () => {
      const user = userEvent.setup();
      let upgraded = 0;
      server.use(
        http.post("*/api/namespaces/:tenant/upgrade", () => {
          upgraded += 1;
          return new HttpResponse(null, { status: 200 });
        }),
      );
      setNamespace({ legacy: true });
      renderSsh();

      await confirmUpgrade(user);

      expect(await screen.findByRole("status")).toHaveTextContent("Saved");
      expect(upgraded).toBe(1);
    });

    it("shows an error when the upgrade fails", async () => {
      server.use(
        http.post("*/api/namespaces/:tenant/upgrade", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      setNamespace({ legacy: true });
      renderSsh();

      await confirmUpgrade(user);

      expect(await screen.findByRole("alert")).toHaveTextContent(
        /couldn't upgrade the namespace/i,
      );
    });
  });

  describe("session recording", () => {
    beforeEach(() => {
      mockedGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
    });

    it("locks the REC switch for a role that may not change it", async () => {
      seedAuthStore({ role: "operator" });
      renderSsh();

      expect(
        await screen.findByRole("switch", { name: "Session recording" }),
      ).toBeDisabled();
    });

    it("says so when recording cannot be changed", async () => {
      server.use(
        http.put("*/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      renderSsh();

      await user.click(
        await screen.findByRole("switch", { name: "Session recording" }),
      );

      expect(await screen.findByRole("alert")).toHaveTextContent(
        /couldn't change session recording/i,
      );
    });

    it("turns recording on from the REC switch", async () => {
      const bodies: unknown[] = [];
      server.use(
        http.put("*/api/namespaces/:tenant", async ({ request }) => {
          bodies.push(await request.json());
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const user = userEvent.setup();
      renderSsh();

      const rec = await screen.findByRole("switch", {
        name: "Session recording",
      });
      expect(rec).not.toBeChecked();
      await user.click(rec);

      expect(await screen.findByRole("status")).toHaveTextContent("Saved");
      expect(bodies).toEqual([{ settings: { session_record: true } }]);
    });

    it("is not offered in community", async () => {
      mockedGetConfig.mockReturnValue({ ...defaultConfig });
      renderSsh();

      await screen.findByText("New");
      expect(
        screen.queryByRole("switch", { name: "Session recording" }),
      ).not.toBeInTheDocument();
    });
  });

  describe("device signatures", () => {
    function recordPuts() {
      const bodies: unknown[] = [];
      server.use(
        http.put("*/api/namespaces/:tenant", async ({ request }) => {
          bodies.push(await request.json());
          return new HttpResponse(null, { status: 204 });
        }),
      );
      return bodies;
    }

    function unsignedDevices(accepted: number, pending = 0) {
      const queries: URLSearchParams[] = [];
      server.use(
        http.get("*/api/devices", ({ request }) => {
          const query = new URL(request.url).searchParams;
          queries.push(query);
          return jsonWithTotal(
            [],
            query.get("status") === "pending" ? pending : accepted,
          );
        }),
      );
      return queries;
    }

    function knownDevices(accepted: number, pending = 0) {
      server.use(
        http.get("*/api/namespaces/:tenant", () =>
          HttpResponse.json({
            ...defaultNs(),
            devices_accepted_count: accepted,
            devices_pending_count: pending,
          }),
        ),
      );
    }

    const findSwitch = () =>
      screen.findByRole("switch", { name: "Device signature enforcement" });

    async function turnOn(user: ReturnType<typeof userEvent.setup>) {
      const toggle = await findSwitch();
      await waitFor(() => expect(toggle).toBeEnabled());
      await user.click(toggle);
    }

    it("shows whether the namespace enforces device signatures", async () => {
      setNamespace({ require_verified_device_key: true });
      renderSsh();

      expect(await findSwitch()).toBeChecked();
    });

    it("locks the switch for a role below administrator", async () => {
      seedAuthStore({ role: "operator" });
      renderSsh();

      expect(await findSwitch()).toBeDisabled();
    });

    it("counts the accepted and pending devices that never signed", async () => {
      const queries = unsignedDevices(0);
      renderSsh();

      await findSwitch();
      await waitFor(() => expect(queries).toHaveLength(2));
      expect(queries.map((query) => query.get("status")).sort()).toEqual([
        "accepted",
        "pending",
      ]);
      for (const query of queries) {
        expect(JSON.parse(atob(query.get("filter") ?? ""))).toEqual([
          {
            type: "property",
            params: { name: "key_verified", operator: "bool", value: false },
          },
        ]);
      }
    });

    it("counts a pending device that can't sign in the warning", async () => {
      unsignedDevices(2, 1);
      knownDevices(40, 2);
      renderSsh();

      expect(
        await screen.findByText(
          "3 of your 42 devices don't support signatures yet. Update them before you enforce them here.",
        ),
      ).toBeInTheDocument();
    });

    it("turns on at once when no accepted or pending device would stop connecting", async () => {
      unsignedDevices(0);
      const bodies = recordPuts();
      const user = userEvent.setup();
      renderSsh();

      await turnOn(user);

      await waitFor(() =>
        expect(bodies).toEqual([
          { settings: { require_verified_device_key: true } },
        ]),
      );
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("warns about the devices that can't sign while it is off", async () => {
      unsignedDevices(3);
      knownDevices(42);
      renderSsh();

      const card = await screen.findByRole("group", {
        name: "Device signature enforcement",
      });
      expect(
        await within(card).findByText(
          "3 of your 42 devices don't support signatures yet. Update them before you enforce them here.",
        ),
      ).toBeInTheDocument();
    });

    it("warns about nothing once it is on", async () => {
      unsignedDevices(3);
      setNamespace({ require_verified_device_key: true });
      renderSsh();

      await findSwitch();
      expect(
        screen.queryByText(/don't support signatures/),
      ).not.toBeInTheDocument();
    });

    it("says how many devices stop connecting before turning on, and saves once confirmed", async () => {
      unsignedDevices(3);
      knownDevices(42);
      const bodies = recordPuts();
      const user = userEvent.setup();
      renderSsh();

      await screen.findByText(/don't support signatures/);
      await turnOn(user);

      const dialog = await screen.findByRole("dialog", {
        name: "Enforce device signatures?",
      });
      expect(dialog).toHaveAccessibleDescription(
        "3 of 42 devices will stop connecting.",
      );
      expect(dialog).toHaveTextContent(
        "ShellHub will refuse their agents until they're updated",
      );
      expect(bodies).toEqual([]);

      await user.click(
        within(dialog).getByRole("button", { name: "Enforce anyway" }),
      );

      expect(await screen.findByRole("status")).toHaveTextContent("Saved");
      expect(bodies).toEqual([
        { settings: { require_verified_device_key: true } },
      ]);
    });

    it("still asks when the devices cannot be counted", async () => {
      server.use(
        http.get("*/api/devices", () => HttpResponse.json({}, { status: 500 })),
      );
      const user = userEvent.setup();
      renderSsh();

      await turnOn(user);

      const dialog = await screen.findByRole("dialog", {
        name: "Enforce device signatures?",
      });
      expect(dialog).toHaveAccessibleDescription(
        "Devices whose agent can't sign will stop connecting.",
      );
    });

    it("saves nothing when turning it on is cancelled", async () => {
      unsignedDevices(3);
      const bodies = recordPuts();
      const user = userEvent.setup();
      renderSsh();

      await turnOn(user);
      await user.click(await screen.findByRole("button", { name: "Cancel" }));

      expect(await findSwitch()).not.toBeChecked();
      expect(bodies).toEqual([]);
    });

    it("turns it off at once, without a confirmation", async () => {
      const bodies = recordPuts();
      setNamespace({ require_verified_device_key: true });
      const user = userEvent.setup();
      renderSsh();

      await user.click(await findSwitch());

      expect(await screen.findByRole("status")).toHaveTextContent("Saved");
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      expect(bodies).toEqual([
        { settings: { require_verified_device_key: false } },
      ]);
    });
  });

  describe("login banner", () => {
    it("edits the banner in a dialog, not on the page", async () => {
      const user = userEvent.setup();
      renderSsh();

      expect(
        await screen.findByRole("group", { name: "Login banner" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("textbox", { name: "Login banner" }),
      ).not.toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: "Edit" }));

      expect(
        within(screen.getByRole("dialog", { name: "Login banner" })).getByRole(
          "textbox",
          { name: "Login banner" },
        ),
      ).toBeInTheDocument();
    });

    it("keeps Save disabled until the banner changes", async () => {
      const user = userEvent.setup();
      renderSsh();

      await user.click(await screen.findByRole("button", { name: "Edit" }));

      expect(
        screen.getByRole("button", { name: "Save banner" }),
      ).toBeDisabled();
    });

    it("saves the edited banner and closes", async () => {
      const bodies: unknown[] = [];
      server.use(
        http.put("*/api/namespaces/:tenant", async ({ request }) => {
          bodies.push(await request.json());
          return new HttpResponse(null, { status: 204 });
        }),
      );
      const user = userEvent.setup();
      renderSsh();

      await submitBanner(user, "Authorized use only");

      expect(bodies).toEqual([
        { settings: { connection_announcement: "Authorized use only" } },
      ]);
      expect(
        screen.queryByRole("dialog", { name: "Login banner" }),
      ).not.toBeInTheDocument();
    });

    it("lets a role that may not edit it read the banner", async () => {
      seedAuthStore({ role: "observer" });
      server.use(
        http.get("*/api/namespaces/:tenant", () =>
          HttpResponse.json(
            mockNamespace({
              settings: {
                session_record: false,
                connection_announcement: "Authorized use only",
                legacy: false,
                require_verified_device_key: false,
              },
            }),
          ),
        ),
      );
      const user = userEvent.setup();
      renderSsh();

      const view = await screen.findByRole("button", { name: "View" });
      expect(
        screen.queryByRole("button", { name: "Edit" }),
      ).not.toBeInTheDocument();
      await user.click(view);

      const viewer = screen.getByRole("dialog", { name: "Login banner" });
      expect(
        within(viewer).getByText(/Authorized use only/),
      ).toBeInTheDocument();
      expect(within(viewer).queryByRole("textbox")).not.toBeInTheDocument();
    });

    it("says why a banner past the limit cannot be saved", async () => {
      const user = userEvent.setup();
      renderSsh();

      await user.click(await screen.findByRole("button", { name: "Edit" }));
      const textarea = screen.getByRole("textbox", { name: "Login banner" });
      await user.click(textarea);
      await user.paste("x".repeat(4097));

      expect(await screen.findByRole("alert")).toHaveTextContent(
        "A banner can be at most 4,096 characters.",
      );
      expect(
        screen.getByRole("button", { name: "Save banner" }),
      ).toBeDisabled();
    });

    it("keeps the dialog open and says why when saving fails", async () => {
      server.use(
        http.put("*/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      renderSsh();

      await submitBanner(user, "Authorized use only");

      expect(await screen.findByRole("alert")).toHaveTextContent(
        "Couldn't save the banner. Try again.",
      );
      expect(
        screen.getByRole("dialog", { name: "Login banner" }),
      ).toBeInTheDocument();
    });
  });
});
