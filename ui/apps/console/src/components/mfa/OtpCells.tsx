import { cn } from "@shellhub/design-system/cn";
import { LABEL } from "@/utils/styles";
import type { useOtpInput } from "@/hooks/useOtpInput";

interface OtpCellsProps {
  otp: ReturnType<typeof useOtpInput>;
  label: string;
  hint?: string;
  size?: "md" | "lg";
  numeric?: boolean;
}

/**
 * The row of one-character cells a code is typed into, one per character of the code the hook
 * was made for. Paste fills them all; typing moves along.
 */
export default function OtpCells({
  otp,
  label,
  hint,
  size = "md",
  numeric = false,
}: OtpCellsProps) {
  const cellId = label.toLowerCase().replace(/\s+/g, "-");

  return (
    <div>
      <p className={LABEL}>{label}</p>
      <div
        className="flex gap-2"
        role="group"
        aria-label={label}
        onPaste={otp.handlePaste}
      >
        {otp.code.map((char, index) => (
          <input
            key={index}
            id={`${cellId}-${index}`}
            ref={(el) => {
              otp.inputRefs.current[index] = el;
            }}
            type="text"
            inputMode={numeric ? "numeric" : undefined}
            maxLength={1}
            value={char}
            aria-label={`${label} character ${index + 1} of ${otp.code.length}`}
            onChange={(e) => otp.handleChange(index, e.target.value)}
            onKeyDown={(e) => otp.handleKeyDown(index, e)}
            className={cn(
              "text-center font-mono bg-card border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20",
              size === "lg" ? "w-12 h-12 text-lg" : "w-10 h-10 text-lg",
              !numeric && "uppercase",
            )}
          />
        ))}
      </div>
      {hint && <p className="text-2xs text-text-muted mt-2">{hint}</p>}
    </div>
  );
}
