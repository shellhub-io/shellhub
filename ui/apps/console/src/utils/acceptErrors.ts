import { isSdkError } from "@/api/errors";
import { isCloud, isEnterprise } from "@/env";

type EntityType = "device" | "container";

function fallback(entityType: EntityType): string {
  return `An error occurred while accepting the ${entityType}. Please try again.`;
}

export function getAcceptErrorMessage(
  err: unknown,
  hasSubscription: boolean,
  canSubscribe: boolean,
  entityType: EntityType = "device",
): string {
  if (!isSdkError(err)) return fallback(entityType);

  switch (err.status) {
    case 402: {
      if (isCloud()) {
        if (hasSubscription) {
          return canSubscribe
            ? `Your subscription needs attention before this namespace can accept ${entityType}s. Open Billing to finish or repair it — your ${entityType} count is not the problem.`
            : `Your namespace's subscription needs attention before it can accept ${entityType}s. Ask the namespace owner to check Billing.`;
        }
        return canSubscribe
          ? `Your subscription plan has reached its ${entityType} limit. Please update your billing plan to accept more ${entityType}s.`
          : `Your namespace has reached its ${entityType} limit. Ask the namespace owner to update the billing plan.`;
      }
      if (isEnterprise())
        return `Your instance has reached its ${entityType} limit. Please update your license to accept more ${entityType}s or contact the instance administrator.`;
      return fallback(entityType);
    }
    case 403:
      return `You do not have permission to accept ${entityType}s in this namespace.`;
    case 409:
      return `A ${entityType} with this name already exists in the namespace. Please rename the ${entityType} and try again.`;
    default:
      return fallback(entityType);
  }
}
