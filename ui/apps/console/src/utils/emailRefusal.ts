const EMAIL_REFUSAL_MESSAGES: Record<string, string> = {
  disposable: "Use a permanent email address",
  unreachable: "This domain cannot receive email. Check the spelling.",
};

/**
 * The message for an email the server refused on Cloud, keyed by the reason it put in the 400's
 * `fields`: a disposable domain or one that cannot receive mail. Returns undefined for any other
 * reason, so the caller keeps its own message.
 */
export function emailRefusalMessage(reason: string | undefined): string | undefined {
  return reason === undefined ? undefined : EMAIL_REFUSAL_MESSAGES[reason];
}
