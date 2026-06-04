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

    it("renders the child element", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      expect(screen.getByRole("button", { name: "Add Key" })).toBeInTheDocument();
    });

    it("does not wrap children in a disabled container", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      expect(button.closest("[aria-disabled]")).toBeNull();
    });

    it("allows the child button to be clicked", async () => {
      const user = userEvent.setup();
      let clicked = false;
      render(
        <RestrictedAction action="publicKey:create">
          <button onClick={() => { clicked = true; }}>Add Key</button>
        </RestrictedAction>,
      );
      await user.click(screen.getByRole("button", { name: "Add Key" }));
      expect(clicked).toBe(true);
    });
  });

  describe("when the user lacks permission", () => {
    beforeEach(() => {
      useAuthStore.setState({ role: "observer" });
    });

    it("still renders the child element (visible but restricted)", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      expect(screen.getByRole("button", { name: "Add Key" })).toBeInTheDocument();
    });

    it("wraps children in a container with aria-disabled=true", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      const wrapper = button.closest("[aria-disabled='true']");
      expect(wrapper).toBeInTheDocument();
    });

    it("shows the default restriction message as a title tooltip", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      const wrapper = button.closest("[title]");
      expect(wrapper).toHaveAttribute("title", "You don't have permission to perform this action.");
    });

    it("shows a custom message when provided", () => {
      render(
        <RestrictedAction action="publicKey:create" message="Admins only.">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      const wrapper = button.closest("[title]");
      expect(wrapper).toHaveAttribute("title", "Admins only.");
    });

    it("prevents click events on children via pointer-events-none", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      const inner = button.closest(".pointer-events-none");
      expect(inner).toBeInTheDocument();
    });

    it("blocks keyboard interaction via inert attribute on inner wrapper", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      const inner = button.closest("[inert]");
      expect(inner).toBeInTheDocument();
    });

    it("applies cursor-not-allowed to the outer wrapper", () => {
      render(
        <RestrictedAction action="publicKey:create">
          <button>Add Key</button>
        </RestrictedAction>,
      );
      const button = screen.getByRole("button", { name: "Add Key" });
      const outer = button.closest(".cursor-not-allowed");
      expect(outer).toBeInTheDocument();
    });
  });

  describe("role transitions", () => {
    it("switches from restricted to allowed when role upgrades", () => {
      useAuthStore.setState({ role: "observer" });
      const { rerender } = render(
        <RestrictedAction action="device:remove">
          <button>Delete</button>
        </RestrictedAction>,
      );
      expect(screen.getByRole("button").closest("[aria-disabled='true']")).toBeInTheDocument();

      useAuthStore.setState({ role: "administrator" });
      rerender(
        <RestrictedAction action="device:remove">
          <button>Delete</button>
        </RestrictedAction>,
      );
      expect(screen.getByRole("button").closest("[aria-disabled]")).toBeNull();
    });
  });
});
