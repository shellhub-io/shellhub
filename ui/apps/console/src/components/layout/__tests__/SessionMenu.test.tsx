import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import SessionMenu from "../SessionMenu";

beforeEach(() => {
  server.use(http.get("*/api/namespaces", () => jsonWithTotal([])));
});

async function openMenu() {
  const user = userEvent.setup();
  render(<SessionMenu placement="tabStrip" />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
  await user.click(screen.getByRole("button", { name: /account menu for/i }));
  return within(await screen.findByLabelText("Account"));
}

describe("SessionMenu", () => {
  it("heads the panel with who is signed in and their email", async () => {
    seedAuthStore({ user: "alice", name: "Alice", email: "alice@example.com" });
    const panel = await openMenu();

    expect(panel.getByText("alice")).toBeInTheDocument();
    expect(panel.getByText("alice@example.com")).toBeInTheDocument();
  });

  it("says the user is logged in when the email is already the name shown", async () => {
    seedAuthStore({ user: null, name: null, email: "alice@example.com" });
    const panel = await openMenu();

    expect(panel.getByText("alice@example.com")).toBeInTheDocument();
    expect(panel.getByText("Logged in")).toBeInTheDocument();
  });
});
