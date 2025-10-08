<template>
  <div
    class="file-text-capture"
    :class="{ 'is-disabled': disabled }"
    ref="rootEl"
    @paste.capture="onPaste"
  >
    <v-file-upload
      v-if="mode === 'file' && !textOnly"
      ref="uploadEl"
      v-model="files"
      :accept="effectiveAccept"
      :max-size
      :multiple="false"
      :disabled
      :show-size
      density="comfortable"
      :border="localError && 'opacity-100 error'"
      :class="localError ? 'text-error bg-v-theme-surface' : 'py-3 bg-v-theme-surface'"
      @update:model-value="onFiles"
      data-test="file-text-capture"
    >
      <template #icon>
        <slot name="icon">
          <v-col cols="12" sm="12" class="py-0">
            <v-icon :size="smAndDown ? 40 : 50">mdi-file-upload-outline</v-icon>
          </v-col>
        </slot>
      </template>

      <template #title>
        <slot name="title">
          <v-row no-gutters class="d-flex justify-start">
            <v-col cols="12">
              <div class="ftc-title text-subtitle-1 font-weight-medium">
                Paste, drop a file or <span class="text-primary">click to browse</span>
              </div>
              <div class="ftc-sub text-body-2 mt-1">
                Accepts .pub, .pem, .key, .txt (max {{ Math.round(props.maxSize / 1024) }} KB)
              </div>
              <div class="ftc-desc text-caption mt-1">
                {{ descriptionText }}
              </div>
            </v-col>
          </v-row>

        </slot>
      </template>

      <template #item="{ file, props: itemProps }">
        <v-file-upload-item
          v-bind="itemProps"
          :key="file?.name || file?.lastModified"
          lines="one"
          nav
        >
          <template #prepend>
            <v-avatar size="32" rounded>
              <v-icon>mdi-file-document-outline</v-icon>
            </v-avatar>
          </template>

          <template #clear="{ props: clearProps }">
            <v-btn
              color="error"
              v-bind="clearProps"
              icon="mdi-close"
              size="small"
              variant="text"
              @click.stop="onClearClick(clearProps)"
              title="Remove file"
            />
          </template>
        </v-file-upload-item>
      </template>
    </v-file-upload>

    <div v-else class="mt-2">
      <v-textarea
        ref="textareaRef"
        v-model="textModel"
        :label="textareaLabel"
        :hint="textareaHint"
        :messages="descriptionText"
        :disabled
        auto-grow
        rows="3"
        :error-messages="localError"
        data-test="ftc-textarea"
        @update:model-value="onTextInput"
      >
        <template #append-inner>
          <v-btn
            icon="mdi-upload"
            variant="text"
            size="small"
            :disabled="disabled"
            @click="switchToFileMode"
            title="Return to file drop zone"
          />
        </template>
      </v-textarea>
    </div>

    <v-row v-if="mode === 'file' && localError" no-gutters class="mt-1">
      <v-col cols="12">
        <div class="text-error text-caption" data-test="ftc-file-error">
          {{ localError }}
        </div>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { useDisplay } from "vuetify";

type Mode = "file" | "text";

/**
 * Component Props
 *
 * This component can work in two modes:
 * - "file": Upload, drag & drop, or paste a file whose content will be read as text.
 * - "text": Directly enter or paste text (e.g., SSH key content).
 *
 * It’s designed to be flexible for both PublicKey and PrivateKey Add or Edit dialogs.
 */
const props = withDefaults(defineProps<{
  /**
   * v-model bound value.
   * Holds the text content read from a file or entered manually.
   */
  modelValue?: string;

  /**
   * Accepted file extensions and MIME types.
   * Example: ".pub,.pem,.key,.txt,text/plain"
   */
  accept?: string;

   /**
   * Accepted file without extension
   * Example: Files created from ssh-keygen
   */
  allowExtensionless?: boolean;

  /**
   * Maximum allowed file size in bytes (default: 512 KB).
   * Prevents very large keys or text blobs.
   */
  maxSize?: number;

  /**
   * Disables all interactions, including file uploads and text input.
   */
  disabled?: boolean;

  /**
   * Whether to show the file size in the upload list (Vuetify prop).
   */
  showSize?: boolean;

  /**
   * A preloaded File object. If passed, the component will automatically
   * read and populate its content as text.
   */
  pastedFile?: File | null;

  /**
   * Optional validation function to check if the read text is valid.
   * Example: `(t) => isKeyValid('public', t)`
   */
  validator?:((text: string) => boolean) | null;

  /**
   * Error message to show when the validator fails.
   */
  invalidMessage?: string;

  /**
   * Enables global and local paste handling.
   * If true, you can paste both text and files directly into the component.
   */
  enablePaste?: boolean;

  /**
   * Label for the textarea mode.
   */
  textareaLabel?: string;

  /**
   * Hint text shown under the textarea (Vuetify hint prop).
   */
  textareaHint?: string;

  /**
   * Description text displayed below the textarea or file upload title.
   * Example: “Supports RSA, DSA, ECDSA and ED25519 key types…”
   */
  descriptionText?: string;

  /**
   * If true, the component starts directly in text mode (bypasses file upload UI).
   */
  startInText?: boolean;

  /**
   * If true, completely disables file upload UI — acts purely as a textarea input.
   * Useful for Edit dialogs that should not allow re-uploading.
   */
  textOnly?: boolean;
}>(), {
  modelValue: "",
  accept: ".pub,.pem,.key,.txt,text/plain,application/x-pem-file,application/octet-stream",
  allowExtensionless: false,
  maxSize: 512 * 1024,
  disabled: false,
  showSize: true,
  density: "compact",
  pastedFile: null,
  validator: null,
  invalidMessage: "Invalid content.",
  enablePaste: true,
  textareaLabel: "Content",
  textareaHint: "",
  descriptionText: "",
  startInText: false,
  textOnly: false,
});

