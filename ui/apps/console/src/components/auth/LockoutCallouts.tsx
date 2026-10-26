import { Callout } from "@shellhub/design-system/primitives";

interface LockoutCalloutsProps {
  error: string | null;
  countdownDisplay: string;
  lockoutExpired: boolean;
  expiredMessage: string;
}

/**
 * The callouts of a form that locks out after too many wrong tries: the error, with the time left
 * when a lockout is running, and expiredMessage once it ends, which replaces the error.
 */
export default function LockoutCallouts({
  error,
  countdownDisplay,
  lockoutExpired,
  expiredMessage,
}: LockoutCalloutsProps) {
  if (lockoutExpired) {
    return <Callout variant="success">{expiredMessage}</Callout>;
  }

  if (!error) return null;

  return (
    <Callout variant="error">
      <span>
        <span>{error}</span>
        {countdownDisplay ? (
          <span className="font-semibold"> ({countdownDisplay})</span>
        ) : null}
      </span>
    </Callout>
  );
}
