import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

vi.mock("@/api/errors", () => ({
  isSdkError: vi.fn(),
}));

vi.mock("@/hooks/useTags", () => ({
  useTags: vi.fn(),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
}));

vi.mock("@/hooks/useClickOutside", () => ({
  useClickOutside: vi.fn(),
}));

vi.mock("@/hooks/useEscapeKey", () => ({
  useEscapeKey: vi.fn(),
}));

import { isSdkError } from "@/api/errors";
import { useTags } from "@/hooks/useTags";
import { useHasPermission } from "@/hooks/useHasPermission";
import TagsSection from "../TagsSection";

function makeSdkError(status: number) {
  return { status };
}

beforeEach(() => {
  vi.clearAllMocks();

  vi.mocked(useHasPermission).mockReturnValue(true);
  vi.mocked(useTags).mockReturnValue({
    tags: [{ name: "existing" }, { name: "shared" }, { name: "deploy" }],
    totalCount: 3,
    isLoading: false,
    error: null,
  } as never);
  vi.mocked(isSdkError).mockReturnValue(false);
});

function renderTagsSection({
  uid = "test-uid",
  tags = [] as string[],
  addTag = vi.fn().mockResolvedValue(undefined),
  removeTag = vi.fn().mockResolvedValue(undefined),
}: Partial<{
  uid: string;
  tags: string[];
  addTag: (opts: { path: { uid: string; name: string } }) => Promise<unknown>;
  removeTag: (opts: {
    path: { uid: string; name: string };
  }) => Promise<unknown>;
}> = {}) {
  const finalProps = { uid, tags, addTag, removeTag };
  return { addTag, removeTag, ...render(<TagsSection {...finalProps} />) };
}

async function typeAndSubmit(input: string) {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText("New tag"), input);
  await user.keyboard("{Enter}");
  return user;
}

