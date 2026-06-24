import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("@/client", () => ({
  login: vi.fn(),
  getUserInfo: vi.fn(),
  updateUser: vi.fn(),
  deleteUser: vi.fn(),
  authMfa: vi.fn(),
  mfaRecover: vi.fn(),
  requestResetMfa: vi.fn(),
  resetMfa: vi.fn(),
  resendEmail: vi.fn(),
  getInfo: vi.fn(),
  getSamlAuthUrl: vi.fn(),
  listNamespaces: vi.fn(),
}));

vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return { ...actual, getConfig: vi.fn(() => actual.getConfig()) };
});

vi.mock("@/utils/features", () => ({
  hasMfaSupport: vi.fn(() => false),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespaces: vi.fn(() => ({ namespaces: [] })),
}));

import Profile from "../Profile";
import * as SettingsCardModule from "@/components/common/SettingsCard";
import * as SettingsRowModule from "@/components/common/SettingsRow";
import { getConfig, defaultConfig } from "@/env";

const mockedGetConfig = vi.mocked(getConfig);

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function renderProfile() {
  return render(
    <MemoryRouter>
      <Profile />
    </MemoryRouter>,
  );
}

function seedAuthStore(overrides: Partial<{
  name: string;
  username: string;
  email: string;
  recoveryEmail: string;
  mfaEnabled: boolean;
}> = {}) {
  useAuthStore.setState({
    name: "Test User",
    user: "testuser",
    username: "testuser",
    email: "test@example.com",
    recoveryEmail: "recovery@example.com",
    mfaEnabled: false,
    loading: false,
    token: "tok",
    userId: "uid-1",
    tenant: "tenant-1",
    role: "owner",
    ...overrides,
  });
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

describe("Profile", () => {
  describe("shared component usage", () => {
    it("uses the shared SettingsCard component (not a local copy)", () => {
      // Spy on the shared SettingsCard to confirm it is called during render
      const spy = vi.spyOn(SettingsCardModule, "default");
      renderProfile();
      expect(spy).toHaveBeenCalled();
      spy.mockRestore();
    });

    it("uses the shared SettingsRow component (not a local copy)", () => {
      // Spy on the shared SettingsRow to confirm it is called during render
      const spy = vi.spyOn(SettingsRowModule, "default");
      renderProfile();
      expect(spy).toHaveBeenCalled();
      spy.mockRestore();
    });
  });

  describe("renders profile sections", () => {
    it("shows the Profile card heading", () => {
      renderProfile();
      // Multiple headings with "Profile" exist (PageHeader h1 + SettingsCard h3)
      const headings = screen.getAllByRole("heading", { name: /^profile$/i });
      expect(headings.length).toBeGreaterThanOrEqual(1);
    });

    it("shows the Security card heading", () => {
      renderProfile();
      expect(screen.getByRole("heading", { name: /^security$/i })).toBeInTheDocument();
    });

    it("shows the Danger Zone card heading", () => {
      renderProfile();
      expect(screen.getByRole("heading", { name: /danger zone/i })).toBeInTheDocument();
    });

    it("renders the user name in the Profile card", () => {
      renderProfile();
      expect(screen.getByText("Test User")).toBeInTheDocument();
    });

    it("renders the user email in the Profile card", () => {
      renderProfile();
      expect(screen.getByText("test@example.com")).toBeInTheDocument();
    });

    it("renders the recovery email in the Profile card", () => {
      renderProfile();
      // The recovery email may appear more than once (e.g. pre-filled in drawers)
      expect(screen.getAllByText("recovery@example.com").length).toBeGreaterThanOrEqual(1);
    });

    it("shows 'Not set' when recovery email is absent", () => {
      seedAuthStore({ recoveryEmail: "" });
      renderProfile();
      expect(screen.getByText(/not set/i)).toBeInTheDocument();
    });
  });

  describe("danger zone", () => {
    it("renders the delete account button", () => {
      renderProfile();
      expect(
        screen.getByRole("button", { name: /delete/i }),
      ).toBeInTheDocument();
    });
  });

  describe("page header", () => {
    it("renders the Edit Profile button in the header", () => {
      renderProfile();
      expect(
        screen.getByRole("button", { name: /edit profile/i }),
      ).toBeInTheDocument();
    });
  });
});
