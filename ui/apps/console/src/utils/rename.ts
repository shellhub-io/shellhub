import { isSdkError } from "@/api/errors";

/**
 * What to tell the user when renaming a device or container fails: a bad name and a taken name
 * get their own wording, anything else a generic one naming the entity.
 */
export function renameErrorMessage(err: unknown, entityLabel: string): string {
  const status = isSdkError(err) ? err.status : undefined;
  if (status === 400) return `Invalid ${entityLabel} name.`;
  if (status === 409) return `A ${entityLabel} with that name already exists.`;
  return `Failed to rename ${entityLabel}.`;
}

/**
 * The accessible label of the name field for an entity, "Device name" for "device".
 */
export function renameFieldLabel(entityLabel: string): string {
  return `${entityLabel.charAt(0).toUpperCase()}${entityLabel.slice(1)} name`;
}