describe("TagsSection", () => {
  describe("rendering", () => {
    it("displays tag pills for each provided tag", () => {
      renderTagsSection({ tags: ["web", "prod"] });
      expect(screen.getByText("web")).toBeInTheDocument();
      expect(screen.getByText("prod")).toBeInTheDocument();
    });

    it("hides the input and shows a message at the 3-tag limit", () => {
      renderTagsSection({ tags: ["aaa", "bbb", "ccc"] });
      expect(
        screen.getByText("Maximum of 3 tags reached."),
      ).toBeInTheDocument();
      expect(screen.queryByLabelText("New tag")).not.toBeInTheDocument();
    });
  });

  describe("permission gating", () => {
    it("hides editing controls when user lacks tag:edit permission", () => {
      vi.mocked(useHasPermission).mockReturnValue(false);
      renderTagsSection({ tags: ["web"] });
      expect(
        screen.queryByRole("button", { name: /remove tag/i }),
      ).not.toBeInTheDocument();
      expect(screen.queryByLabelText("New tag")).not.toBeInTheDocument();
    });

    it("shows remove buttons and input when user has tag:edit permission", () => {
      renderTagsSection({ tags: ["web"] });
      expect(
        screen.getByRole("button", { name: /remove tag web/i }),
      ).toBeInTheDocument();
      expect(screen.getByLabelText("New tag")).toBeInTheDocument();
    });
  });

  describe("adding a tag", () => {
    it("submits the tag on Enter and clears the input", async () => {
      const { addTag } = renderTagsSection();

      await typeAndSubmit("newtag");

      await waitFor(() => {
        expect(addTag).toHaveBeenCalledWith({
          path: { uid: "test-uid", name: "newtag" },
        });
      });
      expect(screen.getByLabelText("New tag")).toHaveValue("");
    });

    it("submits the tag when clicking the add button", async () => {
      const user = userEvent.setup();
      const { addTag } = renderTagsSection();

      await user.type(screen.getByLabelText("New tag"), "newtag");
      await user.click(screen.getByRole("button", { name: "Add tag" }));

      await waitFor(() => {
        expect(addTag).toHaveBeenCalledWith({
          path: { uid: "test-uid", name: "newtag" },
        });
      });
    });

    it.each([
      { input: "ab", message: "Tag must be at least 3 characters." },
      {
        input: "my-tag",
        message: "Tag must contain only letters and numbers.",
      },
    ])('rejects "$input" with validation error', async ({ input, message }) => {
      const { addTag } = renderTagsSection();

      await typeAndSubmit(input);

      expect(screen.getByRole("alert")).toHaveTextContent(message);
      expect(addTag).not.toHaveBeenCalled();
    });

    it("rejects a tag that is already assigned", async () => {
      const { addTag } = renderTagsSection({ tags: ["existing"] });

      await typeAndSubmit("existing");

      expect(screen.getByRole("alert")).toHaveTextContent(
        "This tag is already added.",
      );
      expect(addTag).not.toHaveBeenCalled();
    });

    it.each([
      { status: 403, message: "You don't have permission to add tags." },
      { status: 400, message: /not a valid tag name/ },
    ])("shows API error for status $status", async ({ status, message }) => {
      vi.mocked(isSdkError).mockReturnValue(true);
      renderTagsSection({
        addTag: vi.fn().mockRejectedValue(makeSdkError(status)),
      });

      await typeAndSubmit("newtag");

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(message);
      });
    });

    it("shows generic error on unknown failure", async () => {
      renderTagsSection({
        addTag: vi.fn().mockRejectedValue(new Error("network")),
      });

      await typeAndSubmit("newtag");

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          "Failed to add tag.",
        );
      });
    });
  });

  describe("removing a tag", () => {
    it("calls removeTag with the correct tag name", async () => {
      const user = userEvent.setup();
      const { removeTag } = renderTagsSection({ tags: ["prod"] });

      await user.click(
        screen.getByRole("button", { name: /remove tag prod/i }),
      );

      await waitFor(() => {
        expect(removeTag).toHaveBeenCalledWith({
          path: { uid: "test-uid", name: "prod" },
        });
      });
    });

    it("shows permission error on 403", async () => {
      vi.mocked(isSdkError).mockReturnValue(true);
      const user = userEvent.setup();
      renderTagsSection({
        tags: ["prod"],
        removeTag: vi.fn().mockRejectedValue(makeSdkError(403)),
      });

      await user.click(
        screen.getByRole("button", { name: /remove tag prod/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          "You don't have permission to remove tags.",
        );
      });
    });

    it("shows generic error on unknown failure", async () => {
      const user = userEvent.setup();
      renderTagsSection({
        tags: ["prod"],
        removeTag: vi.fn().mockRejectedValue(new Error("network")),
      });

      await user.click(
        screen.getByRole("button", { name: /remove tag prod/i }),
      );

      await waitFor(() => {
        expect(screen.getByRole("alert")).toHaveTextContent(
          'Failed to remove "prod".',
        );
      });
    });
  });

  describe("search/autocomplete", () => {
    it("shows suggestions matching typed input and hides non-matching ones", async () => {
      const user = userEvent.setup();
      renderTagsSection();

      await user.type(screen.getByLabelText("New tag"), "shar");

      expect(screen.getByText("shared")).toBeInTheDocument();
      expect(screen.queryByText("deploy")).not.toBeInTheDocument();
    });

    it("shows the Create option for a tag that does not exist yet", async () => {
      const user = userEvent.setup();
      renderTagsSection();

      await user.type(screen.getByLabelText("New tag"), "brandnew");

      expect(screen.getByText(/Create/)).toBeInTheDocument();
    });

    it("adds the tag when clicking a suggestion", async () => {
      const user = userEvent.setup();
      const { addTag } = renderTagsSection();

      await user.type(screen.getByLabelText("New tag"), "dep");
      await user.click(screen.getByRole("option", { name: "deploy" }));

      await waitFor(() => {
        expect(addTag).toHaveBeenCalledWith({
          path: { uid: "test-uid", name: "deploy" },
        });
      });
    });

    it("selects a suggestion with ArrowDown + Enter", async () => {
      const user = userEvent.setup();
      const { addTag } = renderTagsSection();
      const input = screen.getByLabelText("New tag");

      await user.type(input, "dep");
      await user.keyboard("{ArrowDown}{Enter}");

      await waitFor(() => {
        expect(addTag).toHaveBeenCalledWith({
          path: { uid: "test-uid", name: "deploy" },
        });
      });
    });

    it("wraps around when arrowing past the last option", async () => {
      const user = userEvent.setup();
      renderTagsSection();
      const input = screen.getByLabelText("New tag");

      await user.type(input, "e");

      const optionCount = screen.getAllByRole("option").length;
      for (let i = 0; i < optionCount; i++) {
        await user.keyboard("{ArrowDown}");
      }
      // One more wraps back to first
      await user.keyboard("{ArrowDown}");

      const firstOption = screen.getAllByRole("option")[0];
      expect(firstOption).toHaveAttribute("aria-selected", "true");
    });
  });

  describe("accessibility", () => {
    it("renders error messages with role=alert", async () => {
      renderTagsSection();
      await typeAndSubmit("ab");
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("provides descriptive aria-labels on remove buttons", () => {
      renderTagsSection({ tags: ["web", "prod"] });
      expect(
        screen.getByRole("button", { name: "Remove tag web" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Remove tag prod" }),
      ).toBeInTheDocument();
    });
  });
});
