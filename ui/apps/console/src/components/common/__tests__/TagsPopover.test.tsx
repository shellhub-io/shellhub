import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { fireEvent } from "@testing-library/react";

vi.mock("@/hooks/useTags", () => ({
  useTags: vi.fn(),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
}));

import { useTags } from "@/hooks/useTags";
import { useHasPermission } from "@/hooks/useHasPermission";
import TagsPopover from "../TagsPopover";

const mockAddTag = vi.fn();
const mockRemoveTag = vi.fn();

function renderPopover(
  overrides: {
    uid?: string;
    tags?: string[];
    onFilterTag?: () => void;
    editLabel?: string;
  } = {},
) {
  return render(
    <TagsPopover
      uid={overrides.uid ?? "entity-1"}
      tags={overrides.tags ?? []}
      addTag={mockAddTag}
      removeTag={mockRemoveTag}
      onFilterTag={overrides.onFilterTag ?? vi.fn()}
      {...(overrides.editLabel ? { editLabel: overrides.editLabel } : {})}
    />,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useTags).mockReturnValue({
    tags: [],
    totalCount: 0,
    isLoading: false,
    error: null,
  });
  vi.mocked(useHasPermission).mockReturnValue(true);
});

describe("TagsPopover", () => {
  describe("tag chip rendering", () => {
    it("renders each tag as a clickable button", () => {
      renderPopover({ tags: ["alpha", "beta"] });

      expect(
        screen.getByRole("button", { name: /^alpha$/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /^beta$/i }),
      ).toBeInTheDocument();
    });

    it("calls onFilterTag with the tag name when a chip is clicked", async () => {
      const onFilterTag = vi.fn();
      renderPopover({ tags: ["alpha"], onFilterTag });

      await userEvent.click(screen.getByRole("button", { name: /^alpha$/i }));

      expect(onFilterTag).toHaveBeenCalledWith("alpha");
    });

    it("shows 'No tags' text when there are no tags", () => {
      renderPopover({ tags: [] });
      expect(screen.getByText("No tags")).toBeInTheDocument();
    });
  });

  describe("edit button visibility", () => {
    it("shows the edit button when user has tag:edit permission", () => {
      vi.mocked(useHasPermission).mockReturnValue(true);
      renderPopover();

      expect(
        screen.getByRole("button", { name: /manage tags/i }),
      ).toBeInTheDocument();
    });

    it("hides the edit button when user lacks tag:edit permission", () => {
      vi.mocked(useHasPermission).mockReturnValue(false);
      renderPopover();

      expect(
        screen.queryByRole("button", { name: /manage tags/i }),
      ).not.toBeInTheDocument();
    });

    it("uses a custom editLabel when provided", () => {
      renderPopover({ editLabel: "Manage container tags" });

      expect(
        screen.getByRole("button", { name: /manage container tags/i }),
      ).toBeInTheDocument();
    });
  });

  describe("popover open", () => {
    it("opens the popover when the edit button is clicked", async () => {
      renderPopover();

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(
        screen.getByRole("dialog", { name: /manage tags/i }),
      ).toBeInTheDocument();
    });

    it("shows the search input when the popover is open", async () => {
      renderPopover();

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(
        screen.getByRole("textbox", { name: /search or create tag/i }),
      ).toBeInTheDocument();
    });

    it("shows current tags with remove buttons in the popover", async () => {
      renderPopover({ tags: ["alpha", "beta"] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const dialog = screen.getByRole("dialog", { name: /manage tags/i });
      expect(
        within(dialog).getByRole("button", { name: /remove tag alpha/i }),
      ).toBeInTheDocument();
      expect(
        within(dialog).getByRole("button", { name: /remove tag beta/i }),
      ).toBeInTheDocument();
    });

    it("closes the popover on Escape key", async () => {
      renderPopover();

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(
        screen.getByRole("dialog", { name: /manage tags/i }),
      ).toBeInTheDocument();

      fireEvent.keyDown(document, { key: "Escape" });

      await waitFor(() => {
        expect(
          screen.queryByRole("dialog", { name: /manage tags/i }),
        ).not.toBeInTheDocument();
      });
    });
  });

  describe("max tags reached", () => {
    it("shows 'Max 3 tags' message when there are already 3 tags", async () => {
      renderPopover({ tags: ["alpha", "beta", "gamma"] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(screen.getByText(/max 3 tags/i)).toBeInTheDocument();
    });

    it("does not show the text input when there are already 3 tags", async () => {
      renderPopover({ tags: ["alpha", "beta", "gamma"] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      expect(
        screen.queryByRole("textbox", { name: /search or create tag/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("adding a tag via suggestion", () => {
    it("calls addTag when a suggestion is clicked", async () => {
      mockAddTag.mockResolvedValue(undefined);
      vi.mocked(useTags).mockReturnValue({
        tags: [{ name: "production" }],
        totalCount: 1,
        isLoading: false,
        error: null,
      } as never);

      renderPopover({ tags: [] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const input = screen.getByRole("textbox", {
        name: /search or create tag/i,
      });
      await userEvent.type(input, "prod");

      await userEvent.click(
        screen.getByRole("button", { name: /^production$/i }),
      );

      await waitFor(() => {
        expect(mockAddTag).toHaveBeenCalledWith({
          path: { uid: "entity-1", name: "production" },
        });
      });
    });
  });

  describe("client-side validation", () => {
    it("does not call addTag when tag has invalid characters", async () => {
      renderPopover({ tags: [] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const input = screen.getByRole("textbox", {
        name: /search or create tag/i,
      });
      await userEvent.type(input, "bad-tag{Enter}");

      expect(mockAddTag).not.toHaveBeenCalled();
    });

    it("does not call addTag when tag is too short", async () => {
      renderPopover({ tags: [] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const input = screen.getByRole("textbox", {
        name: /search or create tag/i,
      });
      await userEvent.type(input, "ab{Enter}");

      expect(mockAddTag).not.toHaveBeenCalled();
    });
  });

  describe("removing a tag", () => {
    it("calls removeTag when remove button is clicked", async () => {
      mockRemoveTag.mockResolvedValue(undefined);

      renderPopover({ tags: ["alpha"] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const dialog = screen.getByRole("dialog", { name: /manage tags/i });
      await userEvent.click(
        within(dialog).getByRole("button", { name: /remove tag alpha/i }),
      );

      await waitFor(() => {
        expect(mockRemoveTag).toHaveBeenCalledWith({
          path: { uid: "entity-1", name: "alpha" },
        });
      });
    });
  });

  describe("error states", () => {
    it("shows an error alert when addTag fails", async () => {
      mockAddTag.mockRejectedValue(new Error("network error"));
      vi.mocked(useTags).mockReturnValue({
        tags: [{ name: "production" }],
        totalCount: 1,
        isLoading: false,
        error: null,
      } as never);

      renderPopover({ tags: [] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const input = screen.getByRole("textbox", {
        name: /search or create tag/i,
      });
      await userEvent.type(input, "prod");

      await userEvent.click(
        screen.getByRole("button", { name: /^production$/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });

    it("shows permission error when addTag fails with 403", async () => {
      mockAddTag.mockRejectedValue({ status: 403 });
      vi.mocked(useTags).mockReturnValue({
        tags: [{ name: "production" }],
        totalCount: 1,
        isLoading: false,
        error: null,
      } as never);

      renderPopover({ tags: [] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const input = screen.getByRole("textbox", {
        name: /search or create tag/i,
      });
      await userEvent.type(input, "prod");

      await userEvent.click(
        screen.getByRole("button", { name: /^production$/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /you don't have permission to add tags/i,
        );
      });
    });

    it("shows permission error when removeTag fails with 403", async () => {
      mockRemoveTag.mockRejectedValue({ status: 403 });

      renderPopover({ tags: ["alpha"] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const dialog = screen.getByRole("dialog", { name: /manage tags/i });
      await userEvent.click(
        within(dialog).getByRole("button", { name: /remove tag alpha/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /you don't have permission to remove tags/i,
        );
      });
    });

    it("shows generic error when removeTag fails with non-403 status", async () => {
      mockRemoveTag.mockRejectedValue(new Error("server error"));

      renderPopover({ tags: ["alpha"] });

      await userEvent.click(
        screen.getByRole("button", { name: /manage tags/i }),
      );

      const dialog = screen.getByRole("dialog", { name: /manage tags/i });
      await userEvent.click(
        within(dialog).getByRole("button", { name: /remove tag alpha/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          /failed to remove "alpha"/i,
        );
      });
    });
  });
});
