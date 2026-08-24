import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse } from "@/tests/sdk";
import { mockNamespace, mockUserAuth } from "@/tests/factories";
import type { Namespace } from "@/client";
import UserMenu from "../UserMenu";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespaces: vi.fn(),
  }),
);

function mockNamespaces(namespaces: Namespace[]) {
  sdk.getNamespaces.mockResolvedValue(mockSdkResponse(namespaces));
}

function renderMenu() {
  return render(<UserMenu />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

async function openDropdown() {
  await userEvent.click(screen.getByRole("button", { name: /admin/i }));
}

describe("UserMenu", () => {
  const auth = mockUserAuth();

  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ ...auth });
    mockNamespaces([mockNamespace()]);
  });

  describe("trigger button", () => {
    it("displays the username", () => {
      renderMenu();
      expect(
        screen.getByRole("button", { name: /admin/i }),
      ).toBeInTheDocument();
    });

    it("returns nothing when there is no logged-in user", () => {
      useAuthStore.setState({ user: null, name: null, email: null });
      const { container } = renderMenu();
      expect(container).toBeEmptyDOMElement();
    });

    it("falls back to the name, not the email, when an SSO user has no username", () => {
      useAuthStore.setState({
        user: null,
        name: "Admin User",
        email: "admin@corp.example",
      });
      renderMenu();
      expect(
        screen.getByRole("button", { name: "Account menu for Admin User" }),
      ).toBeInTheDocument();
    });
  });

  describe("dropdown — with namespaces", () => {
    it("shows Profile, Settings and Logout", async () => {
      renderMenu();
      await openDropdown();
      expect(
        screen.getByRole("button", { name: /profile/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /settings/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /logout/i }),
      ).toBeInTheDocument();
    });
  });

  describe("dropdown — without namespaces", () => {
    it("shows Profile and Logout but hides Settings", async () => {
      mockNamespaces([]);
      renderMenu();
      await openDropdown();
      expect(
        screen.getByRole("button", { name: /profile/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /logout/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /settings/i }),
      ).not.toBeInTheDocument();
    });
  });
});
