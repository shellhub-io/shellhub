import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { SdkHttpError } from "@/api/errors";
import RenameableName from "../RenameableName";

function sdkError(status: number): Error & SdkHttpError {
  return Object.assign(new Error("Request failed"), {
    status,
    headers: new Headers(),
  });
}

function renderName(
  rename = vi.fn().mockResolvedValue(undefined),
  canRename = true,
) {
  render(
    <RenameableName
      uid="uid-1"
      name="build-01"
      entityLabel="device"
      rename={rename}
      canRename={canRename}
      nsName="dev"
    />,
  );
  return { rename };
}

async function openRename() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "Rename device" }));
  const field = await screen.findByLabelText("Device name");
  return { user, field };
}

describe("RenameableName", () => {
  it("offers no rename without the permission", () => {
    renderName(undefined, false);

    expect(screen.getByText("build-01")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Rename device" }),
    ).not.toBeInTheDocument();
  });

  it("previews the SSHID the new name gives", async () => {
    renderName();
    const { user, field } = await openRename();

    await user.clear(field);
    await user.type(field, "web-02");

    expect(
      screen.getByText("dev.web-02@localhost"),
    ).toBeInTheDocument();
  });

  it("closes without a request when the name is unchanged", async () => {
    const { rename } = renderName();
    const { user } = await openRename();

    await user.keyboard("{Enter}");

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
    expect(rename).not.toHaveBeenCalled();
  });

  it("renames the device by its uid", async () => {
    const { rename } = renderName();
    const { user, field } = await openRename();

    await user.clear(field);
    await user.type(field, "  web-02  {Enter}");

    expect(rename).toHaveBeenCalledWith({
      path: { uid: "uid-1" },
      body: { name: "web-02" },
    });
  });

  it("keeps the dialog open when the name is taken", async () => {
    renderName(vi.fn().mockRejectedValue(sdkError(409)));
    const { user, field } = await openRename();

    await user.clear(field);
    await user.type(field, "web-02{Enter}");

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "A device with that name already exists.",
    );
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
