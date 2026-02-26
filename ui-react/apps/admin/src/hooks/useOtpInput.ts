import { useState, useRef, KeyboardEvent } from "react";

export function useOtpInput(length: number = 6, alphanumeric: boolean = false) {
  const [code, setCode] = useState<string[]>(Array(length).fill(""));
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const handleChange = (index: number, value: string) => {
    // Validate input based on mode
    const pattern = alphanumeric ? /^[a-zA-Z0-9]$/ : /^\d$/;
    if (value && !pattern.test(value)) return;

    setCode((prev) => {
      const newCode = [...prev];
      newCode[index] = alphanumeric ? value.toUpperCase() : value;
      return newCode;
    });

    // Auto-advance to next field
    if (value && index < length - 1) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: KeyboardEvent<HTMLInputElement>) => {
    // Backspace: clear current and move to previous
    if (e.key === "Backspace") {
      if (!code[index] && index > 0) {
        setCode((prev) => {
          const newCode = [...prev];
          newCode[index - 1] = "";
          return newCode;
        });
        inputRefs.current[index - 1]?.focus();
      } else {
        setCode((prev) => {
          const newCode = [...prev];
          newCode[index] = "";
          return newCode;
        });
      }
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const rawData = e.clipboardData.getData("text");
    const pastedData = alphanumeric
      ? rawData.replace(/[^a-zA-Z0-9]/g, "").toUpperCase()
      : rawData.replace(/\D/g, "");
    if (pastedData.length === length) {
      setCode(pastedData.split(""));
      inputRefs.current[length - 1]?.focus();
    }
  };

  const reset = () => {
    setCode(Array(length).fill(""));
    inputRefs.current[0]?.focus();
  };

  const getValue = () => code.join("");
  const isComplete = code.every((digit) => digit !== "");

  return {
    code,
    inputRefs,
    handleChange,
    handleKeyDown,
    handlePaste,
    reset,
    getValue,
    isComplete,
  };
}
