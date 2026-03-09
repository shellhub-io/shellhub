import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { useAuthStore } from "../../../stores/authStore";
import { useNamespacesStore } from "../../../stores/namespacesStore";
import UserMenu from "../UserMenu";

afterEach(cleanup);

beforeEach(() => {
  useAuthStore.setState({
    token: "token",
    user: "alice",
    userId: "user-1",
    email: "alice@example.com",
    username: null,
    recoveryEmail: null,
    tenant: "t1",
    role: "owner",
    name: "Alice",
    loading: false,
  });

  useNamespacesStore.setState({
    namespaces: [{ tenant_id: "t1", name: "ns1" }] as never,
    currentNamespace: null,
    loading: false,
    loaded: true,
    error: null,
  });

  vi.clearAllMocks();
});

function renderMenu() {
  return render(
    <MemoryRouter>
      <UserMenu />
    </MemoryRouter>,
  );
}

async function openDropdown() {
  await userEvent.click(screen.getByRole("button", { name: /alice/i }));
}

describe("UserMenu", () => {
  describe("trigger button", () => {
    it("displays the username", () => {
      renderMenu();
      expect(screen.getByRole("button", { name: /alice/i })).toBeInTheDocument();
    });

    it("returns nothing when there is no logged-in user", () => {
      useAuthStore.setState({ user: null });
      const { container } = renderMenu();
      expect(container).toBeEmptyDOMElement();
    });
  });

  describe("dropdown — with namespaces", () => {
    it("shows Profile", async () => {
      renderMenu();
      await openDropdown();
      expect(screen.getByRole("button", { name: /profile/i })).toBeInTheDocument();
    });

    it("shows Settings", async () => {
      renderMenu();
      await openDropdown();
      expect(screen.getByRole("button", { name: /settings/i })).toBeInTheDocument();
    });

    it("shows Logout", async () => {
      renderMenu();
      await openDropdown();
      expect(screen.getByRole("button", { name: /logout/i })).toBeInTheDocument();
    });
  });

  describe("dropdown — without namespaces", () => {
    beforeEach(() => {
      useNamespacesStore.setState({ namespaces: [], loaded: true });
    });

    it("still shows Profile", async () => {
      renderMenu();
      await openDropdown();
      expect(screen.getByRole("button", { name: /profile/i })).toBeInTheDocument();
    });

    it("hides Settings", async () => {
      renderMenu();
      await openDropdown();
      expect(screen.queryByRole("button", { name: /settings/i })).not.toBeInTheDocument();
    });

    it("still shows Logout", async () => {
      renderMenu();
      await openDropdown();
      expect(screen.getByRole("button", { name: /logout/i })).toBeInTheDocument();
    });
  });
});
