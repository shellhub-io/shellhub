import { useEffect, useRef, useState } from "react";
import { Button } from "../primitives/Button";
import { IconButton } from "../primitives/IconButton";
import { cn } from "../primitives/cn";

const SIZE_CLASS: Record<"sm" | "md", string> = {
  sm: "w-3.5 h-3.5",
  md: "w-4 h-4",
};

function ClipboardIcon({ className }: { className: string }) {
  return (
    <svg
      className={className}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={1.5}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9.75a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"
      />
    </svg>
  );
}

function CheckIcon({ className }: { className: string }) {
  return (
    <svg
      className={cn("text-accent-green", className)}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={2}
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="m4.5 12.75 6 6 9-13.5"
      />
    </svg>
  );
}

export interface CopyButtonProps {
  text: string;
  size?: "sm" | "md";
  showLabel?: boolean;
  className?: string;
  onError?: () => void;
}

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
    <CheckIcon className="w-3 h-3" />
  ) : (
    <ClipboardIcon className="w-3 h-3" />
  );
  const iconButtonIcon = copied ? (
    <CheckIcon className={SIZE_CLASS[size]} />
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