/**
 * Emits
 *
 * - update:modelValue → triggered when the text content changes.
 * - error → emitted whenever a validation or file read error occurs.
 * - file-name → emitted when a file is successfully read.
 * - file-processed → emitted after a file has been fully read and validated.
 * - mode-changed → emitted whenever a change from the drop zone or the text area occurs.
 */
const emit = defineEmits<{
  "update:modelValue": [value: string];
  error: [message: string];
  "file-name": [filename: string];
  "file-processed": [];
  "mode-changed": ["file" | "text"];
}>();

// Refs and state
const rootEl = ref<HTMLElement | null>(null);
const mode = ref<Mode>("file");
const files = ref<File[]>([]);
const localError = ref("");
const uploadEl = ref();
const textareaRef = ref<{ $el?: HTMLElement } | null>(null);
const internalUpdate = ref(false);
const { smAndDown } = useDisplay();

// v-model binding
const textModel = defineModel<string>({ default: "" });

/**
 * When allowing extensionless files, we give the OS picker an empty accept string
 * so it doesn’t hide files like `id_ed25519`. Otherwise, use the provided accept.
 */
const effectiveAccept = computed(() => (props.allowExtensionless ? "" : props.accept));

// Sync files with internal state
const setFiles = async (val: File[]) => {
  internalUpdate.value = true;
  files.value = val;
  await nextTick();
  internalUpdate.value = false;
};

// Sets a local + emitted error message
const setError = (msg: string) => {
  localError.value = msg;
  if (msg) emit("error", msg);
};

// Clears the selected file and resets model value
const clearSelection = async () => {
  await setFiles([]);
  emit("update:modelValue", "");
};

// Reads a File as UTF-8 text
const readFileAsText = (file: File) => new Promise<string>((resolve, reject) => {
  const reader = new FileReader();
  reader.onerror = () => reject(new Error("Failed to read file"));
  reader.onload = () => resolve(String(reader.result ?? ""));
  reader.readAsText(file);
});

/** Cross-platform basename without non-null assertions */
const getBaseName = (nameRaw: string) => {
  const safe = String(nameRaw ?? "");
  const parts = safe.split(/[/\\]/);
  return parts.length ? parts[parts.length - 1] : "";
};

/**
 * Helper: Is the filename one of the canonical ssh-keygen private keys?
 * e.g., id_rsa, id_dsa, id_ecdsa, id_ed25519 (+ optional _sk)
 */
const isCanonicalSSHPrivateKeyName = (nameRaw: string) => {
  const base = getBaseName(nameRaw);
  return /^id_(rsa|dsa|ecdsa|ed25519)(?:_sk)?$/i.test(base);
};

/**
 * Checks if file type matches accepted types.
 * Lenient when:
 * - allowExtensionless AND filename has no extension AND (mime is empty or octet-stream)
 * - filename matches canonical ssh private key names
 */
const isTypeAccepted = (file: File) => {
  // Always allow canonical ssh-keygen private-key basenames
  if (isCanonicalSSHPrivateKeyName(file.name)) return true;

  // Determine extension ('.xxx' only if dot is not the first char)
  const dot = file.name.lastIndexOf(".");
  const hasExt = dot > 0; // dot at index 0 means hidden file, not an extension
  const ext = hasExt ? file.name.toLowerCase().slice(dot) : "";
  const mime = (file.type || "").toLowerCase();

  // If we allow extensionless, permit files with no extension and unknown/generic MIME
  if (props.allowExtensionless && !hasExt && (!mime || mime === "application/octet-stream")) {
    return true;
  }

  // Otherwise, fall back to the provided accept list
  if (!props.accept) return true;
  const tokens = props.accept.split(",").map((s) => s.trim().toLowerCase()).filter(Boolean);
  return tokens.some((t) => t === ext
    || (mime && (mime === t || mime.includes(t)))
    || t === "*/*");
};

// Focuses the textarea element programmatically
const focusTextarea = async () => {
  await nextTick();
  const host = textareaRef.value?.$el as HTMLElement | undefined;
  const el = host?.querySelector("textarea") as HTMLTextAreaElement | null;
  el?.focus?.();
};

