import { useState } from "react";
import { CheckIcon, DocumentDuplicateIcon } from "@heroicons/react/24/outline";

const sizes = {
  sm: { button: "p-1 rounded", icon: "w-3.5 h-3.5" },
  md: { button: "p-1.5 rounded-md", icon: "w-4 h-4" },
};

export default function CopyButton({
  text,
  size = "sm",
  className = "",
}: {
  text: string;
  size?: "sm" | "md";
  className?: string;
}) {
  const [copied, setCopied] = useState(false);
  const s = sizes[size];

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        handleCopy();
      }}
      className={`${s.button} text-text-muted hover:text-text-primary hover:bg-white/[0.05] transition-all ${className}`}
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
