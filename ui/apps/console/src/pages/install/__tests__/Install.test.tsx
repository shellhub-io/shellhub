import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import Install from "../index";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function renderPage() {
  return render(<Install />, {
    wrapper: createTestWrapper({ initialEntries: ["/install"] }),
  });
}

describe("Install", () => {
  it("pairs by default, with no credential on the command", () => {
    renderPage();

    expect(screen.getByText(/install\.sh/)).not.toHaveTextContent("TENANT_ID");
    expect(
      screen.getByText(/the device joins the namespace you pick/i),
    ).toBeInTheDocument();
  });

  it("names snap as the method that needs the tenant and lands pending", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(screen.getByRole("tab", { name: /Snap/ }));

    expect(screen.getByText(/install\.sh/)).toHaveTextContent(
      "TENANT_ID=<your-tenant-id>",
    );
    expect(
      screen.getByText(/waits in Pending until somebody accepts it/i),
    ).toBeInTheDocument();
  });
});