// Reads and validates a selected file
const useFile = async (file: File) => {
  await clearSelection();
  setError("");

  if (!isTypeAccepted(file)) {
    setError("Unsupported file type.");
    return;
  }
  if (file.size > props.maxSize) {
    setError(`File too large. Max ${Math.round(props.maxSize / 1024)} KB.`);
    return;
  }

  try {
    const text = (await readFileAsText(file)).trim();
    if (!text) {
      setError("Empty file content.");
      return;
    }

    if (props.validator && !props.validator(text)) {
      setError(props.invalidMessage || "Invalid content.");
      return;
    }

    await setFiles([file]);
    emit("file-name", file.name);
    emit("update:modelValue", text);
    emit("file-processed");
  } catch {
    setError("Could not read the file.");
  }
};

// Handles input from file selection / drag & drop
const onFiles = async (val: File[] | File | null) => {
  if (internalUpdate.value) return;
  if (!val) {
    await clearSelection();
    setError("");
    return;
  }
  const picked = Array.isArray(val) ? val[0] : val;
  if (!picked) {
    await clearSelection();
    setError("");
    return;
  }
  await useFile(picked);
};

// Clear button on file item
const onClearClick = async (clearProps: Record<string, unknown>) => {
  const handler = (clearProps as { onClick?: () => void }).onClick;
  if (typeof handler === "function") handler();
  await clearSelection();
  setError("");
};

// Validates text input in textarea mode
const onTextInput = (val: string) => {
  setError("");
  const text = (val ?? "").trim();
  if (!text) return;
  if (props.validator && !props.validator(text)) {
    setError(props.invalidMessage || "Invalid content.");
  }
};

// Switch from text mode back to file upload UI
const switchToFileMode = async () => {
  mode.value = "file";
  emit("mode-changed", "file");
  const f = files.value?.[0];
  if (f) await useFile(f);
};

// Determines if the target element is editable (so paste shouldn’t trigger file read)
const isEditableTarget = (t: EventTarget | null) => {
  const el = t as HTMLElement | null;
  if (!el) return false;
  const tag = el.tagName;
  if (el.isContentEditable) return true;
  if (!tag) return false;
  return tag === "INPUT" || tag === "TEXTAREA" || el.getAttribute("role") === "textbox";
};

// Handles paste events for both text and file content
const onPaste = async (e: ClipboardEvent) => {
  if (!props.enablePaste || props.disabled) return;

  if (isEditableTarget(e.target)) return;

  const dt = e.clipboardData;
  if (!dt) return;

  const list = Array.from(dt.files || []);
  if (list.length > 0) {
    e.preventDefault();
    if (mode.value !== "file") mode.value = "file";
    await useFile(list[0]);
    return;
  }

  const text = dt.getData("text/plain");
  if (text && text.trim()) {
    e.preventDefault();
    mode.value = "text";
    emit("mode-changed", "text");
    await setFiles([]);
    const t = text.trim();
    emit("update:modelValue", t);
    onTextInput(t);
    await focusTextarea();
  }
};

// Global paste listener — allows pasting even when not focused
const globalPasteListener = (e: ClipboardEvent) => {
  if (!rootEl.value) return;
  onPaste(e);
};

// Lifecycle hooks
onMounted(() => {
  if (props.startInText || props.textOnly) mode.value = "text";
  window.addEventListener("paste", globalPasteListener, { capture: true });
});

onUnmounted(() => {
  window.removeEventListener("paste", globalPasteListener, { capture: true });
});

// Automatically process a provided File (e.g., via drag-drop or external paste)
watch(() => props.pastedFile, async (f) => {
  if (f) {
    if (mode.value !== "file") mode.value = "file";
    await useFile(f);
  }
});

defineExpose({ onFiles, switchToFileMode, localError });
</script>

<style scoped>
/* Existing */
.file-text-capture.is-disabled {
  opacity: 0.6;
  pointer-events: none;
}

/* NEW: better wrapping so long strings don't overflow */
.ftc-title,
.ftc-sub,
.ftc-desc {
  word-break: break-word;
  overflow-wrap: anywhere;
}

/* Tighten paddings and minimum height on phones */
@media (max-width: 600px) {
  /* v-file-upload root inside this component */
  .file-text-capture :deep(.v-file-upload) {
    padding: 8px 10px;
  }

  /* Dashed box look (if you set a dashed border via class) */
  .file-text-capture :deep(.v-file-upload .v-input__control) {
    min-height: 120px; /* smaller on mobile so it doesn't dominate */
  }

  /* Text sizes a bit smaller to prevent wrapping hell */
  .file-text-capture .ftc-title {
    font-size: 0.95rem;
    line-height: 1.2rem;
  }
  .file-text-capture .ftc-sub {
    font-size: 0.82rem;
    line-height: 1.1rem;
  }
  .file-text-capture .ftc-desc {
    font-size: 0.78rem;
    line-height: 1.05rem;
  }
}

</style>
