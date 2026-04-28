import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ChatwootContext, type ChatwootHandle } from "@/hooks/useChatwoot";
import SupportButton from "../SupportButton";
import "@/components/common/__tests__/helpers/setup-dialog";

function renderWithStatus(handle: ChatwootHandle) {
  return render(
    <ChatwootContext.Provider value={handle}>
      <SupportButton />
    </ChatwootContext.Provider>,
  );
}

function makeHandle(
  status: ChatwootHandle["status"],
  openWidget = vi.fn(),
): ChatwootHandle {
  return { status, openWidget };
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe("SupportButton", () => {
  describe("non-cloud branch", () => {
    it("renders a link pointing to the GitHub issue tracker", () => {
      renderWithStatus(makeHandle("non-cloud"));
      const link = screen.getByRole("link", {
        name: /report an issue on github/i,
      });
      expect(link).toHaveAttribute(
        "href",
        "https://github.com/shellhub-io/shellhub/issues/new",
      );
    });

    it("opens the link in a new tab with safe rel attributes", () => {
      renderWithStatus(makeHandle("non-cloud"));
      const link = screen.getByRole("link", {
        name: /report an issue on github/i,
      });
      expect(link).toHaveAttribute("target", "_blank");
      expect(link).toHaveAttribute("rel", "noopener noreferrer");
    });
  });

  describe("unavailable branch", () => {
    it("renders nothing when status is unavailable", () => {
      const { container } = renderWithStatus(makeHandle("unavailable"));
      expect(container).toBeEmptyDOMElement();
    });
  });

  describe("no-subscription branch", () => {
    it("renders an enabled button with the paywall aria-label", () => {
      renderWithStatus(makeHandle("no-subscription"));
      const button = screen.getByRole("button", {
        name: /paid plan required/i,
      });
      expect(button).not.toHaveAttribute("aria-disabled");
    });

    it("does not open the paywall dialog before the user clicks", () => {
      renderWithStatus(makeHandle("no-subscription"));
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("clicking the button opens the paywall dialog and never calls openWidget", async () => {
      const openWidget = vi.fn();
      renderWithStatus(makeHandle("no-subscription", openWidget));
      const user = userEvent.setup();
      await user.click(
        screen.getByRole("button", { name: /paid plan required/i }),
      );

      const dialog = await screen.findByRole("dialog");
      expect(
        within(dialog).getByText(/upgrade to access chat support/i),
      ).toBeInTheDocument();
      expect(openWidget).not.toHaveBeenCalled();
    });

    it("clicking the upgrade button opens the pricing page in a new tab", async () => {
      renderWithStatus(makeHandle("no-subscription"));
      const user = userEvent.setup();
      const openSpy = vi.spyOn(window, "open").mockImplementation(() => null);

      await user.click(
        screen.getByRole("button", { name: /paid plan required/i }),
      );
      const dialog = await screen.findByRole("dialog");
      await user.click(
        within(dialog).getByRole("button", { name: /upgrade/i }),
      );

      expect(openSpy).toHaveBeenCalledWith(
        "https://www.shellhub.io/pricing",
        "_blank",
        "noopener,noreferrer",
      );

      openSpy.mockRestore();
    });

    it("closes the paywall dialog automatically when status flips away from no-subscription", async () => {
      const initial = makeHandle("no-subscription");
      const { rerender } = render(
        <ChatwootContext.Provider value={initial}>
          <SupportButton />
        </ChatwootContext.Provider>,
      );

      const user = userEvent.setup();
      await user.click(
        screen.getByRole("button", { name: /paid plan required/i }),
      );
      expect(await screen.findByRole("dialog")).toBeInTheDocument();

      // Subscription becomes active mid-session — paywall is no longer
      // appropriate, so the dialog must close itself.
      rerender(
        <ChatwootContext.Provider value={makeHandle("ready")}>
          <SupportButton />
        </ChatwootContext.Provider>,
      );

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("ready branch", () => {
    it("renders an enabled button with the correct aria-label", () => {
      renderWithStatus(makeHandle("ready"));
      const button = screen.getByRole("button", { name: /open support chat/i });
      expect(button).not.toHaveAttribute("aria-disabled");
    });

    it("clicking the button calls openWidget", async () => {
      const openWidget = vi.fn();
      renderWithStatus(makeHandle("ready", openWidget));
      const user = userEvent.setup();
      await user.click(
        screen.getByRole("button", { name: /open support chat/i }),
      );
      expect(openWidget).toHaveBeenCalledTimes(1);
    });
  });

  describe("loading branch", () => {
    it("renders an aria-disabled button with the loading aria-label", () => {
      renderWithStatus(makeHandle("loading"));
      const button = screen.getByRole("button", {
        name: /loading support chat/i,
      });
      expect(button).toHaveAttribute("aria-disabled", "true");
    });

    it("clicking the loading button does not call openWidget or open the paywall", async () => {
      const openWidget = vi.fn();
      renderWithStatus(makeHandle("loading", openWidget));
      const user = userEvent.setup();
      await user.click(
        screen.getByRole("button", { name: /loading support chat/i }),
      );
      expect(openWidget).not.toHaveBeenCalled();
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });
});
