import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespace: vi.fn(),
    listAccessPolicies: vi.fn(),
    editNamespace: vi.fn(),
    deleteNamespace: vi.fn(),
    leaveNamespace: vi.fn(),
    setSshAccessMode: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

vi.mock("@/components/billing/BillingSection", () => ({
  default: () => null,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

import Settings from "../Settings";
import * as SettingsCardModule from "@/components/common/SettingsCard";
import * as SettingsRowModule from "@/components/common/SettingsRow";
import { getConfig, defaultConfig } from "@/env";

const mockedGetConfig = vi.mocked(getConfig);

function defaultNamespace(
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

function renderSettings() {
  return render(
    <MemoryRouter>
      <Settings />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockedGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore();
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(defaultNamespace()));
  sdk.listAccessPolicies.mockResolvedValue(paginatedResponse([]));
  sdk.editNamespace.mockResolvedValue(mockSdkResponse(undefined));
  sdk.deleteNamespace.mockResolvedValue(mockSdkResponse(undefined));
  sdk.leaveNamespace.mockResolvedValue(mockSdkResponse(undefined));
  sdk.setSshAccessMode.mockResolvedValue(mockSdkResponse(undefined));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token" }),
  );
});

describe("Settings", () => {
  describe("shared component usage", () => {
    it("uses the shared SettingsCard component (not a local copy)", async () => {
      const spy = vi.spyOn(SettingsCardModule, "default");
      renderSettings();
      await screen.findByRole("heading", { name: /^general$/i });
      expect(spy).toHaveBeenCalled();
      spy.mockRestore();
    });

    it("uses the shared SettingsRow component (not a local copy)", async () => {
      const spy = vi.spyOn(SettingsRowModule, "default");
      renderSettings();
      await screen.findByRole("heading", { name: /^general$/i });
      expect(spy).toHaveBeenCalled();
      spy.mockRestore();
    });
  });

  describe("renders settings sections", () => {
    it("shows the General card heading", async () => {
      renderSettings();
      expect(
        await screen.findByRole("heading", { name: /^general$/i }),
      ).toBeInTheDocument();
    });

    it("shows the SSH card heading", async () => {
      renderSettings();
      expect(
        await screen.findByRole("heading", { name: /^ssh$/i }),
      ).toBeInTheDocument();
    });

    it("shows the Danger Zone card heading", async () => {
      renderSettings();
      expect(
        await screen.findByRole("heading", { name: /danger zone/i }),
      ).toBeInTheDocument();
    });

    it("renders the namespace name", async () => {
      renderSettings();
      expect(await screen.findByText("my-namespace")).toBeInTheDocument();
    });

    it("renders the tenant ID", async () => {
      renderSettings();
      expect(await screen.findByText("tenant-456")).toBeInTheDocument();
    });
  });

  describe("danger zone", () => {
    it("renders the Delete button for owners", async () => {
      renderSettings();
      expect(
        await screen.findByRole("button", { name: /delete/i }),
      ).toBeInTheDocument();
    });
  });

  describe("SSH access mode", () => {
    it("shows the Legacy/Identity toggle for grandfathered namespaces", async () => {
      sdk.getNamespace.mockResolvedValue(
        mockSdkResponse(defaultNamespace({ ssh_legacy_allowed: true })),
      );
      renderSettings();
      expect(
        await screen.findByRole("button", { name: "Legacy" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Identity" }),
      ).toBeInTheDocument();
    });

    it("shows no toggle for namespaces born in identity mode", async () => {
      sdk.getNamespace.mockResolvedValue(
        mockSdkResponse(
          defaultNamespace({
            ssh_access_mode: "identity",
            ssh_legacy_allowed: false,
          }),
        ),
      );
      renderSettings();
      expect(await screen.findByText("Identity")).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Legacy" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Identity" }),
      ).not.toBeInTheDocument();
    });
  });

  describe("EditNameDrawer", () => {
    async function openRenameDrawer() {
      const user = userEvent.setup();
      renderSettings();
      await user.click(
        await screen.findByRole("button", { name: /rename namespace/i }),
      );
      return user;
    }

    it("Save is disabled when the name is unchanged (not dirty)", async () => {
      await openRenameDrawer();
      expect(screen.getByRole("button", { name: /save/i })).toBeDisabled();
    });

    it("Save is disabled when the new name is invalid", async () => {
      const user = await openRenameDrawer();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "ab");
      expect(screen.getByRole("button", { name: /save/i })).toBeDisabled();
    });

    it("Save is enabled when the name is dirty and valid", async () => {
      const user = await openRenameDrawer();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "new-valid-name");
      expect(screen.getByRole("button", { name: /save/i })).not.toBeDisabled();
    });

    it("calls editNamespace with the new name on submit and closes the drawer", async () => {
      const user = await openRenameDrawer();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "new-valid-name");
      await user.click(screen.getByRole("button", { name: /save/i }));
      expect(sdk.editNamespace).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { tenant: "tenant-456" },
          body: { name: "new-valid-name" },
          throwOnError: true,
        }),
      );
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("shows a generic error alert when rename fails", async () => {
      sdk.editNamespace.mockRejectedValue(new Error("server error"));
      const user = await openRenameDrawer();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "new-valid-name");
      await user.click(screen.getByRole("button", { name: /save/i }));
      expect(await screen.findByRole("alert")).toBeInTheDocument();
    });

    it("resets to currentName when the drawer is reopened", async () => {
      const user = await openRenameDrawer();
      const input = screen.getByLabelText(/namespace name/i);
      await user.clear(input);
      await user.type(input, "changed-name");
      await user.click(screen.getByRole("button", { name: /cancel/i }));
      await user.click(
        screen.getByRole("button", { name: /rename namespace/i }),
      );
      expect(screen.getByLabelText(/namespace name/i)).toHaveValue(
        "my-namespace",
      );
    });
  });
});
