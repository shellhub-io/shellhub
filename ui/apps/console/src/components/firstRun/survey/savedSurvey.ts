import { attempt } from "@/utils/failure";
import type { SurveyAnswers } from "./onboardingSurvey";

/**
 * A survey response already sent from this browser: the id to update it by, and the answers it
 * carried, so the form opens on them.
 */
export interface SavedSurvey {
  responseId: string;
  answers: SurveyAnswers;
}

const storageKey = (userId: string) => `onboardingSurvey:${userId}`;

function isSavedSurvey(value: unknown): value is SavedSurvey {
  if (typeof value !== "object" || value === null) return false;
  const { responseId, answers } = value as Partial<SavedSurvey>;
  return (
    typeof responseId === "string" &&
    typeof answers === "object" &&
    answers !== null &&
    ["choices", "other", "contact", "consent"].every(
      (k) =>
        typeof (answers as unknown as Record<string, unknown>)[k] === "object",
    )
  );
}

/**
 * The response this user already sent from this browser, or null. Storage that is blocked or
 * holds something else reads as null, so the survey shows again rather than failing.
 */
export function readSavedSurvey(userId: string): SavedSurvey | null {
  try {
    const raw = localStorage.getItem(storageKey(userId));
    const value: unknown = raw ? JSON.parse(raw) : null;
    return isSavedSurvey(value) ? value : null;
  } catch {
    return null;
  }
}

/**
 * Remembers the response this user sent, so a later visit updates it instead of sending another.
 * Returns false when the browser refuses to store it.
 */
export function writeSavedSurvey(userId: string, saved: SavedSurvey): boolean {
  return attempt(() =>
    localStorage.setItem(storageKey(userId), JSON.stringify(saved)),
  );
}

/**
 * Lays saved answers over a survey's starting answers, so a question added since keeps its blank
 * and one removed since is dropped when the response is built.
 */
export function mergeAnswers(
  base: SurveyAnswers,
  saved: SurveyAnswers,
): SurveyAnswers {
  return {
    anonymous: saved.anonymous === true,
    choices: { ...base.choices, ...saved.choices },
    other: { ...base.other, ...saved.other },
    contact: { ...base.contact, ...saved.contact },
    consent: { ...base.consent, ...saved.consent },
  };
}
