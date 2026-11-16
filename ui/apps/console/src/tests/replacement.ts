import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http } from "msw";
import { expect } from "vitest";
import { jsonWithTotal } from "./msw";

/**
 * The handlers under which the console's lookup for an offline device that an accept would
 * replace finds none, among devices or containers.
 */
export const noReplacedDevice = [
  http.get("*/api/devices", () => jsonWithTotal([])),
  http.get("*/api/containers", () => jsonWithTotal([])),
];

/**
 * Clicks the button named name once it is enabled. An accept stays disabled while the console
 * looks up the device it would replace, so a click before that settles does nothing.
 */
export async function clickOnceEnabled(
  name: RegExp | string,
  user = userEvent.setup(),
) {
  const button = await screen.findByRole("button", { name });
  await waitFor(() => expect(button).toBeEnabled());
  await user.click(button);
}
