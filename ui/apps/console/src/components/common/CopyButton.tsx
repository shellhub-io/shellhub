import { CopyButton as CopyButtonBase } from "@shellhub/design-system/components";
import { useClipboardWarning } from "@/hooks/useCopy";

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
  const triggerWarning = useClipboardWarning();

  return (
    <CopyButtonBase
      text={text}
      size={size}
      showLabel={showLabel}
      className={className}
      onError={triggerWarning}
    />
  );
}
