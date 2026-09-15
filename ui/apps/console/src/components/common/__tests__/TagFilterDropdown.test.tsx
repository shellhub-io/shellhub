import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { setTags } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import TagFilterDropdown from "../TagFilterDropdown";

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
    ...render(<TagFilterDropdown {...props} />, {
      wrapper: createTestWrapper(),
    }),
  };
}

describe("TagFilterDropdown", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setTags(["alpha", "beta", "gamma"]);
  });

  describe("trigger button", () => {
    it("does not show a count badge when no tags are active", () => {
      renderDropdown({ filterTags: [] });
      const btn = screen.getByRole("button", { name: /tags/i });
      expect(btn.textContent?.match(/\d/)).toBeNull();
    });

    it("shows a count badge with the number of active filter tags", () => {
      renderDropdown({ filterTags: ["alpha", "beta"] });
      const btn = screen.getByRole("button", { name: /tags/i });
      expect(btn.textContent).toContain("2");
    });
  });

  describe("opening the popover", () => {
    it("clears the search state when reopened", async () => {
      renderDropdown();
      const trigger = screen.getByRole("button", { name: /tags/i });

      await userEvent.click(trigger);
      const input = await screen.findByPlaceholderText("Search tags...");
      await userEvent.type(input, "alp");
      expect(input).toHaveValue("alp");

      await userEvent.click(trigger);
      await userEvent.click(trigger);

      expect(screen.getByPlaceholderText("Search tags...")).toHaveValue("");
    });
  });

  describe("search filtering", () => {
    it("filters the tag list as the user types in the search input", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await screen.findByRole("button", { name: /^alpha$/i });

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
      await screen.findByRole("button", { name: /^alpha$/i });

      await userEvent.type(
        screen.getByPlaceholderText("Search tags..."),
        "zzznomatch",
      );

      expect(screen.getByText("No tags found")).toBeInTheDocument();
    });

    it("is case-insensitive when filtering", async () => {
      renderDropdown();
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await screen.findByRole("button", { name: /^alpha$/i });

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
      await userEvent.click(
        await screen.findByRole("button", { name: /^alpha$/i }),
      );

      expect(onAdd).toHaveBeenCalledWith("alpha");
    });
  });

  describe("removing a tag (toggle)", () => {
    it("calls onRemove when an active tag is clicked", async () => {
      const onRemove = vi.fn();
      renderDropdown({ filterTags: ["alpha"], onRemove });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await userEvent.click(
        await screen.findByRole("button", { name: /^alpha$/i }),
      );

      expect(onRemove).toHaveBeenCalledWith("alpha");
    });
  });

  describe("clear all button", () => {
    it("calls onClearAll when Clear all is clicked", async () => {
      const onClearAll = vi.fn();
      renderDropdown({ filterTags: ["alpha"], onClearAll });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await screen.findByPlaceholderText("Search tags...");
      await userEvent.click(screen.getByRole("button", { name: /clear all/i }));

      expect(onClearAll).toHaveBeenCalledTimes(1);
    });
  });

  describe("manage tags button", () => {
    it("is NOT rendered when onManageTags is not provided", async () => {
      renderDropdown({ onManageTags: undefined });
      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await screen.findByPlaceholderText("Search tags...");
      expect(
        screen.queryByRole("button", { name: /manage tags/i }),
      ).not.toBeInTheDocument();
    });

    it("calls onManageTags when the button is clicked", async () => {
      const onManageTags = vi.fn();
      renderDropdown({ onManageTags });

      await userEvent.click(screen.getByRole("button", { name: /tags/i }));
      await screen.findByPlaceholderText("Search tags...");
      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(onManageTags).toHaveBeenCalledTimes(1);
    });
  });
});
