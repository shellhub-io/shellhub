/**
 * Saves a freshly generated set of recovery codes as a text file through a normal browser
 * download. The codes are shown once, so this is the only chance to save them.
 */
export function downloadRecoveryCodes(codes: string[]): void {
  const content = `ShellHub Recovery Codes\n\n${codes.join("\n")}\n\nKeep these codes in a safe place. Each code can only be used once.`;
  const url = URL.createObjectURL(new Blob([content], { type: "text/plain" }));
  const a = document.createElement("a");
  a.href = url;
  a.download = "shellhub-recovery-codes.txt";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
