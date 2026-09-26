import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { getConfig, defaultConfig } from "@/env";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

import GeneralSettings from "../GeneralSettings";

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

function renderGeneral() {
  return render(
    <MemoryRouter>
      <GeneralSettings />
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
    http.delete(
      "*/api/namespaces/:tenant",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.delete("*/api/namespaces/:tenant/members", () =>
      HttpResponse.json({}),
    ),
    http.put(
      "*/api/namespaces/ssh-access-mode/:tenant",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token" }),
    ),
    http.get("*/api/stats", () =>
      HttpResponse.json({
        registered_devices: 3,
        online_devices: 2,
        active_sessions: 0,
        pending_devices: 0,
        rejected_devices: 0,
      }),
    ),
  );
});

describe("GeneralSettings", () => {
  it("shows the namespace name and tenant ID", async () => {
    renderGeneral();
    expect(await screen.findByText("my-namespace")).toBeInTheDocument();
    expect(
      within(screen.getByRole("group", { name: "Tenant ID" })).getByText(
        "tenant-456",
      ),
    ).toBeInTheDocument();
  });

  it("shows how many devices are online", async () => {
    renderGeneral();
    const online = await screen.findByRole("group", { name: "Online" });
    expect(await within(online).findByText("2")).toBeInTheDocument();
  });

  describe("as the owner", () => {
    it("names the role without repeating who owns it", async () => {
      renderGeneral();
      expect(await screen.findByText("owner")).toBeInTheDocument();
      expect(screen.queryByText(/owned by/i)).not.toBeInTheDocument();
    });

    it("offers to delete the namespace, not to leave it", async () => {
      renderGeneral();
      expect(
        await screen.findByRole("button", { name: "Delete namespace" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Leave" }),
      ).not.toBeInTheDocument();
    });

    it("asks for the name before deleting", async () => {
      const user = userEvent.setup();
      renderGeneral();

      await user.click(
        await screen.findByRole("button", { name: "Delete namespace" }),
      );
      const dialog = screen.getByRole("dialog", { name: "Delete namespace" });
      const confirm = within(dialog).getByRole("button", {
        name: "Delete namespace",
      });
      expect(confirm).toBeDisabled();

      await user.type(
        within(dialog).getByLabelText(/type "my-namespace" to confirm/i),
        "my-namespace",
      );
      expect(confirm).toBeEnabled();
    });
  });

  describe("as a member who does not own it", () => {
    beforeEach(() => {
      seedAuthStore({ userId: "user-999", role: "operator" });
    });

    it("names the role and who owns the namespace", async () => {
      renderGeneral();
      expect(await screen.findByText("operator")).toBeInTheDocument();
      expect(screen.getByText(/owned by admin@test\.com/)).toBeInTheDocument();
    });

    it("offers to leave instead of delete", async () => {
      renderGeneral();
      expect(
        await screen.findByRole("button", { name: "Leave" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Delete namespace" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Rename" }),
      ).not.toBeInTheDocument();
    });

    it("confirms before leaving", async () => {
      const user = userEvent.setup();
      renderGeneral();

      await user.click(await screen.findByRole("button", { name: "Leave" }));

      expect(
        screen.getByRole("dialog", { name: "Leave namespace" }),
      ).toBeInTheDocument();
    });
  });

  describe("rename", () => {
    async function openRenameModal() {
      const user = userEvent.setup();
      renderGeneral();
      await user.click(await screen.findByRole("button", { name: "Rename" }));
      return user;
    }

    it("keeps Save disabled while the name is unchanged", async () => {
      await openRenameModal();
      expect(screen.getByRole("button", { name: /save/i })).toBeDisabled();
    });

    it("keeps Save disabled while the new name is invalid", async () => {
      const user = await openRenameModal();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "ab");
      expect(screen.getByRole("button", { name: /save/i })).toBeDisabled();
    });

    it("enables Save once the name is changed and valid", async () => {
      const user = await openRenameModal();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "new-valid-name");
      expect(screen.getByRole("button", { name: /save/i })).toBeEnabled();
    });

    it("closes after a successful rename", async () => {
      const user = await openRenameModal();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "new-valid-name");
      await user.click(screen.getByRole("button", { name: /save/i }));
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("starts from the current name when reopened", async () => {
      const user = await openRenameModal();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "changed-name");
      await user.click(screen.getByRole("button", { name: /cancel/i }));
      await user.click(screen.getByRole("button", { name: "Discard changes" }));
      await user.click(screen.getByRole("button", { name: "Rename" }));
      expect(screen.getByLabelText(/namespace name/i)).toHaveValue(
        "my-namespace",
      );
    });

    it("shows an error when the rename fails", async () => {
      server.use(
        http.put("*/api/namespaces/:tenant", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const user = await openRenameModal();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "new-valid-name");
      await user.click(screen.getByRole("button", { name: /save/i }));
      expect(await screen.findByRole("alert")).toBeInTheDocument();
    });
  });
});
