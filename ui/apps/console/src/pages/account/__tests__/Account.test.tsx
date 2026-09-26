import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { getConfig, defaultConfig } from "@/env";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useAuthStore } from "@/stores/authStore";

const viewport = await vi.hoisted(async () =>
  (await import("@/tests/viewport")).installViewport(),
);

import AccountLayout from "../AccountLayout";
import AccountProfile from "../AccountProfile";
import AccountSecurity from "../AccountSecurity";
import AccountDangerZone from "../AccountDangerZone";

const mockGetConfig = vi.mocked(getConfig);
const fetchUser = vi.fn(() => Promise.resolve());

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/account" element={<AccountLayout />}>
          <Route path="profile" element={<AccountProfile />} />
          <Route path="security" element={<AccountSecurity />} />
          <Route path="danger-zone" element={<AccountDangerZone />} />
        </Route>
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  viewport.wide = true;
  seedAuthStore();
  useAuthStore.setState({ fetchUser });
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.patch("*/api/users", () => new HttpResponse(null, { status: 204 })),
  );
});

describe("Account", () => {
  describe("layout", () => {
    it("opens the profile on a wide window", () => {
      renderAt("/account");
      expect(
        screen.getByRole("heading", { name: "Profile" }),
      ).toBeInTheDocument();
    });

    it("shows the section menu alone on a narrow window", () => {
      viewport.wide = false;
      renderAt("/account");
      expect(screen.getByRole("link", { name: "Security" })).toBeInTheDocument();
      expect(
        screen.queryByRole("heading", { name: "Profile" }),
      ).not.toBeInTheDocument();
    });

    it("refreshes the signed-in user when it opens", () => {
      renderAt("/account/profile");
      expect(fetchUser).toHaveBeenCalled();
    });
  });

  describe("renders profile sections", () => {
    it("renders the signed-in user's details", () => {
      renderAt("/account/profile");
      expect(screen.getByText("Admin User")).toBeInTheDocument();
      expect(screen.getByText("admin@test.com")).toBeInTheDocument();
      expect(
        screen.getAllByText("recovery@test.com").length,
      ).toBeGreaterThanOrEqual(1);
    });

    it("leaves out the deprecated username", () => {
      renderAt("/account/profile");
      expect(
        screen.queryByRole("group", { name: "Username" }),
      ).not.toBeInTheDocument();
    });

    it("shows 'Not set' when recovery email is absent", () => {
      seedAuthStore({ recoveryEmail: "" });
      renderAt("/account/profile");
      expect(screen.getByText(/not set/i)).toBeInTheDocument();
    });
  });

  describe("SSO users", () => {
    it("hides password and MFA controls, showing the managed-by-IdP notice", () => {
      seedAuthStore({ origin: "saml" });
      renderAt("/account/security");

      expect(
        screen.getByText(/managed by your identity provider/i),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /^change password$/i }),
      ).not.toBeInTheDocument();
    });

    it("shows the password control for local users", () => {
      seedAuthStore({ origin: "local" });
      renderAt("/account/security");

      expect(
        screen.getByRole("button", { name: /^change password$/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByText(/managed by your identity provider/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("ChangePasswordModal", () => {
    async function openChangePasswordModal() {
      const user = userEvent.setup();
      renderAt("/account/security");
      await user.click(
        screen.getByRole("button", { name: /^change password$/i }),
      );
      return user;
    }

    function getModalSubmitButton() {
      const all = screen.getAllByRole("button", { name: /^change password$/i });
      return all[all.length - 1];
    }

    it("disables the submit button when fields are empty", async () => {
      await openChangePasswordModal();
      expect(getModalSubmitButton()).toBeDisabled();
    });

    it("enables the submit button only when all three fields contain valid values", async () => {
      const user = await openChangePasswordModal();

      await user.type(screen.getByLabelText(/current password/i), "oldpass1");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );

      expect(getModalSubmitButton()).toBeEnabled();
    });

    it("shows 'Current password is incorrect.' on 403", async () => {
      server.use(
        http.patch("*/api/users", () => HttpResponse.json({}, { status: 403 })),
      );
      const user = await openChangePasswordModal();

      await user.type(screen.getByLabelText(/current password/i), "wrong");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );
      await user.click(getModalSubmitButton());

      expect(
        await screen.findByText(/current password is incorrect/i),
      ).toBeInTheDocument();
    });

    it("shows success message after a successful password change", async () => {
      const user = await openChangePasswordModal();

      await user.type(screen.getByLabelText(/current password/i), "oldpass1");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );
      await user.click(getModalSubmitButton());

      expect(
        await screen.findByText(/password changed successfully/i),
      ).toBeInTheDocument();
    });

    it("resets the form when the modal is reopened", async () => {
      const user = await openChangePasswordModal();

      await user.type(screen.getByLabelText(/current password/i), "somevalue");

      await user.click(screen.getByRole("button", { name: /cancel/i }));
      await user.click(screen.getByRole("button", { name: "Discard changes" }));
      await user.click(
        screen.getByRole("button", { name: /^change password$/i }),
      );

      await waitFor(() => {
        expect(screen.getByLabelText(/current password/i)).toHaveValue("");
      });
    });
  });
});
