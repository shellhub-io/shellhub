import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { fireEvent } from "@testing-library/react";
import { setTags } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
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
    { wrapper: createTestWrapper() },
  );
}

async function openPopover() {
  await userEvent.click(screen.getByRole("button", { name: /manage tags/i }));
  return screen.getByRole("dialog", { name: /manage tags/i });
}

function tagInput() {
  return screen.getByRole("textbox", { name: /search or create tag/i });
}

async function addProductionViaSuggestion() {
  await openPopover();
  await userEvent.type(tagInput(), "prod");
  await userEvent.click(
    await screen.findByRole("button", { name: /^production$/i }),
  );
}

async function removeAlpha() {
  const dialog = await openPopover();
  await userEvent.click(
    within(dialog).getByRole("button", { name: /remove tag alpha/i }),
  );
}

describe("TagsPopover", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setTags([]);
    useAuthStore.setState({ role: "owner" });
  });

  describe("tag chip rendering", () => {
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
    it("hides the edit button when user lacks tag:edit permission", () => {
      useAuthStore.setState({ role: "observer" });
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

  it("closes the popover on Escape key", async () => {
    renderPopover();

    await openPopover();
    fireEvent.keyDown(document, { key: "Escape" });

    await waitFor(() => {
      expect(
        screen.queryByRole("dialog", { name: /manage tags/i }),
      ).not.toBeInTheDocument();
    });
  });

  it("replaces the tag input with a 'Max 3 tags' message at the limit", async () => {
    renderPopover({ tags: ["alpha", "beta", "gamma"] });

    await openPopover();

    expect(screen.getByText(/max 3 tags/i)).toBeInTheDocument();
    expect(
      screen.queryByRole("textbox", { name: /search or create tag/i }),
    ).not.toBeInTheDocument();
  });

  it("calls addTag when a suggestion is clicked", async () => {
    mockAddTag.mockResolvedValue(undefined);
    setTags(["production"]);
    renderPopover({ tags: [] });

    await addProductionViaSuggestion();

    await waitFor(() => {
      expect(mockAddTag).toHaveBeenCalledWith({
        uid: "entity-1",
        name: "production",
      });
    });
  });

  it.each([
    ["invalid characters", "bad-tag"],
    ["too few characters", "ab"],
  ])("does not call addTag for a tag with %s", async (_label, value) => {
    renderPopover({ tags: [] });

    await openPopover();
    await userEvent.type(tagInput(), `${value}{Enter}`);

    expect(mockAddTag).not.toHaveBeenCalled();
  });

  it("calls removeTag when remove button is clicked", async () => {
    mockRemoveTag.mockResolvedValue(undefined);
    renderPopover({ tags: ["alpha"] });

    await removeAlpha();

    await waitFor(() => {
      expect(mockRemoveTag).toHaveBeenCalledWith({
        uid: "entity-1",
        name: "alpha",
      });
    });
  });

  describe("error states", () => {
    it.each([
      ["a network error", new Error("network error"), null],
      ["a 403", { status: 403 }, /you don't have permission to add tags/i],
    ] as const)("addTag rejecting with %s raises an alert", async (
      _label,
      rejection,
      message,
    ) => {
      mockAddTag.mockRejectedValue(rejection);
      setTags(["production"]);
      renderPopover({ tags: [] });

      await addProductionViaSuggestion();

      await waitFor(() => {
        const alert = screen.getByRole("alert");
        if (message) expect(alert).toHaveTextContent(message);
      });
    });

    it.each([
      ["a 403", { status: 403 }, /you don't have permission to remove tags/i],
      [
        "a non-403 status",
        new Error("server error"),
        /failed to remove "alpha"/i,
      ],
    ] as const)("removeTag rejecting with %s reports '%s'", async (
      _label,
      rejection,
      message,
    ) => {
      mockRemoveTag.mockRejectedValue(rejection);
      renderPopover({ tags: ["alpha"] });

      await removeAlpha();

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(message);
      });
    });
  });
});
