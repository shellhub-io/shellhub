import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import RenameSection, { type RenameSectionProps } from "../RenameSection";

function makeSdkError(status: number) {
  return Object.assign(new Error("Request failed"), {
    status,
    headers: new Headers(),
  });
}

const mockRename = vi
  .fn<RenameSectionProps["rename"]>()
  .mockResolvedValue(undefined);

const defaultProps: RenameSectionProps = {
  uid: "abc-123",
  currentName: "my-device",
  rename: mockRename,
  entityLabel: "device",
};

function renderAndEdit(overrides: Partial<RenameSectionProps> = {}) {
  const user = userEvent.setup();
  render(<RenameSection {...defaultProps} {...overrides} />);
  return { user };
}

async function enterEditMode(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: /rename device/i }));
}

async function typeAndSave(
  user: ReturnType<typeof userEvent.setup>,
  name: string,
) {
  await enterEditMode(user);
  const input = screen.getByRole("textbox");
  await user.clear(input);
  await user.type(input, name);
  await user.click(
    screen.getByRole("button", { name: /save device name/i }),
  );
}

beforeEach(() => {
  mockRename.mockReset().mockResolvedValue(undefined);
});

describe("RenameSection", () => {
  describe("display mode", () => {
    it("renders the current name as a heading", () => {
      renderAndEdit();
      expect(
        screen.getByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });

    it("shows the rename button when canRename is true", () => {
      renderAndEdit();
      expect(
        screen.getByRole("button", { name: /rename device/i }),
      ).toBeInTheDocument();
    });

    it("hides the rename button when canRename is false", () => {
      renderAndEdit({ canRename: false });
      expect(
        screen.queryByRole("button", { name: /rename device/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("edit mode", () => {
    it("shows an input pre-filled with the current name", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);

      expect(screen.getByRole("textbox")).toHaveValue("my-device");
    });

    it("shows save and cancel buttons", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);

      expect(
        screen.getByRole("button", { name: /save device name/i }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: /cancel rename/i }),
      ).toBeInTheDocument();
    });

    it("cancels editing when the cancel button is clicked", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);

      await user.click(screen.getByRole("button", { name: /cancel rename/i }));

      expect(
        screen.getByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });

    it("cancels editing on Escape", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);

      await user.click(screen.getByRole("textbox"));
      await user.keyboard("{Escape}");

      expect(
        screen.getByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });
  });

  describe("saving", () => {
    it("calls rename with the trimmed name and exits edit mode", async () => {
      const { user } = renderAndEdit();
      await typeAndSave(user, "  new-name  ");

      expect(mockRename).toHaveBeenCalledWith({
        path: { uid: "abc-123" },
        body: { name: "new-name" },
      });
      expect(
        screen.getByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });

    it("saves on Enter", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);
      const input = screen.getByRole("textbox");
      await user.clear(input);
      await user.type(input, "new-name{Enter}");

      expect(mockRename).toHaveBeenCalledWith({
        path: { uid: "abc-123" },
        body: { name: "new-name" },
      });
    });

    it("exits without calling rename when name is unchanged", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);
      await user.click(
        screen.getByRole("button", { name: /save device name/i }),
      );

      expect(mockRename).not.toHaveBeenCalled();
      expect(
        screen.getByRole("heading", { name: "my-device" }),
      ).toBeInTheDocument();
    });

    it("exits without calling rename when name is empty", async () => {
      const { user } = renderAndEdit();
      await enterEditMode(user);
      await user.clear(screen.getByRole("textbox"));
      await user.click(
        screen.getByRole("button", { name: /save device name/i }),
      );

      expect(mockRename).not.toHaveBeenCalled();
    });
  });

  describe("error handling", () => {
    it.each([
      { status: 400, message: /invalid device name/i },
      { status: 409, message: /already exists/i },
      { status: 500, message: /failed to rename device/i },
      { status: undefined, message: /failed to rename device/i },
    ])(
      "shows the right error for status $status",
      async ({ status, message }) => {
        mockRename.mockRejectedValue(
          status ? makeSdkError(status) : new Error("network error"),
        );
        const { user } = renderAndEdit();
        await typeAndSave(user, "taken-name");

        expect(screen.getByRole("alert")).toHaveTextContent(message);
      },
    );

    it("uses entityLabel in error messages", async () => {
      mockRename.mockRejectedValue(makeSdkError(409));
      const user = userEvent.setup();
      render(
        <RenameSection {...defaultProps} entityLabel="container" />,
      );
      await user.click(
        screen.getByRole("button", { name: /rename container/i }),
      );

      const input = screen.getByRole("textbox");
      await user.clear(input);
      await user.type(input, "taken-name");
      await user.click(
        screen.getByRole("button", { name: /save container name/i }),
      );

      expect(screen.getByRole("alert")).toHaveTextContent(
        /a container with that name already exists/i,
      );
    });
  });
});
