import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http } from "msw";
import { ChatwootContext, type ChatwootHandle } from "@/hooks/useChatwoot";
import { createTestWrapper } from "@/tests/wrapper";
import { server, jsonWithTotal } from "@/tests/msw";
import AccountMenuItems from "../AccountMenuItems";

function renderItems(
  status: ChatwootHandle["status"] | null,
  handlers: { openWidget?: () => void; onHelpNeedsPlan?: () => void } = {},
) {
  const items = (
    <AccountMenuItems
      onDone={vi.fn()}
      onHelpNeedsPlan={handlers.onHelpNeedsPlan}
    />
  );
  return render(
    status === null ? (
      items
    ) : (
      <ChatwootContext.Provider
        value={{ status, openWidget: handlers.openWidget ?? vi.fn() }}
      >
        {items}
      </ChatwootContext.Provider>
    ),
    { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
  );
}

beforeEach(() => {
  server.use(http.get("*/api/namespaces", () => jsonWithTotal([])));
});

describe("AccountMenuItems getting help", () => {
  it("is left out where there is no support provider", () => {
    renderItems(null);

    expect(screen.queryByText("Getting Help")).not.toBeInTheDocument();
  });

  it("is left out when support is unavailable", () => {
    renderItems("unavailable");

    expect(screen.queryByText("Getting Help")).not.toBeInTheDocument();
  });

  it("points outside cloud to the issue tracker, in a new tab", () => {
    renderItems("non-cloud");

    const link = screen.getByRole("link", { name: "Getting Help" });
    expect(link).toHaveAttribute(
      "href",
      "https://github.com/shellhub-io/shellhub/issues/new/choose",
    );
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("opens the chat on a plan with support", async () => {
    const user = userEvent.setup();
    const openWidget = vi.fn();
    renderItems("ready", { openWidget });

    await user.click(screen.getByRole("button", { name: "Getting Help" }));

    expect(openWidget).toHaveBeenCalledOnce();
  });

  it("offers the upgrade on a plan without support, never the chat", async () => {
    const user = userEvent.setup();
    const openWidget = vi.fn();
    const onHelpNeedsPlan = vi.fn();
    renderItems("no-subscription", { openWidget, onHelpNeedsPlan });

    await user.click(screen.getByRole("button", { name: "Getting Help" }));

    expect(onHelpNeedsPlan).toHaveBeenCalledOnce();
    expect(openWidget).not.toHaveBeenCalled();
  });

  it("does nothing while the chat is still loading", async () => {
    const user = userEvent.setup();
    const openWidget = vi.fn();
    renderItems("loading", { openWidget });

    await user.click(screen.getByRole("button", { name: "Getting Help" }));

    expect(openWidget).not.toHaveBeenCalled();
  });
});
