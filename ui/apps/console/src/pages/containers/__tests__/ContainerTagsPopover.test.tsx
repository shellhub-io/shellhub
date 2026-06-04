import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { fireEvent } from "@testing-library/react";
import type { NormalizedContainer } from "@/hooks/useContainers";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useContainerMutations", () => ({
  useAddContainerTag: vi.fn(),
  useRemoveContainerTag: vi.fn(),
}));

vi.mock("@/hooks/useTags", () => ({
  useTags: vi.fn(),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
}));

// useEscapeKey attaches a keydown listener to document; we keep the real impl.
// The popover uses createPortal — RTL queries document.body by default, so
// portal content is reachable without any mock.

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import {
  useAddContainerTag,
  useRemoveContainerTag,
} from "@/hooks/useContainerMutations";
import { useTags } from "@/hooks/useTags";
import { useHasPermission } from "@/hooks/useHasPermission";
import ContainerTagsPopover from "../ContainerTagsPopover";

// ── Helpers ───────────────────────────────────────────────────────────────────

const mockAddMutateAsync = vi.fn();
const mockRemoveMutateAsync = vi.fn();

function makeContainer(
  overrides: Partial<NormalizedContainer> = {},
): NormalizedContainer {
  return {
    uid: "container-1",
    name: "my-container",
    tags: [],
    tenant_id: "tenant-abc",
    status: "accepted",
    online: true,
    namespace: "dev",
    created_at: "2024-01-01T00:00:00Z",
    last_seen: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useAddContainerTag).mockReturnValue({
    mutateAsync: mockAddMutateAsync,
  } as never);
  vi.mocked(useRemoveContainerTag).mockReturnValue({
    mutateAsync: mockRemoveMutateAsync,
  } as never);
  vi.mocked(useTags).mockReturnValue({
    tags: [],
    totalCount: 0,
    isLoading: false,
    error: null,
  });
  vi.mocked(useHasPermission).mockReturnValue(true);
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("ContainerTagsPopover", () => {
  describe("tag chip rendering", () => {
    it("renders each tag as a clickable button", () => {
      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha", "beta"] })}
          onFilterTag={vi.fn()}
        />,
      );
      // Chip buttons have their tag name as accessible text content.
      // The title attr ("Filter by \"alpha\"") is not the accessible name.
      expect(
        screen.getByRole("button", { name: /^alpha$/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /^beta$/i }),
      ).toBeInTheDocument();
    });

    it("calls onFilterTag with the tag name when a chip is clicked", async () => {
      const onFilterTag = vi.fn();
      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha"] })}
          onFilterTag={onFilterTag}
        />,
      );

      await userEvent.click(screen.getByRole("button", { name: /^alpha$/i }));

      expect(onFilterTag).toHaveBeenCalledWith("alpha");
    });

    it("shows 'No tags' text when container has no tags", () => {
      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: [] })}
          onFilterTag={vi.fn()}
        />,
      );
      expect(screen.getByText("No tags")).toBeInTheDocument();
    });
  });

  describe("edit button visibility", () => {
    it("shows the edit button when user has tag:edit permission", () => {
      vi.mocked(useHasPermission).mockReturnValue(true);
      render(
        <ContainerTagsPopover
          container={makeContainer()}
          onFilterTag={vi.fn()}
        />,
      );
      expect(
        screen.getByRole("button", { name: /manage container tags/i }),
      ).toBeInTheDocument();
    });

    it("hides the edit button when user lacks tag:edit permission", () => {
      vi.mocked(useHasPermission).mockReturnValue(false);
      render(
        <ContainerTagsPopover
          container={makeContainer()}
          onFilterTag={vi.fn()}
        />,
      );
      expect(
        screen.queryByRole("button", { name: /manage container tags/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("popover open", () => {
    it("opens the popover when the edit button is clicked", async () => {
      render(
        <ContainerTagsPopover
          container={makeContainer()}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
      );

      expect(
        screen.getByRole("dialog", { name: /manage tags/i }),
      ).toBeInTheDocument();
    });

    it("shows the search input when the popover is open", async () => {
      render(
        <ContainerTagsPopover
          container={makeContainer()}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
      );

      expect(
        screen.getByRole("textbox", { name: /search or create tag/i }),
      ).toBeInTheDocument();
    });

    it("shows current tags with remove buttons in the popover", async () => {
      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha", "beta"] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
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
      render(
        <ContainerTagsPopover
          container={makeContainer()}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
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
    it("shows 'Max 3 tags' message when container already has 3 tags", async () => {
      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha", "beta", "gamma"] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
      );

      expect(screen.getByText(/max 3 tags/i)).toBeInTheDocument();
    });

    it("does not show the text input when container already has 3 tags", async () => {
      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha", "beta", "gamma"] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
      );

      expect(
        screen.queryByRole("textbox", { name: /search or create tag/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("adding a tag via suggestion", () => {
    it("calls addTag mutation when a suggestion is clicked", async () => {
      mockAddMutateAsync.mockResolvedValue(undefined);
      vi.mocked(useTags).mockReturnValue({
        tags: [{ name: "production" }],
        totalCount: 1,
        isLoading: false,
        error: null,
      } as never);

      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: [] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
      );

      const input = screen.getByRole("textbox", {
        name: /search or create tag/i,
      });
      await userEvent.type(input, "prod");

      const suggestionBtn = screen.getByRole("button", {
        name: /^production$/i,
      });
      await userEvent.click(suggestionBtn);

      await waitFor(() => {
        expect(mockAddMutateAsync).toHaveBeenCalledWith({
          path: { uid: "container-1", name: "production" },
        });
      });
    });
  });

  describe("removing a tag", () => {
    it("calls removeTag mutation when remove button is clicked", async () => {
      mockRemoveMutateAsync.mockResolvedValue(undefined);

      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha"] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
      );

      const dialog = screen.getByRole("dialog", { name: /manage tags/i });
      await userEvent.click(
        within(dialog).getByRole("button", { name: /remove tag alpha/i }),
      );

      await waitFor(() => {
        expect(mockRemoveMutateAsync).toHaveBeenCalledWith({
          path: { uid: "container-1", name: "alpha" },
        });
      });
    });
  });

  describe("error states", () => {
    it("shows an error alert when addTag mutation fails", async () => {
      mockAddMutateAsync.mockRejectedValue(new Error("network error"));
      vi.mocked(useTags).mockReturnValue({
        tags: [{ name: "production" }],
        totalCount: 1,
        isLoading: false,
        error: null,
      } as never);

      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: [] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
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

    it("shows permission error (403) when addTag fails with 403", async () => {
      mockAddMutateAsync.mockRejectedValue({ status: 403 });
      vi.mocked(useTags).mockReturnValue({
        tags: [{ name: "production" }],
        totalCount: 1,
        isLoading: false,
        error: null,
      } as never);

      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: [] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
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

    it("shows permission error (403) when removeTag fails with 403", async () => {
      mockRemoveMutateAsync.mockRejectedValue({ status: 403 });

      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha"] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
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
      mockRemoveMutateAsync.mockRejectedValue(new Error("server error"));

      render(
        <ContainerTagsPopover
          container={makeContainer({ tags: ["alpha"] })}
          onFilterTag={vi.fn()}
        />,
      );

      await userEvent.click(
        screen.getByRole("button", { name: /manage container tags/i }),
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
