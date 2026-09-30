import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useSessionPlayerStore } from "@/stores/sessionPlayerStore";
import {
  TERMINAL_FONTS,
  useTerminalThemeStore,
  type TerminalTheme,
} from "@/stores/terminalThemeStore";
import TerminalPreferences from "../TerminalPreferences";

function terminalTheme(name: string, dark: boolean): TerminalTheme {
  const background = dark ? "#101010" : "#f8f8f8";
  const foreground = dark ? "#e0e0e0" : "#202020";
  return { name, dark, colors: { background, foreground } };
}

function renderTerminal() {
  const user = userEvent.setup();
  render(<TerminalPreferences />);
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
  useSessionPlayerStore.getState().setControls("auto");
});

describe("TerminalPreferences", () => {
  it("sets how the session player shows its controls", async () => {
    const user = renderTerminal();

    await user.click(screen.getByRole("radio", { name: "Always" }));

    expect(useSessionPlayerStore.getState().controls).toBe("always");
  });

  it("picks the terminal theme", async () => {
    const user = renderTerminal();

    await user.click(screen.getByRole("button", { name: /Paper/ }));

    expect(useTerminalThemeStore.getState().themeName).toBe("Paper");
  });

  it("sets the terminal font size and shows it in the preview", async () => {
    useTerminalThemeStore.getState().setFontSize(14);
    const user = renderTerminal();

    await user.click(
      screen.getByRole("button", { name: "Increase font size" }),
    );

    expect(useTerminalThemeStore.getState().fontSize).toBe(15);
    expect(screen.getByLabelText("Terminal preview")).toHaveStyle({
      fontSize: "15px",
    });
  });

  it("sets the terminal font family", async () => {
    const user = renderTerminal();
    const other = TERMINAL_FONTS.find(
      (f) => f !== useTerminalThemeStore.getState().fontFamily,
    )!;

    await user.click(screen.getByRole("button", { name: /^Font family:/ }));
    await user.click(screen.getByRole("menuitemradio", { name: other }));

    expect(useTerminalThemeStore.getState().fontFamily).toBe(other);
  });

  it("draws the player's controls in the preview unless they are hidden", async () => {
    const user = renderTerminal();
    expect(screen.getByText("01:24 / 04:10")).toBeInTheDocument();

    await user.click(screen.getByRole("radio", { name: "Hidden" }));

    expect(screen.queryByText("01:24 / 04:10")).not.toBeInTheDocument();
  });
});
