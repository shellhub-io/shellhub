import { apiErrorFields, isSdkError } from "@/api/errors";
import { isCloud, isEnterprise } from "@/env";

type EntityType = "device" | "container";

const NAMESPACE_DEVICE_LIMIT_MESSAGE =
  "maximum number of accepted devices reached";

function fallback(entityType: EntityType): string {
  return `An error occurred while accepting the ${entityType}. Please try again.`;
}

/**
 * Turns a failed accept into the sentence shown to the user. A 402 is the interesting case: on
 * cloud it means the namespace is over its device limit, and the message asks for an upgrade
 * or a subscription depending on hasSubscription and canSubscribe, because a member who cannot
 * reach billing must not be told to go there. A 403 is a missing permission unless the server's
 * message says the namespace is at the device limit set on it, which the user lifts by removing
 * a device or asking the instance administrator. Anything the SDK did not raise falls back to a
 * generic message naming entityType.
 */
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
      if (err.message === NAMESPACE_DEVICE_LIMIT_MESSAGE)
        return `This namespace has reached its ${entityType} limit. Remove a ${entityType} or ask the instance administrator to raise the limit, then accept this ${entityType} again.`;
      return `You do not have permission to accept ${entityType}s in this namespace.`;
    case 409:
      if ("identity.mac" in apiErrorFields(err))
        return `Another ${entityType} with the same MAC address is connected. Stop its agent, then accept this ${entityType} again.`;
      return `A ${entityType} with this name already exists in the namespace. Please rename the ${entityType} and try again.`;
    default:
      return fallback(entityType);
  }
}
