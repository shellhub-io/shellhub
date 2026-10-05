import { useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { getConfig } from "@/env";
import {
  findOnboardingSurvey,
  type OnboardingSurvey,
  type SurveyAnswers,
} from "@/utils/onboardingSurvey";
import { readSavedResponseId, saveResponseId } from "@/utils/savedSurvey";

const REQUEST_TIMEOUT_MS = 5000;

function surveyApiRoot(): string {
  return getConfig().onboardingUrl.replace(/\/+$/, "");
}

async function surveyFetch(url: string, init?: RequestInit) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
  try {
    return await fetch(url, { ...init, signal: controller.signal });
  } finally {
    clearTimeout(timer);
  }
}

async function fetchOnboardingSurvey(
  root: string,
): Promise<OnboardingSurvey | null> {
  const res = await surveyFetch(`${root}/environment`);
  if (!res.ok) throw new Error(`survey environment: HTTP ${res.status}`);
  return findOnboardingSurvey(await res.json());
}

/**
 * The survey step of a first-run trail, from the Formbricks workspace whose client API root is the
 * configured onboardingUrl. `visible` is true while the survey loads and once it has loaded; it
 * turns false, and the caller drops the step, when `enabled` is false, no URL is configured, the
 * workspace does not answer within a few seconds or answers with an error, or it has no survey the
 * console can render. Fetched once per page, since the choice order is shuffled at parse time.
 *
 * `record` takes the answers and response id a send resolved to. `stepProps` hands them back to
 * OnboardingStep, so stepping back to the survey shows the same answers and updates the same
 * response. Given a `userId`, the response id is also kept in this browser per user and survey,
 * and `responseId` starts from it on a later visit.
 */
export function useOnboardingSurveyStep({
  enabled,
  userId = null,
}: {
  enabled: boolean;
  userId?: string | null;
}) {
  const root = enabled ? surveyApiRoot() : "";
  const query = useQuery({
    queryKey: ["onboarding-survey", root],
    queryFn: () => fetchOnboardingSurvey(root),
    enabled: root !== "",
    staleTime: Infinity,
    retry: false,
    refetchOnWindowFocus: false,
  });
  const survey = query.data ?? null;
  const [sent, setSent] = useState<{
    answers: SurveyAnswers;
    responseId: string;
  } | null>(null);
  const responseId =
    sent?.responseId ??
    (survey ? readSavedResponseId(userId, survey.id) : null);

  return {
    visible: root !== "" && (query.isPending || survey !== null),
    responseId,
    stepProps: {
      survey,
      initialAnswers: sent?.answers ?? null,
      responseId,
    },
    record: (answers: SurveyAnswers, id: string) => {
      if (survey) saveResponseId(userId, survey.id, id);
      setSent({ answers, responseId: id });
    },
  };
}

interface SubmitVariables {
  surveyId: string;
  responseId: string | null;
  data: Record<string, string | string[]>;
}

function sendJson(url: string, method: "POST" | "PUT", body: unknown) {
  return surveyFetch(url, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}

async function submitResponse({
  surveyId,
  responseId,
  data,
}: SubmitVariables): Promise<string> {
  if (import.meta.env.MODE === "development") {
    console.info("Onboarding survey response not sent in development", data);
    return responseId ?? "development";
  }

  const root = surveyApiRoot();
  if (responseId) {
    const res = await sendJson(
      `${root}/responses/${encodeURIComponent(responseId)}`,
      "PUT",
      { finished: true, data },
    );
    if (res.ok) return responseId;
    if (res.status !== 404) {
      throw new Error(`survey response: HTTP ${res.status}`);
    }
  }

  const res = await sendJson(`${root}/responses`, "POST", {
    surveyId,
    finished: true,
    data,
  });
  if (!res.ok) throw new Error(`survey response: HTTP ${res.status}`);

  const body = (await res.json()) as { data?: { id?: unknown } };
  if (typeof body.data?.id !== "string") {
    throw new Error("survey response: no id returned");
  }
  return body.data.id;
}

/**
 * Sends the onboarding answers to the configured Formbricks workspace and resolves to the
 * response id. The first send creates the response; passing the id it resolved to updates that
 * response instead, so a user who goes back and changes an answer leaves one response, not two.
 * An update whose response no longer exists, say deleted in the dashboard, creates a new one. A
 * request that does not finish within a few seconds fails like any other. Under the Vite dev
 * server nothing is sent: the answers are logged and the call resolves as if it had been, so
 * local runs never land in the real survey.
 */
export function useSubmitOnboardingSurvey() {
  return useMutation({ mutationFn: submitResponse });
}
