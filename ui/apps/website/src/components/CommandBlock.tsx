import { WindowChrome } from "@shellhub/design-system/primitives";
import { CopyButton } from "@shellhub/design-system/components";

/**
 * Props of CommandBlock. command is rendered verbatim after the prompt and is what the copy
 * button puts on the clipboard, so it must be the line a reader can paste unchanged.
 */
export interface CommandBlockProps {
  command: string;
  className?: string;
}

/**
 * A terminal-chrome window showing a single copyable shell command, with the
 * copy button docked in the title bar. Used across marketing pages (Quick
 * Start, getting-started) to present the one-line install command.
 */
export function CommandBlock({ command, className }: CommandBlockProps) {
  return (
    <WindowChrome
      variant="terminal"
      className={className}
      titleBarSlot={<CopyButton text={command} />}
    >
      <code className="text-sm text-accent-cyan">
        <span className="text-text-muted">$ </span>
        {command}
      </code>
    </WindowChrome>
  );
}
