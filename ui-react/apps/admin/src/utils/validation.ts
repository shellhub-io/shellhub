export function validatePassword(value: string): string | null {
  if (value.length < 5 || value.length > 32)
    return "Password must be 5–32 characters long";
  return null;
}
