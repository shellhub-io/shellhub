import {
  useState,
  useRef,
  useCallback,
  useEffect,
  type DragEvent,
  type ChangeEvent,
  type RefObject,
} from "react";

interface UseKeyFileInputOptions {
  validate: (text: string) => boolean;
  onChange: (text: string) => void;
  onFileName?: (name: string) => void;
  disabled?: boolean;
}

interface UseKeyFileInputReturn {
  fileInputRef: RefObject<HTMLInputElement>;
  dragging: boolean;
  inputMode: "file" | "text";
  setInputMode: (mode: "file" | "text") => void;
  fileSizeError: boolean;
  fileReadError: boolean;
  processFile: (file: File) => void;
  handleDrop: (e: DragEvent) => void;
  handleFileInputChange: (e: ChangeEvent<HTMLInputElement>) => void;
  setDragging: (v: boolean) => void;
}

export function useKeyFileInput({
  validate,
  onChange,
  onFileName,
  disabled,
}: UseKeyFileInputOptions): UseKeyFileInputReturn {
  const [inputMode, setInputMode] = useState<"file" | "text">(
    disabled ? "text" : "file",
  );
  const [dragging, setDragging] = useState(false);
  const [fileSizeError, setFileSizeError] = useState(false);
  const [fileReadError, setFileReadError] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const onChangeRef = useRef(onChange);
  const validateRef = useRef(validate);
  const onFileNameRef = useRef(onFileName);

  useEffect(() => {
    onChangeRef.current = onChange;
    validateRef.current = validate;
    onFileNameRef.current = onFileName;
  }, [onChange, validate, onFileName]);

  const processFile = useCallback((file: File) => {
    if (file.size > 512 * 1024) {
      setFileSizeError(true);
      setFileReadError(false);
      return;
    }
    setFileSizeError(false);
    setFileReadError(false);
    const reader = new FileReader();
    reader.onload = () => {
      const text = reader.result as string;
      if (!validateRef.current(text.trim())) return;
      onChangeRef.current(text);
      if (onFileNameRef.current) {
        const base = file.name.replace(/\.[^.]+$/, "");
        onFileNameRef.current(base);
      }
    };
    reader.onerror = () => {
      setFileReadError(true);
    };
    reader.readAsText(file);
  }, []);

  const handleDrop = useCallback(
    (e: DragEvent) => {
      e.preventDefault();
      setDragging(false);
      const file = e.dataTransfer?.files?.[0];
      if (file) processFile(file);
    },
    [processFile],
  );

  const handleFileInputChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) processFile(file);
      e.target.value = "";
    },
    [processFile],
  );

  const handlePaste = useCallback((e: ClipboardEvent) => {
    const active = document.activeElement;
    if (active instanceof HTMLInputElement || active instanceof HTMLTextAreaElement) return;

    const text = e.clipboardData?.getData("text");
    if (text && validateRef.current(text.trim())) {
      e.preventDefault();
      onChangeRef.current(text);
    }
  }, []);

  useEffect(() => {
    if (disabled) return;
    document.addEventListener("paste", handlePaste);
    return () => document.removeEventListener("paste", handlePaste);
  }, [disabled, handlePaste]);

  return {
    fileInputRef,
    dragging,
    fileSizeError,
    fileReadError,
    inputMode,
    setInputMode,
    processFile,
    handleDrop,
    handleFileInputChange,
    setDragging,
  };
}
