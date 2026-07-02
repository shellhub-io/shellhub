import { Fragment } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import { useOtpInput } from "@/hooks/useOtpInput";

const CODE_LENGTH = 8;
const GROUP_SPLIT = 4;

/**
 * Collects a device pairing code typed by hand — read off a headless device's
 * terminal. The code is shown the way the device prints it: eight monospace
 * cells split into two groups of four, in the console's terminal accent.
 *
 * By default it navigates to the accept page with the canonical (ungrouped)
 * code. Pass `onSubmit` to handle the code in place instead — the pairing modal
 * uses this to resolve and accept without leaving the page.
 */
export default function PairingCodeForm({
  submitLabel = "Continue",
  onSubmit,
}: {
  submitLabel?: string;
  onSubmit?: (code: string) => void;
}) {
  const navigate = useNavigate();
  const otp = useOtpInput(CODE_LENGTH, true);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        if (!otp.isComplete) return;
        const code = otp.getValue();
        if (onSubmit) {
          onSubmit(code);
        } else {
          void navigate(`/accept-device?code=${code}`);
        }
      }}
      className="space-y-6"
    >
      <div
        className="flex items-center justify-center gap-1.5 sm:gap-2"
        role="group"
        aria-label="Pairing code"
        onPaste={otp.handlePaste}
      >
        {otp.code.map((char, index) => (
          <Fragment key={index}>
            {index === GROUP_SPLIT && (
              <span
                aria-hidden="true"
                className="px-0.5 font-mono text-xl text-text-muted/40 select-none"
              >
                &ndash;
              </span>
            )}
            <input
              ref={(el) => {
                otp.inputRefs.current[index] = el;
              }}
              type="text"
              inputMode="text"
              autoCapitalize="characters"
              spellCheck={false}
              maxLength={1}
              value={char}
              aria-label={`Character ${index + 1} of ${CODE_LENGTH}`}
              onChange={(e) => otp.handleChange(index, e.target.value)}
              onKeyDown={(e) => otp.handleKeyDown(index, e)}
              className="w-10 h-14 sm:w-11 text-center text-xl font-mono uppercase bg-background border border-border rounded-lg text-accent-cyan caret-accent-cyan focus:outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/25 transition-all duration-150"
            />
          </Fragment>
        ))}
      </div>

      <Button
        type="submit"
        variant="primary"
        size="lg"
        fullWidth
        disabled={!otp.isComplete}
      >
        {submitLabel}
      </Button>
    </form>
  );
}
