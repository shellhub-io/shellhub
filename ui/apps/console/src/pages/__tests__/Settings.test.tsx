import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("@/client", () => ({
  getNamespace: vi.fn(),
  editNamespace: vi.fn(),
  deleteNamespace: vi.fn(),
  leaveNamespace: vi.fn(),
  setDeviceAutoAccept: vi.fn(),
}));

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn(() => actual.getConfig()) };
});

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: vi.fn(() => ({
    namespace: {
      name: "my-ns",
      tenant_id: "00000000-0000-4000-0000-000000000000",
      type: "personal",
      settings: {
        session_record: false,
        device_auto_accept: false,
        connection_announcement: "",
      },
    },
  })),
}));

vi.mock("@/hooks/useNamespaceMutations", () => ({
  useEditNamespace: vi.fn(() => ({ mutateAsync: vi.fn() })),
  useDeleteNamespace: vi.fn(() => ({ mutateAsync: vi.fn() })),
  useLeaveNamespace: vi.fn(() => ({ mutateAsync: vi.fn() })),
  useSetDeviceAutoAccept: vi.fn(() => ({ mutateAsync: vi.fn() })),
}));

vi.mock("@/components/billing/BillingSection", () => ({
  default: () => null,
}));

vi.mock("@/components/common/CopyButton", () => ({
  default: () => null,
}));

/* ------------------------------------------------------------------ */
/* Deferred imports (must come after vi.mock calls)                    */
/* ------------------------------------------------------------------ */

import Settings from "../Settings";
import * as SettingsCardModule from "@/components/common/SettingsCard";
import * as SettingsRowModule from "@/components/common/SettingsRow";
import { getConfig, defaultConfig } from "@/env";

const mockedGetConfig = vi.mocked(getConfig);

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function seedAuthStore() {
  useAuthStore.setState({
    name: "Test User",
    user: "testuser",
    username: "testuser",
    email: "test@example.com",
    recoveryEmail: "",
    mfaEnabled: false,
    loading: false,
    token: "tok",
    userId: "uid-1",
    tenant: "00000000-0000-4000-0000-000000000000",
    role: "owner",
  });
}

function renderSettings() {
  return render(
    <MemoryRouter>
      <Settings />
    </MemoryRouter>,
  );
}

/* ------------------------------------------------------------------ */
/* Setup / teardown                                                    */
/* ------------------------------------------------------------------ */

afterEach(cleanup);

beforeEach(() => {
  mockedGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore();
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("Settings", () => {
  describe("shared component usage", () => {
    it("uses the shared SettingsCard component (not a local copy)", () => {
      const spy = vi.spyOn(SettingsCardModule, "default");
      renderSettings();
      expect(spy).toHaveBeenCalled();
      spy.mockRestore();
    });

    it("uses the shared SettingsRow component (not a local copy)", () => {
      const spy = vi.spyOn(SettingsRowModule, "default");
      renderSettings();
      expect(spy).toHaveBeenCalled();
      spy.mockRestore();
    });
  });

  describe("renders settings sections", () => {
    it("shows the General card heading", () => {
      renderSettings();
      expect(screen.getByRole("heading", { name: /^general$/i })).toBeInTheDocument();
    });

    it("shows the SSH card heading", () => {
      renderSettings();
      expect(screen.getByRole("heading", { name: /^ssh$/i })).toBeInTheDocument();
    });

    it("shows the Devices card heading", () => {
      renderSettings();
      expect(screen.getByRole("heading", { name: /^devices$/i })).toBeInTheDocument();
    });

    it("shows the Danger Zone card heading", () => {
      renderSettings();
      expect(screen.getByRole("heading", { name: /danger zone/i })).toBeInTheDocument();
    });

    it("renders the namespace name", () => {
      renderSettings();
      expect(screen.getByText("my-ns")).toBeInTheDocument();
    });

    it("renders the tenant ID", () => {
      renderSettings();
      expect(
        screen.getByText("00000000-0000-4000-0000-000000000000"),
      ).toBeInTheDocument();
    });
  });

  describe("danger zone", () => {
    it("renders the Delete button for owners", () => {
      renderSettings();
      expect(screen.getByRole("button", { name: /delete/i })).toBeInTheDocument();
    });
  });
});
