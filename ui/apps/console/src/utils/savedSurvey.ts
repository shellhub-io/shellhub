import { attempt } from "@/utils/failure";

const storageKey = (userId: string, surveyId: string) =>
  `onboardingSurvey:${userId}:${surveyId}`;

/**
 * The id of the response this user already sent to this survey from this browser, or null. Only
 * the id is kept, never the answers, so nothing the user typed outlives the session. Keyed by
 * survey as well, so moving the trigger to another survey starts afresh rather than updating a
 * response to the old one. Storage that is blocked or empty, and a user without an id, read as
 * null.
 */
export function readSavedResponseId(
  userId: string | null,
  surveyId: string,
): string | null {
  if (!userId) return null;
  try {
    return localStorage.getItem(storageKey(userId, surveyId)) || null;
  } catch {
    return null;
  }
}

/**
 * Remembers the response this user sent to this survey, so a later visit updates it instead of
 * sending another. A browser that refuses to store it, or a user without an id, gets a second
 * response on the next visit and nothing worse.
 */
export function saveResponseId(
  userId: string | null,
  surveyId: string,
  responseId: string,
) {
  if (!userId) return;
  attempt(() => localStorage.setItem(storageKey(userId, surveyId), responseId));
}
