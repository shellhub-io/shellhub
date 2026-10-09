import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, setTags } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import ManageTagsModal from "@/components/ManageTagsModal";

async function confirmDelete(name: string) {
  const user = userEvent.setup();
  const onTagDeleted = vi.fn();

  render(
    <ManageTagsModal open onClose={vi.fn()} onTagDeleted={onTagDeleted} />,
    {
      wrapper: createTestWrapper(),
    },
  );

  await screen.findByText(name);
  await user.click(screen.getByTitle("Delete"));

  const dialog = await screen.findByRole("dialog", { name: "Delete tag" });
  await user.click(within(dialog).getByRole("button", { name: "Delete tag" }));

  return { dialog, onTagDeleted };
}

describe("ManageTagsModal", () => {
  beforeEach(() => {
    seedAuthStore();
    setTags(["production"]);
  });

  it("names what holds a tag the server refuses to delete", async () => {
    server.use(
      http.delete("*/api/tags/:name", () =>
        HttpResponse.json(
          {
            message: "tag in use",
            fields: {
              name: 'public key "deploy", access policy "operators"',
            },
          },
          { status: 409 },
        ),
      ),
    );

    const { dialog, onTagDeleted } = await confirmDelete("production");

    expect(await within(dialog).findByRole("alert")).toHaveTextContent(
      '"production" is used in the device filter of public key "deploy", access policy "operators". Remove it from there before deleting the tag.',
    );
    expect(onTagDeleted).not.toHaveBeenCalled();
  });

  it("says the delete failed when the server gives no reason", async () => {
    server.use(
      http.delete(
        "*/api/tags/:name",
        () => new HttpResponse(null, { status: 500 }),
      ),
    );

    const { dialog } = await confirmDelete("production");

    expect(await within(dialog).findByRole("alert")).toHaveTextContent(
      'Failed to delete "production".',
    );
  });

  it("closes the dialog once the tag is deleted", async () => {
    server.use(
      http.delete(
        "*/api/tags/:name",
        () => new HttpResponse(null, { status: 204 }),
      ),
    );

    const { onTagDeleted } = await confirmDelete("production");

    await vi.waitFor(() =>
      expect(onTagDeleted).toHaveBeenCalledWith("production"),
    );
    expect(
      screen.queryByRole("dialog", { name: "Delete tag" }),
    ).not.toBeInTheDocument();
  });
});
