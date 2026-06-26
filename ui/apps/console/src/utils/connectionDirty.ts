// Does the live identity form differ from the saved connection record?
// Normalizes username (trim) and the auth method default so merely opening a
// connection never reads as dirty. keyFingerprint is "" when no usable key is
// selected, or "manual" for a pasted key (always counts as a change).
export interface ConnectionAuthForm {
  username: string;
  authMethod: "password" | "key";
  keyFingerprint: string;
}

export interface ConnectionAuthRecord {
  username: string;
  auth_method: string;
  key_fingerprint: string;
}

export function connectionDirty(
  form: ConnectionAuthForm,
  record: ConnectionAuthRecord,
): boolean {
  const savedMethod = record.auth_method === "key" ? "key" : "password";
  if (form.username.trim() !== (record.username ?? "").trim()) return true;
  if (form.authMethod !== savedMethod) return true;
  if (
    form.authMethod === "key" &&
    form.keyFingerprint !== "" &&
    form.keyFingerprint !== (record.key_fingerprint ?? "")
  ) {
    return true;
  }
  return false;
}
