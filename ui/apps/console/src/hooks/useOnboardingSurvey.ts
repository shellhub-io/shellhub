import { useMutation, useQuery } from "@tanstack/react-query";
import {
  findOnboardingSurvey,
  type OnboardingSurvey,
} from "@/pages/setup/onboardingSurvey";

async function fetchOnboardingSurvey(
  baseUrl: string,
): Promise<OnboardingSurvey | null> {
  const res = await fetch(`${baseUrl}/environment`);
  if (!res.ok) throw new Error(`survey environment: HTTP ${res.status}`);
  return findOnboardingSurvey(await res.json());
}

/**
 * The onboarding survey served by the Formbricks workspace at baseUrl, its client API root
 * (`https://<host>/api/v1/client/<workspaceId>`). Resolves to null when the workspace has no
 * survey setup can render, and fails when the workspace cannot be reached; either way the caller
 * skips the survey. Fetched once per page, since the choice order is shuffled at parse time.
 */
export function useOnboardingSurvey(baseUrl: string) {
  return useQuery({
    queryKey: ["onboarding-survey", baseUrl],
    queryFn: () => fetchOnboardingSurvey(baseUrl),
    enabled: baseUrl !== "",
    staleTime: Infinity,
    retry: false,
    refetchOnWindowFocus: false,
  });
}

interface SubmitVariables {
  surveyId: string;
  responseId: string | null;
  data: Record<string, string | string[]>;
}

async function submitResponse(
  baseUrl: string,
  { surveyId, responseId, data }: SubmitVariables,
): Promise<string> {
  const res = await fetch(
    responseId
      ? `${baseUrl}/responses/${encodeURIComponent(responseId)}`
      : `${baseUrl}/responses`,
    {
      method: responseId ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(
        responseId
          ? { finished: true, data }
          : { surveyId, finished: true, data },
      ),
    },
  );
  if (!res.ok) throw new Error(`survey response: HTTP ${res.status}`);
  if (responseId) return responseId;

  const body = (await res.json()) as { data?: { id?: unknown } };
  if (typeof body.data?.id !== "string") {
    throw new Error("survey response: no id returned");
  }
  return body.data.id;
}

/**
 * Sends the onboarding answers to the Formbricks workspace at baseUrl. The first send creates the
 * response; passing the id it resolved to updates that response instead, so a user who goes back
 * and changes an answer leaves one response, not two.
 */
export function useSubmitOnboardingSurvey(baseUrl: string) {
  return useMutation({
    mutationFn: (vars: SubmitVariables) => submitResponse(baseUrl, vars),
  });
}
