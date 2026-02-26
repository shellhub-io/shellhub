export function useRecoveryCodeActions() {
  const handleDownload = (codes: string[]): void => {
    const content = `ShellHub Recovery Codes\n\n${codes.join("\n")}\n\nKeep these codes in a safe place. Each code can only be used once.`;
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "shellhub-recovery-codes.txt";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleCopy = async (codes: string[]): Promise<void> => {
    try {
      const content = codes.join("\n");
      await navigator.clipboard.writeText(content);
    } catch (error) {
      // Fallback: create temporary textarea for older browsers
      const textarea = document.createElement("textarea");
      textarea.value = codes.join("\n");
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.select();
      try {
        document.execCommand("copy");
      } catch {
        console.error("Failed to copy recovery codes");
      }
      document.body.removeChild(textarea);
    }
  };

  return { handleDownload, handleCopy };
}
