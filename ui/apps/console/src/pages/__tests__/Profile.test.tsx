import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import Profile from "../Profile";
import { getConfig, defaultConfig } from "@/env";
import { seedAuthStore } from "@/tests/seedAuthStore";

const mockGetConfig = vi.mocked(getConfig);

function renderProfile() {
  return render(
    <MemoryRouter>
      <Profile />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.patch("*/api/users", () => new HttpResponse(null, { status: 204 })),
  );
});

describe("Profile", () => {
  describe("renders profile sections", () => {
    it("renders the signed-in user's details", () => {
      renderProfile();
      expect(screen.getByText("Admin User")).toBeInTheDocument();
      expect(screen.getByText("admin@test.com")).toBeInTheDocument();
      expect(
        screen.getAllByText("recovery@test.com").length,
      ).toBeGreaterThanOrEqual(1);
    });

    it("shows 'Not set' when recovery email is absent", () => {
      seedAuthStore({ recoveryEmail: "" });
      renderProfile();
      expect(screen.getByText(/not set/i)).toBeInTheDocument();
    });
  });

  describe("SSO users", () => {
    it("hides password and MFA controls, showing the managed-by-IdP notice", () => {
      seedAuthStore({ origin: "saml" });
      renderProfile();

      expect(
        screen.getByText(/managed by your identity provider/i),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /^change password$/i }),
      ).not.toBeInTheDocument();
    });

    it("shows the password control for local users", () => {
      seedAuthStore({ origin: "local" });
      renderProfile();

      expect(
        screen.getByRole("button", { name: /^change password$/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByText(/managed by your identity provider/i),
      ).not.toBeInTheDocument();
    });
  });

  describe("ChangePasswordDrawer", () => {
    async function openChangePasswordDrawer() {
      const user = userEvent.setup();
      renderProfile();
      await user.click(
        screen.getByRole("button", { name: /^change password$/i }),
      );
      return user;
    }

    function getDrawerSubmitButton() {
      const all = screen.getAllByRole("button", { name: /^change password$/i });
      return all[all.length - 1];
    }

    it("disables the submit button when fields are empty", async () => {
      await openChangePasswordDrawer();
      expect(getDrawerSubmitButton()).toBeDisabled();
    });

    it("enables the submit button only when all three fields contain valid values", async () => {
      const user = await openChangePasswordDrawer();

      await user.type(screen.getByLabelText(/current password/i), "oldpass1");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );

      expect(getDrawerSubmitButton()).toBeEnabled();
    });

    it("shows 'Current password is incorrect.' on 403", async () => {
      server.use(
        http.patch("*/api/users", () =>
          HttpResponse.json({}, { status: 403 }),
        ),
      );
      const user = await openChangePasswordDrawer();

      await user.type(screen.getByLabelText(/current password/i), "wrong");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );
      await user.click(getDrawerSubmitButton());

      expect(
        await screen.findByText(/current password is incorrect/i),
      ).toBeInTheDocument();
    });

    it("shows success message after a successful password change", async () => {
      const user = await openChangePasswordDrawer();

      await user.type(screen.getByLabelText(/current password/i), "oldpass1");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );
      await user.click(getDrawerSubmitButton());

      expect(
        await screen.findByText(/password changed successfully/i),
      ).toBeInTheDocument();
    });

    it("resets the form when the drawer is reopened", async () => {
      const user = await openChangePasswordDrawer();

      await user.type(screen.getByLabelText(/current password/i), "somevalue");

      await user.click(screen.getByRole("button", { name: /cancel/i }));
      await user.click(
        screen.getByRole("button", { name: /^change password$/i }),
      );

      await waitFor(() => {
        expect(screen.getByLabelText(/current password/i)).toHaveValue("");
      });
    });
  });
});
