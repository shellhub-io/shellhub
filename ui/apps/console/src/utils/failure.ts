/**
 * A failure this call site is written to absorb: the promise resolves to undefined and nothing
 * else happens. Pass it to .catch() in place of an empty catch block, so the intent is a name in
 * the code rather than a comment explaining an empty pair of braces.
 *
 * Reach for it only where the failure genuinely has no consequence — a best-effort cleanup, an
 * optional integration that may not be loaded. Where the error is shown somewhere else, say that
 * instead: a react-query mutation surfaces its own error through mutate().
 */
export function ignoreFailure(): undefined {
  return undefined;
}

/**
 * The same, for a value that could not be read: the promise resolves to null, so the caller
 * branches on the absence rather than on an exception.
 */
export function nullOnFailure(): null {
  return null;
}

/**
 * Runs a synchronous action that is allowed to throw, reporting whether it worked. For an API
 * that may not be there at all — an optional renderer, a third-party widget mid-bootstrap, a
 * browser that refuses the call — where the alternative is a try with an empty catch at every
 * call site instead of one named function here.
 */
export function attempt(action: () => void): boolean {
  try {
    action();

    return true;
  } catch {
    return false;
  }
}
