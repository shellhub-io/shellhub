import { Fragment } from "react";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useOtpInput } from "@/hooks/useOtpInput";

const GROUP_SPLIT = 4;

interface WizardStepCodeProps {
  /** OTP state owned by the wizard, so the footer's Pair button can read the
   *  value and completeness. */
  otp: ReturnType<typeof useOtpInput>;
  error?: string;
}

/**
 * The "not on that machine" path: the user types the pairing code the agent
 * printed during install (no extra command; it's already in that output) to
 * pair the device into the current namespace. Wizard-native, left-aligned like
 * the install step; pairing is driven from the footer, so there's no button
 * here.
 */
export default function WizardStepCode({ otp, error }: WizardStepCodeProps) {
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: ns } = useNamespace(tenantId);
  const namespace = ns?.name;

  return (
    <div className="py-2 flex flex-col gap-5">
      <div>
        <h2 className="text-lg font-semibold text-text-primary mb-1">
          Enter your device&rsquo;s code
        </h2>
        <p className="text-sm text-text-muted">
          When you ran{" "}
          <code className="font-mono text-xs text-text-primary bg-background border border-border rounded px-1 py-px">
            install.sh
          </code>
          , your device printed a pairing code. Enter it to pair the device into{" "}
          {namespace ? (
            <>
              <span className="font-mono text-primary">{namespace}</span>{" "}
              namespace
            </>
          ) : (
            "your namespace"
          )}
          .
        </p>
      </div>

      <div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
          Pairing code
        </p>
        <div
          className="flex items-center gap-1.5 sm:gap-2"
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
                aria-label={`Character ${index + 1} of 8`}
                onChange={(e) => otp.handleChange(index, e.target.value)}
                onKeyDown={(e) => otp.handleKeyDown(index, e)}
                className="w-10 h-14 sm:w-11 text-center text-xl font-mono uppercase bg-background border border-border rounded-lg text-accent-cyan caret-accent-cyan focus:outline-none focus:border-primary/60 focus:ring-1 focus:ring-primary/25 transition-all duration-150"
              />
            </Fragment>
          ))}
        </div>
      </div>

      {error && (
        <p className="text-sm text-accent-red" role="alert">
          {error}
        </p>
      )}

      <div className="flex items-start gap-2.5 bg-primary/[0.05] border border-primary/15 rounded-xl px-3.5 py-3">
        <InformationCircleIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
        <span className="text-xs text-text-secondary leading-relaxed">
          The code is in the install output, just below the link the agent
          printed. Sitting at that device? Open that link instead.
        </span>
      </div>
    </div>
  );
}
