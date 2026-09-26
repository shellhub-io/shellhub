import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ContextualDialog from "../ContextualDialog";

function renderDialog(onSubmit = vi.fn().mockResolvedValue(undefined)) {
  render(
    <ContextualDialog
      trigger={<button type="button">Rename</button>}
      icon={<svg />}
      title="Rename device"
      description="Changes its SSHID too"
      submitLabel="Save name"
      onSubmit={onSubmit}
    >
      <input aria-label="Name" defaultValue="build-01" />
    </ContextualDialog>,
  );
  return { onSubmit };
}

async function open() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "Rename" }));
  return { user, dialog: await screen.findByRole("dialog") };
}

describe("ContextualDialog", () => {
  it("opens from its trigger, named and described by its header", async () => {
    renderDialog();
    const { dialog } = await open();

    expect(dialog).toHaveAccessibleName("Rename device");
    expect(dialog).toHaveAccessibleDescription("Changes its SSHID too");
  });

  it("puts focus in the field on open", async () => {
    renderDialog();
    await open();

    await waitFor(() => expect(screen.getByLabelText("Name")).toHaveFocus());
  });

  it("submits on Enter and closes, handing focus back to the trigger", async () => {
    const { onSubmit } = renderDialog();
    const { user } = await open();

    await user.type(screen.getByLabelText("Name"), "{Enter}");

    expect(onSubmit).toHaveBeenCalledOnce();
    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
    expect(screen.getByRole("button", { name: "Rename" })).toHaveFocus();
  });

  it("cancels on Escape without submitting", async () => {
    const { onSubmit } = renderDialog();
    const { user } = await open();

    await user.keyboard("{Escape}");

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("stays open and shows the message when the submit fails", async () => {
    renderDialog(vi.fn().mockRejectedValue(new Error("Name already taken.")));
    const { user } = await open();

    await user.click(screen.getByRole("button", { name: "Save name" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Name already taken.",
    );
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("won't dismiss while a submit is pending, so its error still shows", async () => {
    let reject!: (err: Error) => void;
    renderDialog(
      vi.fn(
        () =>
          new Promise<void>((_, rej) => {
            reject = rej;
          }),
      ),
    );
    const { user } = await open();

    await user.click(screen.getByRole("button", { name: "Save name" }));
    await user.keyboard("{Escape}");
    reject(new Error("Name already taken."));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Name already taken.",
    );
  });

  describe("inside a clickable row", () => {
    function renderInRow() {
      const onRowActivate = vi.fn();
      const onSubmit = vi.fn().mockResolvedValue(undefined);
      render(
        <div
          role="row"
          tabIndex={0}
          onClick={onRowActivate}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              onRowActivate();
            }
          }}
        >
          <ContextualDialog
            trigger={<button type="button">Rename</button>}
            icon={<svg />}
            title="Rename device"
            description="Changes its SSHID too"
            submitLabel="Save name"
            onSubmit={onSubmit}
          >
            <input aria-label="Name" defaultValue="web" />
          </ContextualDialog>
        </div>,
      );
      return { onRowActivate, onSubmit };
    }

    it("opens from the keyboard without activating the row", async () => {
      const { onRowActivate } = renderInRow();
      const user = userEvent.setup();

      screen.getByRole("button", { name: "Rename" }).focus();
      await user.keyboard("{Enter}");

      expect(await screen.findByRole("dialog")).toBeInTheDocument();
      expect(onRowActivate).not.toHaveBeenCalled();
    });

    it("keeps typing and Enter inside the dialog", async () => {
      const { onRowActivate, onSubmit } = renderInRow();
      const user = userEvent.setup();
      await user.click(screen.getByRole("button", { name: "Rename" }));
      const field = await screen.findByLabelText("Name");

      await user.type(field, " server{Enter}");

      expect(onSubmit).toHaveBeenCalledOnce();
      expect(onRowActivate).not.toHaveBeenCalled();
    });

    it("still closes on Escape", async () => {
      renderInRow();
      const user = userEvent.setup();
      await user.click(screen.getByRole("button", { name: "Rename" }));
      await screen.findByRole("dialog");

      await user.keyboard("{Escape}");

      await waitFor(() =>
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
      );
    });
  });
});
