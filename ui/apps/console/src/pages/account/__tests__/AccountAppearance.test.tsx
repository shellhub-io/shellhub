import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useThemeStore } from "@/stores/themeStore";
import { useSidebarStore } from "@/stores/sidebarStore";
import { useSessionPlayerStore } from "@/stores/sessionPlayerStore";
import {
  TERMINAL_FONTS,
  useTerminalThemeStore,
  type TerminalTheme,
} from "@/stores/terminalThemeStore";
import AccountAppearance from "../AccountAppearance";

function terminalTheme(name: string, dark: boolean): TerminalTheme {
  const background = dark ? "#101010" : "#f8f8f8";
  const foreground = dark ? "#e0e0e0" : "#202020";
  return {
    name,
    dark,
    preview: { background, foreground },
    colors: { background, foreground },
  };
}

function renderAppearance() {
  const user = userEvent.setup();
  render(<AccountAppearance />);
  return user;
}

const THEMES = [terminalTheme("Night", true), terminalTheme("Paper", false)];

beforeEach(() => {
  localStorage.clear();
  useTerminalThemeStore.setState({
    themes: THEMES,
    themeName: "Night",
    theme: THEMES[0],
  });
  useThemeStore.getState().setPreference("system");
  useSidebarStore.getState().setPin("auto");
  useSessionPlayerStore.getState().setControls("auto");
});

describe("AccountAppearance", () => {
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

  it("pins the sidebar, the same choice the layout's pin button makes", async () => {
    const user = renderAppearance();

    await user.click(screen.getByRole("radio", { name: "Folded" }));

    expect(useSidebarStore.getState().pin).toBe("rail");
  });

  it("sets how the session player shows its controls", async () => {
    const user = renderAppearance();

    await user.click(screen.getByRole("radio", { name: "Always" }));

    expect(useSessionPlayerStore.getState().controls).toBe("always");
  });

  it("picks the terminal theme", async () => {
    const user = renderAppearance();

    await user.click(screen.getByRole("button", { name: /Paper/ }));

    expect(useTerminalThemeStore.getState().themeName).toBe("Paper");
  });

  it("shows a theme in the preview while the pointer is on it, without choosing it", async () => {
    const user = renderAppearance();
    const preview = screen.getByLabelText("Terminal preview");

    await user.hover(screen.getByRole("button", { name: /Paper/ }));
    expect(preview).toHaveStyle({ color: "#202020" });
    expect(useTerminalThemeStore.getState().themeName).toBe("Night");

    await user.unhover(screen.getByRole("button", { name: /Paper/ }));
    expect(preview).toHaveStyle({ color: "#e0e0e0" });
  });

  it("keeps previewing the focused theme after the pointer passes over another", async () => {
    const user = renderAppearance();
    const preview = screen.getByLabelText("Terminal preview");
    const paper = screen.getByRole("button", { name: /Paper/ });
    const night = screen.getByRole("button", { name: /Night/ });

    paper.focus();
    await user.hover(night);
    await user.unhover(night);

    expect(preview).toHaveStyle({ color: "#202020" });
  });

  it("sets the terminal font size and shows it in the preview", async () => {
    useTerminalThemeStore.getState().setFontSize(14);
    const user = renderAppearance();

    await user.click(
      screen.getByRole("button", { name: "Increase font size" }),
    );

    expect(useTerminalThemeStore.getState().fontSize).toBe(15);
    expect(screen.getByLabelText("Terminal preview")).toHaveStyle({
      fontSize: "15px",
    });
  });

  it("sets the terminal font family", async () => {
    const user = renderAppearance();
    const other = TERMINAL_FONTS.find(
      (f) => f !== useTerminalThemeStore.getState().fontFamily,
    )!;

    await user.click(screen.getByRole("button", { name: /^Font family:/ }));
    await user.click(screen.getByRole("menuitemradio", { name: other }));

    expect(useTerminalThemeStore.getState().fontFamily).toBe(other);
  });

  it("draws the player's controls in the preview unless they are hidden", async () => {
    const user = renderAppearance();
    expect(screen.getByText("01:24 / 04:10")).toBeInTheDocument();

    await user.click(screen.getByRole("radio", { name: "Hidden" }));

    expect(screen.queryByText("01:24 / 04:10")).not.toBeInTheDocument();
  });
});
