import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import RuleDrawer from "../RuleDrawer";

vi.mock("@/hooks/useFirewallRuleMutations", () => ({
  useCreateFirewallRule: () => ({ mutateAsync: vi.fn() }),
  useUpdateFirewallRule: () => ({ mutateAsync: vi.fn() }),
}));

function renderDrawer() {
  return render(<RuleDrawer open onClose={vi.fn()} editRule={null} />);
}

describe("RuleDrawer — status toggle", () => {
  it("renders the status control as a switch defaulting to on", () => {
    renderDrawer();

    const toggle = screen.getByRole("switch", { name: /status/i });
    expect(toggle).toHaveAttribute("aria-checked", "true");
  });

  it("toggles the status off on click", async () => {
    const user = userEvent.setup();
    renderDrawer();

    const toggle = screen.getByRole("switch", { name: /status/i });
    await user.click(toggle);

    expect(toggle).toHaveAttribute("aria-checked", "false");
  });
});
