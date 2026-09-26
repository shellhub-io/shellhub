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

function defaultNs(
  settings: Partial<{
    ssh_access_mode: "legacy" | "identity";
    ssh_legacy_allowed: boolean;
  }> = {},
) {
  return mockNamespace({
    settings: {
      session_record: false,
      connection_announcement: "",
      ssh_access_mode: "legacy",
      ssh_legacy_allowed: true,
      ...settings,
    },
  });
}

function setNamespace(
  settings: Partial<{
    ssh_access_mode: "legacy" | "identity";
    ssh_legacy_allowed: boolean;
  }> = {},
) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(defaultNs(settings)),
    ),
  );
}

async function confirmSwitch(user: ReturnType<typeof userEvent.setup>) {
  await user.click(
    await screen.findByRole("button", { name: "Switch to identity" }),
  );
  await user.click(
    within(
      screen.getByRole("dialog", { name: "Switch to identity access?" }),
    ).getByRole("button", { name: "Switch to identity" }),
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
    http.put(
      "*/api/namespaces/:tenant",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.put(
      "*/api/namespaces/ssh-access-mode/:tenant",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("SshSettings", () => {
  describe("access mode", () => {
    it("offers the other mode to a namespace that may switch", async () => {
      setNamespace({ ssh_access_mode: "legacy", ssh_legacy_allowed: true });
      renderSsh();
      expect(await screen.findByText("Legacy")).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Switch to identity" }),
      ).toBeInTheDocument();
    });

    it("shows identity without a choice to a namespace born in it", async () => {
      setNamespace({ ssh_access_mode: "identity", ssh_legacy_allowed: false });
      renderSsh();
      expect(await screen.findByText("Identity")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /switch to/i }),
      ).not.toBeInTheDocument();
    });

    it("asks before switching, and warns that no policies deny every login", async () => {
      const user = userEvent.setup();
      setNamespace({ ssh_access_mode: "legacy", ssh_legacy_allowed: true });
      renderSsh();

      await user.click(
        await screen.findByRole("button", { name: "Switch to identity" }),
      );

      await waitFor(() =>
        expect(
          screen.getByRole("dialog", { name: "Switch to identity access?" }),
        ).toHaveAccessibleDescription(/every SSH login will be denied/i),
      );
    });

    it("offers no switch to a role that may not change the mode", async () => {
      seedAuthStore({ role: "operator" });
      setNamespace({ ssh_access_mode: "legacy", ssh_legacy_allowed: true });
      renderSsh();

      expect(await screen.findByText("Legacy")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /switch to/i }),
      ).not.toBeInTheDocument();
    });

    it("keeps the mode when the switch is cancelled", async () => {
      const user = userEvent.setup();
      setNamespace({ ssh_access_mode: "legacy", ssh_legacy_allowed: true });
      renderSsh();

      await user.click(
        await screen.findByRole("button", { name: "Switch to identity" }),
      );
      await user.click(screen.getByRole("button", { name: "Cancel" }));

      expect(screen.getByText("Legacy")).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Switch to identity" }),
      ).toBeInTheDocument();
    });

    it("says it saved once the switch is confirmed", async () => {
      const user = userEvent.setup();
      setNamespace({ ssh_access_mode: "legacy", ssh_legacy_allowed: true });
      renderSsh();

      await confirmSwitch(user);

      expect(await screen.findByRole("status")).toHaveTextContent("Saved");
    });

    it("shows an error when the switch fails", async () => {
      server.use(
        http.put("*/api/namespaces/ssh-access-mode/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = userEvent.setup();
      setNamespace({ ssh_access_mode: "legacy", ssh_legacy_allowed: true });
      renderSsh();

      await confirmSwitch(user);

      expect(await screen.findByRole("alert")).toHaveTextContent(
        /couldn't change the ssh access mode/i,
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

      await screen.findByText("Legacy");
      expect(
        screen.queryByRole("switch", { name: "Session recording" }),
      ).not.toBeInTheDocument();
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

      expect(screen.getByRole("button", { name: "Save banner" })).toBeDisabled();
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
                ssh_access_mode: "legacy",
                ssh_legacy_allowed: true,
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
      expect(within(viewer).getByText(/Authorized use only/)).toBeInTheDocument();
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
      expect(screen.getByRole("button", { name: "Save banner" })).toBeDisabled();
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
