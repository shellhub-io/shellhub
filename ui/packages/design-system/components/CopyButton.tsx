import { useEffect, useRef, useState } from "react";
import { ClipboardIcon, CheckIcon } from "@heroicons/react/24/outline";
import { Button } from "../primitives/Button";
import { IconButton } from "../primitives/IconButton";
import { cn } from "../primitives/cn";

const SIZE_CLASS: Record<"sm" | "md", string> = {
  sm: "w-3.5 h-3.5",
  md: "w-4 h-4",
};

/**
 * Props of CopyButton. onError is called when the clipboard is unreachable — an insecure
 * context, or a write the browser rejected — because the button reports nothing itself.
 */
export interface CopyButtonProps {
  text: string;
  size?: "sm" | "md";
  showLabel?: boolean;
  className?: string;
  onError?: () => void;
}

/**
 * Copies text to the clipboard and confirms with a check mark for 1.5 seconds.
 * Outside a secure context navigator.clipboard does not exist, so the write is not
 * attempted at all and onError is called instead. The click is stopped from propagating,
 * which is what lets the button sit inside a row that is itself clickable.
 */
export function CopyButton({
  text,
  size = "sm",
  showLabel = false,
  className = "",
  onError,
}: CopyButtonProps) {
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    },
    [],
  );

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation();

    if (!globalThis.isSecureContext) {
      onError?.();
      return;
    }

    navigator.clipboard.writeText(text).then(
      () => {
        setCopied(true);
        if (timerRef.current !== null) {
          clearTimeout(timerRef.current);
        }
        timerRef.current = setTimeout(() => {
          setCopied(false);
          timerRef.current = null;
        }, 1500);
      },
      () => {
        onError?.();
      },
    );
  };

  const labelIcon = copied ? (
    <CheckIcon className="w-3 h-3 text-accent-green" strokeWidth={2} />
  ) : (
    <ClipboardIcon className="w-3 h-3" />
  );
  const iconButtonIcon = copied ? (
    <CheckIcon
      className={cn(SIZE_CLASS[size], "text-accent-green")}
      strokeWidth={2}
    />
  ) : (
    <ClipboardIcon className={SIZE_CLASS[size]} />
  );

  if (showLabel) {
    return (
      <Button
        variant={copied ? "successSoft" : "secondary"}
        size="sm"
        icon={labelIcon}
        className={cn("shrink-0", className)}
        onClick={handleCopy}
      >
        {copied ? "Copied" : "Copy"}
      </Button>
    );
  }

  return (
    <IconButton
      size={size}
      title="Copy"
      className={className}
      onClick={handleCopy}
    >
      {iconButtonIcon}
    </IconButton>
  );
}
