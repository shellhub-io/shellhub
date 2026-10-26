import { WindowChrome } from "@shellhub/design-system/primitives";
import { CopyButton } from "@shellhub/design-system/components";

/**
 * Props of CommandBlock. command is what the copy button puts on the clipboard, so it must be
 * what a reader can paste unchanged. Each line of it is a command of its own and gets its own
 * prompt.
 */
export interface CommandBlockProps {
  command: string;
  className?: string;
}

/**
 * A terminal-chrome window showing copyable shell commands, one per line, with the copy button
 * docked in the title bar. Used across marketing pages (Quick Start, getting-started) to present
 * the install commands.
 */
export function CommandBlock({ command, className }: CommandBlockProps) {
  return (
    <WindowChrome
      variant="terminal"
      className={className}
      titleBarSlot={<CopyButton text={command} />}
    >
      <code className="flex flex-col gap-1 text-sm text-accent-cyan">
        {command.split("\n").map((line, i) => (
          <span key={i}>
            <span className="text-text-muted">$ </span>
            {line}
          </span>
        ))}
      </code>
    </WindowChrome>
  );
}
