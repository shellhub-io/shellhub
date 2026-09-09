import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useAuthStore } from "@/stores/authStore";
import RestrictedAction from "../RestrictedAction";

beforeEach(() => {
  useAuthStore.setState({ role: null });
});

describe("RestrictedAction", () => {
  describe("when the user has permission", () => {
    beforeEach(() => {
      useAuthStore.setState({ role: "administrator" });
    });

    it("renders the child undisabled and lets it be clicked", async () => {
      const user = userEvent.setup();
      let clicked = false;
      render(
        <RestrictedAction action="publicKey:create">
          <button
            type="button"
            onClick={() => {
              clicked = true;
            }}
          >
            Add Key
          </button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      expect(button.closest("[aria-disabled]")).toBeNull();

      await user.click(button);
      expect(clicked).toBe(true);
    });
  });

  describe("when the user lacks permission", () => {
    beforeEach(() => {
      useAuthStore.setState({ role: "observer" });
    });

    it("marks the wrapper aria-disabled, inert, and titled with the default message", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button type="button">Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      expect(button.closest("[aria-disabled='true']")).toBeInTheDocument();
      expect(button.closest("[inert]")).toBeInTheDocument();
      expect(button.closest("[title]")).toHaveAttribute(
        "title",
        "You don't have permission to perform this action.",
      );
    });

    it("shows a custom message when provided", () => {
      render(
        <RestrictedAction action="publicKey:create" message="Admins only.">
          <button type="button">Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      expect(button.closest("[title]")).toHaveAttribute("title", "Admins only.");
    });
  });

  describe("role transitions", () => {
    it("switches from restricted to allowed when role upgrades", () => {
      useAuthStore.setState({ role: "observer" });
      const { rerender } = render(
        <RestrictedAction action="device:remove">
          <button type="button">Delete</button>
        </RestrictedAction>,
      );
      expect(
        screen.getByRole("button").closest("[aria-disabled='true']"),
      ).toBeInTheDocument();

      useAuthStore.setState({ role: "administrator" });
      rerender(
        <RestrictedAction action="device:remove">
          <button type="button">Delete</button>
        </RestrictedAction>,
      );
      expect(screen.getByRole("button").closest("[aria-disabled]")).toBeNull();
    });
  });
});
