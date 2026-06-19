import { CheckIcon, DocumentDuplicateIcon } from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useCopy } from "@/hooks/useCopy";

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
  const { copy, copied } = useCopy();
  const s = sizes[size];

  if (showLabel) {
    return (
      <Button
        size="sm"
        variant={copied ? "successSoft" : "secondary"}
        icon={
          copied ? (
            <CheckIcon className="w-3 h-3" strokeWidth={2.5} />
          ) : (
            <DocumentDuplicateIcon className="w-3 h-3" />
          )
        }
        onClick={(e) => {
          e.stopPropagation();
          copy(text);
        }}
        className={`shrink-0 ${className}`}
      >
        {copied ? "Copied" : "Copy"}
      </Button>
    );
  }

  return (
    <IconButton
      size={size}
      title="Copy"
      onClick={(e) => {
        e.stopPropagation();
        copy(text);
      }}
      className={className}
    >
      {copied ? (
        <CheckIcon className={`${s.icon} text-accent-green`} strokeWidth={2} />
      ) : (
        <DocumentDuplicateIcon className={s.icon} />
      )}
    </IconButton>
  );
}
