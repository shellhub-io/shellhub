import { ITerminalTheme } from "@/interfaces/ITerminal";

export const mockTerminalThemes: ITerminalTheme[] = [
  {
    name: "ShellHub Dark",
    description: "Dark theme",
    colors: {
      background: "#0f1526",
      foreground: "#ffffff",
      cursor: "#ffffff",
      selection: "#264f78",
    },
  },
  {
    name: "ShellHub Light",
    description: "Light theme",
    colors: {
      background: "#ffffff",
      foreground: "#000000",
      cursor: "#000000",
      selection: "#add6ff",
    },
  },
  {
    name: "Dracula",
    description: "Dark theme",
    colors: {
      background: "#282a36",
      foreground: "#f8f8f2",
      cursor: "#f8f8f0",
      selection: "#44475a",
    },
  },
];
