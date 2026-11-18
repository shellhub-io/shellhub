import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Route } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { renderWithSignInPage } from "@/tests/renderWithSignInPage";
import { getConfig, defaultConfig } from "@/env";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockInfo } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";
import ProtectedRoute from "@/components/common/ProtectedRoute";

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
  return renderWithSignInPage(
    <Route element={<ProtectedRoute />}>
      <Route path="/account" element={<AccountLayout />}>
        <Route path="profile" element={<AccountProfile />} />
        <Route path="security" element={<AccountSecurity />} />
        <Route path="danger-zone" element={<AccountDangerZone />} />
      </Route>
    </Route>,
    path,
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
      expect(
        screen.getByRole("link", { name: "Security" }),
      ).toBeInTheDocument();
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

    it("names the lockout a wrong current password starts", async () => {
      server.use(
        http.patch("*/api/users", () =>
          HttpResponse.json(
            { message: "too many wrong attempts" },
            {
              status: 429,
              headers: {
                "X-Account-Lockout": String(
                  Math.floor(Date.now() / 1000) + 270,
                ),
              },
            },
          ),
        ),
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
        await screen.findByText(
          "Too many wrong attempts. Try again in 4 minutes.",
        ),
      ).toBeInTheDocument();
    });

    it("signs out and asks to sign in with the new password after a successful change", async () => {
      server.use(http.get("*/info", () => HttpResponse.json(mockInfo())));
      const user = await openChangePasswordModal();

      await user.type(screen.getByLabelText(/current password/i), "oldpass1");
      await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
      await user.type(
        screen.getByLabelText(/confirm new password/i),
        "newpass123",
      );
      await user.click(getModalSubmitButton());

      expect(
        await screen.findByText(
          "Password changed. Sign in with your new password.",
        ),
      ).toBeInTheDocument();
      expect(useAuthStore.getState().token).toBeNull();
    });

    describe("on an account with MFA", () => {
      async function fillPasswords() {
        seedAuthStore({ mfaEnabled: true });
        const user = await openChangePasswordModal();
        await user.type(screen.getByLabelText(/current password/i), "oldpass1");
        await user.type(screen.getByLabelText(/^new password$/i), "newpass123");
        await user.type(
          screen.getByLabelText(/confirm new password/i),
          "newpass123",
        );
        return user;
      }

      it("keeps the submit button disabled until a code is entered", async () => {
        const user = await fillPasswords();
        expect(getModalSubmitButton()).toBeDisabled();

        await user.type(
          screen.getByLabelText("Authenticator or Recovery Code"),
          "123456",
        );
        expect(getModalSubmitButton()).toBeEnabled();
      });

      it("sends the code, a recovery code included, with the new password", async () => {
        let body: unknown;
        server.use(
          http.get("*/info", () => HttpResponse.json(mockInfo())),
          http.patch("*/api/users", async ({ request }) => {
            body = await request.json();
            return HttpResponse.json({});
          }),
        );
        const user = await fillPasswords();
        await user.type(
          screen.getByLabelText("Authenticator or Recovery Code"),
          "ABCDEFGHIJKLMNOP",
        );
        await user.click(getModalSubmitButton());

        await waitFor(() =>
          expect(body).toEqual({
            current_password: "oldpass1",
            password: "newpass123",
            mfa_code: "ABCDEFGHIJKLMNOP",
          }),
        );
      });

      it("says why the server refused the code", async () => {
        server.use(
          http.patch("*/api/users", () =>
            HttpResponse.json(
              { message: "bad request", fields: { mfa_code: "is invalid" } },
              { status: 400 },
            ),
          ),
        );
        const user = await fillPasswords();
        await user.type(
          screen.getByLabelText("Authenticator or Recovery Code"),
          "000000",
        );
        await user.click(getModalSubmitButton());

        expect(
          await screen.findByText("The code is incorrect."),
        ).toBeInTheDocument();
      });
    });

    it("asks no code from an account without MFA", async () => {
      await openChangePasswordModal();
      expect(
        screen.queryByLabelText("Authenticator or Recovery Code"),
      ).not.toBeInTheDocument();
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

  describe("DeleteAccountDialog", () => {
    const deleteUser = vi.fn(() => Promise.resolve());
    const storeDeleteUser = useAuthStore.getState().deleteUser;

    afterEach(() => {
      useAuthStore.setState({ deleteUser: storeDeleteUser });
    });

    async function openDeleteDialog() {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      const user = userEvent.setup();
      renderAt("/account/danger-zone");
      await user.click(
        await screen.findByRole("button", { name: "Delete account" }),
      );
      return user;
    }

    function getConfirmButton() {
      return within(screen.getByRole("dialog")).getByRole("button", {
        name: "Delete account",
      });
    }

    async function submitFactor(
      user: ReturnType<typeof userEvent.setup>,
      label: "Password" | "Verification code",
      value: string,
    ) {
      await user.type(screen.getByLabelText(label), value);
      await user.click(getConfirmButton());
    }

    it("asks a user without MFA for the password and sends it", async () => {
      useAuthStore.setState({ deleteUser });
      const user = await openDeleteDialog();

      expect(getConfirmButton()).toBeDisabled();
      expect(
        screen.queryByLabelText("Verification code"),
      ).not.toBeInTheDocument();

      await submitFactor(user, "Password", "correct-horse");

      await waitFor(() =>
        expect(deleteUser).toHaveBeenCalledWith({ password: "correct-horse" }),
      );
    });

    it("sends a password made only of spaces as typed", async () => {
      useAuthStore.setState({ deleteUser });
      const user = await openDeleteDialog();

      await submitFactor(user, "Password", "     ");

      await waitFor(() =>
        expect(deleteUser).toHaveBeenCalledWith({ password: "     " }),
      );
    });

    it("asks a user with MFA for the code instead of the password", async () => {
      seedAuthStore({ mfaEnabled: true });
      useAuthStore.setState({ deleteUser });
      const user = await openDeleteDialog();

      expect(screen.queryByLabelText("Password")).not.toBeInTheDocument();
      await submitFactor(user, "Verification code", "123456");

      await waitFor(() =>
        expect(deleteUser).toHaveBeenCalledWith({ code: "123456" }),
      );
    });

    it("sends the factor in the request body", async () => {
      let body: unknown;
      server.use(
        http.delete("*/api/user", async ({ request }) => {
          body = await request.json();
          return HttpResponse.json({ message: "forbidden" }, { status: 403 });
        }),
      );
      const user = await openDeleteDialog();

      await submitFactor(user, "Password", "wrong");

      await waitFor(() => expect(body).toEqual({ password: "wrong" }));
    });

    it("says the password is incorrect when the server refuses it", async () => {
      server.use(
        http.delete("*/api/user", () =>
          HttpResponse.json({ message: "forbidden" }, { status: 403 }),
        ),
      );
      const user = await openDeleteDialog();

      await submitFactor(user, "Password", "wrong");

      expect(
        await screen.findByText("The password is incorrect."),
      ).toBeInTheDocument();
      expect(screen.getByLabelText("Password")).toHaveValue("");
    });

    it("says the account owns namespaces when the server refuses the deletion for it", async () => {
      server.use(
        http.delete("*/api/user", () =>
          HttpResponse.json(
            { message: "user still has a owned namespace" },
            { status: 403 },
          ),
        ),
      );
      const user = await openDeleteDialog();

      await submitFactor(user, "Password", "correct-horse");

      expect(
        await screen.findByText(
          "You cannot delete your account while you own namespaces.",
        ),
      ).toBeInTheDocument();
      expect(
        screen.queryByText("The password is incorrect."),
      ).not.toBeInTheDocument();
    });

    it("says the code is incorrect when the server refuses it", async () => {
      seedAuthStore({ mfaEnabled: true });
      server.use(
        http.delete("*/api/user", () =>
          HttpResponse.json({ message: "forbidden" }, { status: 403 }),
        ),
      );
      const user = await openDeleteDialog();

      await submitFactor(user, "Verification code", "000000");

      expect(
        await screen.findByText("The code is incorrect."),
      ).toBeInTheDocument();
    });

    it("names the lockout on 429", async () => {
      const end = Math.floor(Date.now() / 1000) + 270;
      server.use(
        http.delete("*/api/user", () =>
          HttpResponse.json(
            { message: "too many requests" },
            { status: 429, headers: { "X-Account-Lockout": String(end) } },
          ),
        ),
      );
      const user = await openDeleteDialog();

      await submitFactor(user, "Password", "wrong");

      expect(
        await screen.findByText(
          "Too many wrong attempts. Try again in 4 minutes.",
        ),
      ).toBeInTheDocument();
    });

    it("refuses an SSO user, who has no factor to give, and sends nothing", async () => {
      seedAuthStore({ origin: "saml" });
      useAuthStore.setState({ deleteUser });
      const user = await openDeleteDialog();

      expect(screen.queryByLabelText("Password")).not.toBeInTheDocument();
      expect(
        screen.queryByLabelText("Verification code"),
      ).not.toBeInTheDocument();
      expect(
        screen.getByText(/ask an administrator to delete it/i),
      ).toBeInTheDocument();
      expect(getConfirmButton()).toBeDisabled();
      await user.click(getConfirmButton());
      expect(deleteUser).not.toHaveBeenCalled();
    });
  });
});
