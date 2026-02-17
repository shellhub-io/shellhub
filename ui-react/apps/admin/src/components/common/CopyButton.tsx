import { useState } from "react";
import { CheckIcon, DocumentDuplicateIcon } from "@heroicons/react/24/outline";

const sizes = {
  sm: { button: "p-1 rounded", icon: "w-3.5 h-3.5" },
  md: { button: "p-1.5 rounded-md", icon: "w-4 h-4" },
};

export default function CopyButton({
  text,
  size = "sm",
  showLabel = false,
  className = "",
}: {
  text: string;
  size?: "sm" | "md";
  showLabel?: boolean;
  className?: string;
}) {
  const [copied, setCopied] = useState(false);
  const s = sizes[size];

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  if (showLabel) {
    return (
      <button
        onClick={(e) => {
          e.stopPropagation();
          handleCopy();
        }}
        className={`shrink-0 px-3 py-1.5 rounded-md text-2xs font-semibold transition-all ${
          copied
            ? "bg-accent-green/15 text-accent-green border border-accent-green/25"
            : "bg-hover-medium text-text-muted hover:text-text-primary hover:bg-hover-strong border border-transparent"
        } ${className}`}
      >
        <span className="flex items-center gap-1">
          {copied ? (
            <>
              <CheckIcon className="w-3 h-3" strokeWidth={2.5} />
              Copied
            </>
          ) : (
            <>
              <DocumentDuplicateIcon className="w-3 h-3" />
              Copy
            </>
          )}
        </span>
      </button>
    );
  }

  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        handleCopy();
      }}
      className={`${s.button} text-text-muted hover:text-text-primary hover:bg-hover-medium transition-all ${className}`}
      title="Copy"
    >
      {copied ? (
        <CheckIcon className={`${s.icon} text-accent-green`} strokeWidth={2} />
      ) : (
        <DocumentDuplicateIcon className={s.icon} />
      )}
    </button>
  );
}
