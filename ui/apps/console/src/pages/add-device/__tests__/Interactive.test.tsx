import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import Interactive from "../Interactive";

function renderPage() {
  return render(
    <ClipboardProvider>
      <Interactive />
    </ClipboardProvider>,
    { wrapper: createTestWrapper({ initialEntries: ["/devices/add"] }) },
  );
}

async function pickMethod(label: string, user = userEvent.setup()) {
  await user.click(screen.getByRole("button", { name: /Show all methods/ }));
  await user.click(screen.getByRole("radio", { name: new RegExp(label) }));
}

beforeEach(() => {
  seedAuthStore();
});

describe("Interactive", () => {
  it("installs with no credential where the installer can pair", () => {
    renderPage();

    const command = screen.getByText(/install\.sh/);
    expect(command).not.toHaveTextContent("TENANT_ID");
    expect(command).not.toHaveTextContent("PROVISIONING_KEY");
  });

  it("puts the tenant on the command for a method that cannot pair", async () => {
    const user = userEvent.setup();
    renderPage();

    await pickMethod("Snap", user);

    expect(screen.getByText(/install\.sh/)).toHaveTextContent(
      /INSTALL_METHOD=snap TENANT_ID=\S+/,
    );
  });

  it("points a manual method at its guide instead of a command", async () => {
    const user = userEvent.setup();
    renderPage();

    await pickMethod("Yocto", user);

    expect(screen.queryByText(/install\.sh/)).not.toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /Yocto Project guide/ }),
    ).toBeInTheDocument();
  });
});
