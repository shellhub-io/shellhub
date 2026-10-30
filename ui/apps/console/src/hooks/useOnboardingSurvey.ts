import { useMutation, useQuery } from "@tanstack/react-query";
import {
  recordOnboardingSurveyMutation,
  type UserOnboardingSurvey,
} from "@/client";
import { getConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import {
  findOnboardingSurvey,
  type OnboardingSurvey,
} from "@/utils/onboardingSurvey";

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
 * The onboarding survey from the Formbricks workspace whose client API root is the configured
 * onboardingUrl. `visible` is true while the survey loads and once it has loaded, with `survey`
 * null until then; it turns false when `enabled` is false, no URL is configured, the workspace
 * does not answer within a few seconds or answers with an error, or it has no survey the console
 * can render. Fetched once per page, since the choice order is shuffled at parse time.
 */
export function useOnboardingSurvey(enabled: boolean) {
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

  return {
    visible: root !== "" && (query.isPending || survey !== null),
    survey,
  };
}

/**
 * Records on the server whether the signed-in user answered or skipped the onboarding survey. The
 * auth store takes the outcome as the call starts, so the console moves on at once; when the call
 * fails, the error is logged, the server still holds pending and the next sign-in brings the
 * survey back.
 */
export function useRecordOnboardingSurvey() {
  const setOnboardingSurvey = useAuthStore((s) => s.setOnboardingSurvey);
  const mutation = useMutation({
    ...recordOnboardingSurveyMutation(),
    onError: (err) =>
      console.error("onboarding survey: failed to record the outcome", err),
  });

  return (outcome: Exclude<UserOnboardingSurvey, "pending">) => {
    setOnboardingSurvey(outcome);
    mutation.mutate({ body: { outcome } });
  };
}

interface SubmitVariables {
  surveyId: string;
  data: Record<string, string | string[]>;
}

async function submitResponse({ surveyId, data }: SubmitVariables) {
  if (import.meta.env.MODE === "development") {
    console.info("Onboarding survey response not sent in development", data);
    return;
  }

  const res = await surveyFetch(`${surveyApiRoot()}/responses`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ surveyId, finished: true, data }),
  });
  if (!res.ok) throw new Error(`survey response: HTTP ${res.status}`);
}

/**
 * Sends the onboarding answers to the configured Formbricks workspace as one finished response. A
 * request that does not finish within a few seconds fails like any other. Under the Vite dev
 * server nothing is sent: the answers are logged and the call resolves as if it had been, so local
 * runs never land in the real survey.
 */
export function useSubmitOnboardingSurvey() {
  return useMutation({ mutationFn: submitResponse });
}
