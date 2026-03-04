import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useRef, useState } from "react";
import { useFocusTrap } from "../useFocusTrap";

afterEach(cleanup);

// Helper component that exercises the hook
function Trap({
  active,
  buttons = ["First", "Second", "Third"],
}: {
  active: boolean;
  buttons?: string[];
}) {
  const ref = useRef<HTMLDivElement>(null);
  useFocusTrap(ref, active);

  return (
    <div>
      <button>Outside</button>
      <div ref={ref} data-testid="container">
        {buttons.map((label) => (
          <button key={label}>{label}</button>
        ))}
      </div>
    </div>
  );
}

// Component that toggles `active` to test deactivation / restore
function ToggleTrap() {
  const [active, setActive] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useFocusTrap(ref, active);

  return (
    <div>
      <button data-testid="trigger" onClick={() => setActive((v) => !v)}>
        Toggle
      </button>
      <div ref={ref} data-testid="container">
        <button>Inside</button>
      </div>
    </div>
  );
}

describe("useFocusTrap", () => {
  beforeEach(() => {
    // requestAnimationFrame runs the callback synchronously in jsdom with fake timers,
    // but without fake timers we can mock it to run immediately.
    vi.spyOn(window, "requestAnimationFrame").mockImplementation((cb) => {
      cb(0);
      return 0;
    });
    vi.spyOn(window, "cancelAnimationFrame").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("when active=false", () => {
    it("does not move focus into the container", () => {
      render(<Trap active={false} />);
      // Default focus is on body — nothing inside the container should be focused
      expect(document.activeElement).toBe(document.body);
    });
  });

  describe("when active=true", () => {
    it("moves focus to the first focusable element", () => {
      render(<Trap active={true} />);
      expect(document.activeElement).toBe(screen.getByRole("button", { name: "First" }));
    });

    it("wraps Tab forward from the last element to the first", async () => {
      const user = userEvent.setup({ document });
      render(<Trap active={true} />);

      // Tab to Last (Second → Third)
      await user.tab();
      await user.tab();

      expect(document.activeElement).toBe(screen.getByRole("button", { name: "Third" }));

      // Tab again — should wrap to First
      await user.tab();
      expect(document.activeElement).toBe(screen.getByRole("button", { name: "First" }));
    });

    it("wraps Shift+Tab backward from the first element to the last", async () => {
      const user = userEvent.setup({ document });
      render(<Trap active={true} />);

      // Focus starts on First; Shift+Tab should wrap to Last
      expect(document.activeElement).toBe(screen.getByRole("button", { name: "First" }));
      await user.tab({ shift: true });
      expect(document.activeElement).toBe(screen.getByRole("button", { name: "Third" }));
    });

    it("does not trap Tab when focus is on a middle element", async () => {
      const user = userEvent.setup({ document });
      render(<Trap active={true} />);

      // Focus is on First; Tab moves to Second (no wrap)
      await user.tab();
      expect(document.activeElement).toBe(screen.getByRole("button", { name: "Second" }));
    });
  });

  describe("restoring focus on deactivation", () => {
    it("restores focus to the previously focused element when the trap is deactivated", async () => {
      const user = userEvent.setup({ document });
      render(<ToggleTrap />);

      const trigger = screen.getByTestId("trigger");
      trigger.focus();
      expect(document.activeElement).toBe(trigger);

      // Activate trap
      await user.click(trigger);
      expect(document.activeElement).toBe(screen.getByRole("button", { name: "Inside" }));

      // Deactivate trap — focus should return to trigger
      await user.click(trigger);
      expect(document.activeElement).toBe(trigger);
    });
  });

  describe("cleanup on unmount", () => {
    it("removes the keydown listener when the component unmounts", async () => {
      const removeListener = vi.spyOn(HTMLDivElement.prototype, "removeEventListener");

      const { unmount } = render(<Trap active={true} />);
      unmount();

      expect(removeListener).toHaveBeenCalledWith("keydown", expect.any(Function));
      removeListener.mockRestore();
    });
  });

  describe("edge case — no focusable elements", () => {
    it("handles Tab gracefully when the container has no focusable children", async () => {
      function EmptyTrap() {
        const ref = useRef<HTMLDivElement>(null);
        useFocusTrap(ref, true);
        return (
          <div ref={ref} data-testid="container">
            <span>Not focusable</span>
          </div>
        );
      }

      const user = userEvent.setup({ document });
      render(<EmptyTrap />);
      // Tab should not throw
      await expect(user.tab()).resolves.toBeUndefined();
    });
  });
});
