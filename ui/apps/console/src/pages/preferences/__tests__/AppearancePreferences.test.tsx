import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useThemeStore } from "@/stores/themeStore";
import { useSidebarStore } from "@/stores/sidebarStore";
import AppearancePreferences from "../AppearancePreferences";

function renderAppearance() {
  const user = userEvent.setup();
  render(<AppearancePreferences />);
  return user;
}

beforeEach(() => {
  localStorage.clear();
  useThemeStore.getState().setPreference("system");
  useSidebarStore.getState().setPin("auto");
});

describe("AppearancePreferences", () => {
  it("switches the console theme", async () => {
    const user = renderAppearance();

    await user.click(screen.getByRole("radio", { name: "Light" }));

    expect(useThemeStore.getState().preference).toBe("light");
  });

  it("names each choice's group by its title", () => {
    renderAppearance();

    expect(
      screen.getByRole("radiogroup", { name: "Theme" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("radiogroup", { name: "Sidebar" }),
    ).toBeInTheDocument();
  });

  it("pins the sidebar", async () => {
    const user = renderAppearance();

    await user.click(screen.getByRole("radio", { name: "Folded" }));

    expect(useSidebarStore.getState().pin).toBe("rail");
  });
});
