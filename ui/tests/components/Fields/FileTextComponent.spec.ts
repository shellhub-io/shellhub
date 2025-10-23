import { setActivePinia, createPinia } from "pinia";
import { mount, flushPromises } from "@vue/test-utils";
import { describe, it, expect, beforeEach } from "vitest";
import { createVuetify } from "vuetify";
import * as components from "vuetify/components";
import * as directives from "vuetify/directives";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";

type FRHandler = (() => void) | null;
type GlobalWithFR = typeof globalThis & { FileReader: typeof FileReader };

const mockFileReader = (text: string, triggerError = false) => {
  const OriginalFR = (globalThis as GlobalWithFR).FileReader;

  class MockFileReader {
    public onload: FRHandler = null;

    public onerror: FRHandler = null;

    public result: string | null = null;

    readAsText() {
      if (triggerError) {
        this.onerror?.();
      } else {
        this.result = text;
        this.onload?.();
      }
    }
  }

  (globalThis as GlobalWithFR).FileReader = MockFileReader as unknown as typeof FileReader;

  return () => {
    (globalThis as GlobalWithFR).FileReader = OriginalFR;
  };
};

describe("FileTextComponent", () => {
  const mountCmp = (props: Record<string, unknown> = {}) => mount(FileTextComponent, {
    props: {
      modelValue: "",
      ...props,
    },
    global: {
      plugins: [
        createVuetify({ components, directives }),
        createPinia(),
      ],
      stubs: {
        "v-file-upload": true,
        "v-file-upload-item": true,
      },
    },
  });

  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("renders in file mode by default", async () => {
    const wrapper = mountCmp();
    await flushPromises();
    expect(wrapper.find('[data-test="file-text-capture"]').exists()).toBe(true);
    expect(wrapper.find('[data-test="ftc-textarea"]').exists()).toBe(false);
  });

  it("starts in text mode if startInText is true", async () => {
    const wrapper = mountCmp({ startInText: true });
    await flushPromises();
    expect(wrapper.find('[data-test="ftc-textarea"]').exists()).toBe(true);
  });

  it("only shows textarea if textOnly is true", async () => {
    const wrapper = mountCmp({ textOnly: true });
    await flushPromises();
    expect(wrapper.find('[data-test="file-text-capture"]').exists()).toBe(false);
    expect(wrapper.find('[data-test="ftc-textarea"]').exists()).toBe(true);
  });

  it("validates file content with validator", async () => {
    const restore = mockFileReader("BAD CONTENT");
    const file = new File(["BAD CONTENT"], "bad.pub", { type: "text/plain" });
    const wrapper = mountCmp({
      validator: (t: string) => t.startsWith("GOOD"),
      invalidMessage: "Not valid",
    });

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    expect(wrapper.vm.errorMessage).toEqual("Not valid");
    restore();
  });

  it("accepts valid file and emits modelValue", async () => {
    const restore = mockFileReader("GOOD KEY");
    const file = new File(["GOOD KEY"], "good.pub", { type: "text/plain" });
    const wrapper = mountCmp({
      validator: (t: string) => t.startsWith("GOOD"),
    });

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    const updates = wrapper.emitted("update:modelValue") ?? [];
    expect(updates[updates.length - 1]).toEqual(["GOOD KEY"]);
    expect(wrapper.emitted("file-name")?.[0]).toEqual(["good.pub"]);
    expect(wrapper.emitted("file-processed")).toHaveLength(1);
    restore();
  });

  it("clears modelValue and errors when cleared", async () => {
    const restore = mockFileReader("WRONG");
    const file = new File(["WRONG"], "bad.pub", { type: "text/plain" });
    const wrapper = mountCmp({
      validator: () => false,
      invalidMessage: "Nope",
    });

    await wrapper.vm.onFiles([file]);
    expect(wrapper.vm.errorMessage).toBe("Nope");
    await flushPromises();

    await wrapper.vm.onFiles(null);
    await flushPromises();

    const updates = wrapper.emitted("update:modelValue") ?? [];
    expect(updates[updates.length - 1]).toEqual([""]);
    restore();
  });

  it("displays error when FileReader fails", async () => {
    const restore = mockFileReader("", true);
    const file = new File(["X"], "bad.pub", { type: "text/plain" });
    const wrapper = mountCmp();

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("Could not read the file.");
    restore();
  });

  it("validates text input in text mode", async () => {
    const wrapper = mountCmp({
      startInText: true,
      validator: (t: string) => t === "VALID",
      invalidMessage: "Bad",
    });
    await flushPromises();

    const textarea = wrapper.find("textarea");
    expect(textarea.exists()).toBe(true);

    await textarea.setValue("INVALID");
    await flushPromises();
    expect(wrapper.vm.errorMessage).toBe("Bad");

    await textarea.setValue("VALID");
    await flushPromises();
    expect(wrapper.vm.errorMessage).toBe("");
  });

  it("switches mode and reuses existing file on button click", async () => {
    const restore = mockFileReader("SSH KEY");
    const file = new File(["SSH KEY"], "file.pub", { type: "text/plain" });
    const wrapper = mountCmp({ startInText: true });

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    await wrapper.vm.switchToFileMode();
    await flushPromises();

    expect(wrapper.emitted("mode-changed")?.[0]).toEqual(["file"]);
    expect(wrapper.emitted("file-name")?.[0]).toEqual(["file.pub"]);
    restore();
  });

  it("accepts extensionless file when allowExtensionless is true and MIME is octet-stream", async () => {
    const restore = mockFileReader("GOOD KEY");
    const file = new File(["GOOD KEY"], "id_ed25519", { type: "application/octet-stream" });
    const wrapper = mountCmp({
      allowExtensionless: true,
      validator: (t: string) => t.startsWith("GOOD"),
    });

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    const updates = wrapper.emitted("update:modelValue") ?? [];
    expect(updates[updates.length - 1]).toEqual(["GOOD KEY"]);
    expect(wrapper.emitted("file-name")?.[0]).toEqual(["id_ed25519"]);
    expect(wrapper.emitted("file-processed")).toHaveLength(1);
    restore();
  });

  it("accepts canonical ssh-keygen names regardless of extension when validator passes", async () => {
    const restore = mockFileReader("GOOD KEY");
    const file = new File(["GOOD KEY"], "id_rsa", { type: "" });
    const wrapper = mountCmp({
      allowExtensionless: false,
      validator: (t: string) => t === "GOOD KEY",
    });

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    const updates = wrapper.emitted("update:modelValue") ?? [];
    expect(updates[updates.length - 1]).toEqual(["GOOD KEY"]);
    restore();
  });

  it("rejects random extensionless file when allowExtensionless is false", async () => {
    const restore = mockFileReader("WHATEVER");
    const file = new File(["WHATEVER"], "randomfile", { type: "" });
    const wrapper = mountCmp({ allowExtensionless: false });

    await wrapper.vm.onFiles([file]);
    await flushPromises();

    expect(wrapper.vm.errorMessage).toBe("Unsupported file type.");
    restore();
  });
});
