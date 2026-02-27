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
    if (file.size > 512 * 1024) return;
    const reader = new FileReader();
    reader.onload = () => {
      const text = reader.result as string;
      onChangeRef.current(text);
      if (onFileNameRef.current) {
        const base = file.name.replace(/\.[^.]+$/, "");
        onFileNameRef.current(base);
      }
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

  const handlePaste = useCallback((e: Event) => {
    const ce = e as ClipboardEvent;
    const text = ce.clipboardData?.getData("text");
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
    inputMode,
    setInputMode,
    processFile,
    handleDrop,
    handleFileInputChange,
    setDragging,
  };
}
