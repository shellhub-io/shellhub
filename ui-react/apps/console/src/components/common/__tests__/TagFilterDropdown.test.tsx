import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { fireEvent } from "@testing-library/react";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useTags", () => ({
  useTags: vi.fn(),
}));

// useEscapeKey attaches a keydown listener to document; keep the real impl.
// The popover uses createPortal — RTL queries document.body, so portal
// content is reachable without any mock.

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useTags } from "@/hooks/useTags";
import TagFilterDropdown from "../TagFilterDropdown";

// ── Helpers ───────────────────────────────────────────────────────────────────

function defaultTagObjects(names: string[]) {
  return names.map((name) => ({ name }));
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useTags).mockReturnValue({
    tags: defaultTagObjects(["alpha", "beta", "gamma"]),
    totalCount: 3,
    isLoading: false,
    error: null,
  } as never);
});

function renderDropdown(
  overrides: Partial<{
    filterTags: string[];
    onAdd: (tag: string) => void;
    onRemove: (tag: string) => void;
    onClearAll: () => void;
    onManageTags: (() => void) | undefined;
  }> = {},
) {
  const defaults = {
    filterTags: [],
    onAdd: vi.fn(),
    onRemove: vi.fn(),
    onClearAll: vi.fn(),
    onManageTags: undefined,
  };
  const props = { ...defaults, ...overrides };
  return {
    onAdd: props.onAdd,
    onRemove: props.onRemove,
    onClearAll: props.onClearAll,
    onManageTags: props.onManageTags,
    ...render(<TagFilterDropdown {...props} />),
  };
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("TagFilterDropdown", () => {
  describe("trigger button", () => {
    it("renders the Tags trigger button", () => {
      renderDropdown();
      expect(screen.getByRole("button", { name: /tags/i })).toBeInTheDocument();
    });

    it("does not show a count badge when no tags are active", () => {
      renderDropdown({ filterTags: [] });
      // The badge would show a number; no digits should be in the button text
      const btn = screen.getByRole("button", { name: /tags/i });
      expect(btn.textContent?.match(/\d/)).toBeNull();
    });

    it("shows a count badge with the number of active filter tags", () => {
      renderDropdown({ filterTags: ["alpha", "beta"] });
      // The badge renders the count as text inside the button
      const btn = screen.getByRole("button", { name: /tags/i });
      expect(btn.textContent).toContain("2");
    });

    it("shows badge with count 1 when one tag is active", () => {
      renderDropdown({ filterTags: ["alpha"] });
      const btn = screen.getByRole("button", { name: /tags/i });
      expect(btn.textContent).toContain("1");
    });
  });

  describe("opening the popover", () => {
    it("opens the popover when the trigger is clicked", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(screen.getByPlaceholderText("Search tags...")).toBeInTheDocument();
    });

    it("renders all available tags in the list when opened", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(
        screen.getByRole("button", { name: /^alpha$/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /^beta$/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /^gamma$/i }),
      ).toBeInTheDocument();
    });

    it("clears the search state when reopened", async () => {
      renderDropdown();
      const trigger = screen.getByRole("button", { name: /tags/i });

      // Open and type
      await userEvent.click(trigger);
      const input = screen.getByPlaceholderText("Search tags...");
      await userEvent.type(input, "alp");
      expect(input).toHaveValue("alp");

      // Close by clicking trigger again
      await userEvent.click(trigger);
      // Reopen
      await userEvent.click(trigger);

      // Search should be cleared
      expect(screen.getByPlaceholderText("Search tags...")).toHaveValue("");
    });
  });

  describe("closing the popover", () => {
    it("closes the popover on Escape key", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(screen.getByPlaceholderText("Search tags...")).toBeInTheDocument();

      fireEvent.keyDown(document, { key: "Escape" });

      await waitFor(() => {
        expect(
          screen.queryByPlaceholderText("Search tags..."),
        ).not.toBeInTheDocument();
      });
    });

    it("toggles closed when trigger is clicked while open", async () => {
      renderDropdown();
      const trigger = screen.getByRole("button", { name: /tags/i });

      await userEvent.click(trigger);
      expect(screen.getByPlaceholderText("Search tags...")).toBeInTheDocument();

      await userEvent.click(trigger);
      expect(
        screen.queryByPlaceholderText("Search tags..."),
      ).not.toBeInTheDocument();
    });
  });

  describe("search filtering", () => {
    it("filters the tag list as the user types in the search input", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));

      await userEvent.type(
        screen.getByPlaceholderText("Search tags..."),
        "alp",
      );

      expect(
        screen.getByRole("button", { name: /^alpha$/i }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /^beta$/i }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /^gamma$/i }),
      ).not.toBeInTheDocument();
    });

    it("shows 'No tags found' when search matches nothing", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));

      await userEvent.type(
        screen.getByPlaceholderText("Search tags..."),
        "zzznomatch",
      );

      expect(screen.getByText("No tags found")).toBeInTheDocument();
    });

    it("is case-insensitive when filtering", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));

      await userEvent.type(
        screen.getByPlaceholderText("Search tags..."),
        "ALP",
      );

      expect(
        screen.getByRole("button", { name: /^alpha$/i }),
      ).toBeInTheDocument();
    });
  });

  describe("adding a tag", () => {
    it("calls onAdd when an inactive tag is clicked", async () => {
      const onAdd = vi.fn();
      renderDropdown({ filterTags: [], onAdd });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(screen.getByRole("button", { name: /^alpha$/i }));

      expect(onAdd).toHaveBeenCalledWith("alpha");
    });

    it("does not call onRemove when clicking an inactive tag", async () => {
      const onRemove = vi.fn();
      renderDropdown({ filterTags: [], onRemove });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(screen.getByRole("button", { name: /^alpha$/i }));

      expect(onRemove).not.toHaveBeenCalled();
    });
  });

  describe("removing a tag (toggle)", () => {
    it("calls onRemove when an active tag is clicked", async () => {
      const onRemove = vi.fn();
      renderDropdown({ filterTags: ["alpha"], onRemove });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(screen.getByRole("button", { name: /^alpha$/i }));

      expect(onRemove).toHaveBeenCalledWith("alpha");
    });

    it("does not call onAdd when clicking an active tag", async () => {
      const onAdd = vi.fn();
      renderDropdown({ filterTags: ["alpha"], onAdd });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(screen.getByRole("button", { name: /^alpha$/i }));

      expect(onAdd).not.toHaveBeenCalled();
    });
  });

  describe("clear all button", () => {
    it("is not rendered when no tags are active", async () => {
      renderDropdown({ filterTags: [] });
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(
        screen.queryByRole("button", { name: /clear all/i }),
      ).not.toBeInTheDocument();
    });

    it("is rendered when at least one tag is active", async () => {
      renderDropdown({ filterTags: ["alpha"] });
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(
        screen.getByRole("button", { name: /clear all/i }),
      ).toBeInTheDocument();
    });

    it("calls onClearAll when Clear all is clicked", async () => {
      const onClearAll = vi.fn();
      renderDropdown({ filterTags: ["alpha"], onClearAll });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(screen.getByRole("button", { name: /clear all/i }));

      expect(onClearAll).toHaveBeenCalledTimes(1);
    });

    it("closes the popover after clearing all", async () => {
      const onClearAll = vi.fn();
      renderDropdown({ filterTags: ["alpha"], onClearAll });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(screen.getByRole("button", { name: /clear all/i }));

      await waitFor(() => {
        expect(
          screen.queryByPlaceholderText("Search tags..."),
        ).not.toBeInTheDocument();
      });
    });
  });

  describe("manage tags button — absent", () => {
    it("is NOT rendered when onManageTags is not provided", async () => {
      renderDropdown({ onManageTags: undefined });
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(
        screen.queryByRole("button", { name: /manage tags/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("manage tags button — present", () => {
    it("IS rendered when onManageTags is provided", async () => {
      renderDropdown({ onManageTags: vi.fn() });
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      expect(
        screen.getByRole("button", { name: /manage tags/i }),
      ).toBeInTheDocument();
    });

    it("calls onManageTags when the button is clicked", async () => {
      const onManageTags = vi.fn();
      renderDropdown({ onManageTags });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(onManageTags).toHaveBeenCalledTimes(1);
    });

    it("closes the popover after clicking Manage tags", async () => {
      const onManageTags = vi.fn();
      renderDropdown({ onManageTags });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      await waitFor(() => {
        expect(
          screen.queryByPlaceholderText("Search tags..."),
        ).not.toBeInTheDocument();
      });
    });
  });
});
