import { apiErrorFields, isSdkError } from "@/api/errors";
import { features } from "@/edition";

type EntityType = "device" | "container";

const NAMESPACE_DEVICE_LIMIT_MESSAGE =
  "maximum number of accepted devices reached";

const REPLACE_FORBIDDEN_MESSAGE =
  "accepting the device replaces an offline device with the same identity, which needs permission to remove devices";

/**
 * Tells whether a failed accept was refused because it would replace an offline device and the
 * user may not remove devices. Trying again changes nothing until the namespace or the user does.
 */
export function isReplaceForbidden(err: unknown): boolean {
  return (
    isSdkError(err) &&
    err.status === 403 &&
    err.message === REPLACE_FORBIDDEN_MESSAGE
  );
}

/**
 * Names the offline device a failed accept would replace when the server refused it for want of
 * confirmation, a 409 naming the device in its replace field, and is empty for any other error.
 */
export function replacementToConfirm(err: unknown): string {
  if (!isSdkError(err) || err.status !== 409) return "";
  return apiErrorFields(err).replace ?? "";
}

function fallback(entityType: EntityType): string {
  return `An error occurred while accepting the ${entityType}. Please try again.`;
}

/**
 * Turns a failed accept into the sentence shown to the user. A 402 is the interesting case: on
 * cloud it means the namespace is over its device limit, and the message asks for an upgrade
 * or a subscription depending on hasSubscription and canSubscribe, because a member who cannot
 * reach billing must not be told to go there. A 403 is a missing permission unless the server's
 * message says the namespace is at the device limit set on it, which the user lifts by removing
 * a device or asking the instance administrator, or that the accept would replace an offline
 * device and the user may not remove devices. A 409 naming a device in its replace field means
 * the accept would replace that device without having confirmed it, which happens when it went
 * offline after the console checked. Anything the SDK did not raise falls back to a
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
      if (features().billing) {
        if (hasSubscription) {
          return canSubscribe
            ? `Your subscription needs attention before this namespace can accept ${entityType}s. Open Billing to finish or repair it — your ${entityType} count is not the problem.`
            : `Your namespace's subscription needs attention before it can accept ${entityType}s. Ask the namespace owner to check Billing.`;
        }
        return canSubscribe
          ? `Your subscription plan has reached its ${entityType} limit. Please update your billing plan to accept more ${entityType}s.`
          : `Your namespace has reached its ${entityType} limit. Ask the namespace owner to update the billing plan.`;
      }
      if (features().instanceLicense)
        return `Your instance has reached its ${entityType} limit. Please update your license to accept more ${entityType}s or contact the instance administrator.`;
      return fallback(entityType);
    }
    case 403:
      if (isReplaceForbidden(err))
        return `Accepting this ${entityType} replaces an offline ${entityType} that reports the same identity, which only a member allowed to remove devices can do.`;
      if (err.message === NAMESPACE_DEVICE_LIMIT_MESSAGE)
        return `This namespace has reached its ${entityType} limit. Remove a ${entityType} or ask the instance administrator to raise the limit, then accept this ${entityType} again.`;
      return `You do not have permission to accept ${entityType}s in this namespace.`;
    case 409: {
      const replaced = replacementToConfirm(err);
      if (replaced)
        return `Accepting this ${entityType} replaces ${replaced}, which went offline after the check. Reopen the accept to confirm the replacement.`;
      const fields = apiErrorFields(err);
      if ("identity.mac" in fields)
        return `Another ${entityType} with the same MAC address is connected. Stop its agent, then accept this ${entityType} again.`;
      return `A ${entityType} with this name already exists in the namespace. Please rename the ${entityType} and try again.`;
    }
    default:
      return fallback(entityType);
  }
}
