import { describe, it, expect, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
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
  let wrapper: VueWrapper<InstanceType<typeof FileTextComponent>>;

  const mountWrapper = async (props: Partial<InstanceType<typeof FileTextComponent>["$props"]> = {}) => {
    wrapper = mountComponent(FileTextComponent, {
      props: {
        modelValue: "",
        ...props,
      },
      global: { stubs: ["VFileUpload", "VFileUploadItem"] },
    });
    await flushPromises();
  };

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("Display modes", () => {
    it("Renders in file mode by default", async () => {
      await mountWrapper();

      expect(wrapper.find('[data-test="file-text-capture"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="ftc-textarea"]').exists()).toBe(false);
    });

    it("Starts in text mode when startInText is true", async () => {
      await mountWrapper({ startInText: true });

      expect(wrapper.find('[data-test="ftc-textarea"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="file-text-capture"]').exists()).toBe(false);
    });

    it("Only shows textarea when textOnly is true", async () => {
      await mountWrapper({ textOnly: true });

      expect(wrapper.find('[data-test="file-text-capture"]').exists()).toBe(false);
      expect(wrapper.find('[data-test="ftc-textarea"]').exists()).toBe(true);
    });
  });

  describe("File upload", () => {
    it("Accepts valid file and emits modelValue", async () => {
      const restore = mockFileReader("GOOD KEY");
      const file = new File(["GOOD KEY"], "good.pub", { type: "text/plain" });

      await mountWrapper({
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

    it("Validates file content with validator", async () => {
      const restore = mockFileReader("BAD CONTENT");
      const file = new File(["BAD CONTENT"], "bad.pub", { type: "text/plain" });

      await mountWrapper({
        validator: (t: string) => t.startsWith("GOOD"),
        invalidMessage: "Not valid",
      });

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      expect(wrapper.vm.errorMessage).toEqual("Not valid");

      restore();
    });

    it("Shows error message when file validation fails", async () => {
      const restore = mockFileReader("INVALID");
      const file = new File(["INVALID"], "bad.pub", { type: "text/plain" });

      await mountWrapper({
        validator: () => false,
        invalidMessage: "Invalid file content",
      });

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      const errorEl = wrapper.find('[data-test="ftc-file-error"]');
      expect(errorEl.exists()).toBe(true);
      expect(errorEl.text()).toBe("Invalid file content");

      restore();
    });

    it("Clears modelValue and errors when file is cleared", async () => {
      const restore = mockFileReader("WRONG");
      const file = new File(["WRONG"], "bad.pub", { type: "text/plain" });

      await mountWrapper({
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

    it("Displays error when FileReader fails", async () => {
      const restore = mockFileReader("", true);
      const file = new File(["X"], "bad.pub", { type: "text/plain" });

      await mountWrapper();

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      expect(wrapper.vm.errorMessage).toBe("Could not read the file.");

      restore();
    });
  });

  describe("Text input mode", () => {
    it("Renders textarea with correct label", async () => {
      await mountWrapper({
        startInText: true,
        textareaLabel: "SSH Key Content",
      });

      const textarea = wrapper.find('[data-test="ftc-textarea"]');
      expect(textarea.exists()).toBe(true);
      expect(textarea.text()).toContain("SSH Key Content");
    });

    it("Updates modelValue when text is entered", async () => {
      await mountWrapper({ startInText: true });

      const textarea = wrapper.find("textarea");
      await textarea.setValue("ssh-rsa AAAAB3...");
      await flushPromises();

      const updates = wrapper.emitted("update:modelValue") ?? [];
      expect(updates[updates.length - 1]).toEqual(["ssh-rsa AAAAB3..."]);
    });

    it("Validates text input with validator", async () => {
      await mountWrapper({
        startInText: true,
        validator: (t: string) => t === "VALID",
        invalidMessage: "Bad",
      });

      const textarea = wrapper.find("textarea");
      await textarea.setValue("INVALID");
      await flushPromises();

      expect(wrapper.vm.errorMessage).toBe("Bad");

      await textarea.setValue("VALID");
      await flushPromises();

      expect(wrapper.vm.errorMessage).toBe("");
    });

    it("Shows error message in textarea when validation fails", async () => {
      await mountWrapper({
        startInText: true,
        validator: () => false,
        invalidMessage: "Invalid text",
      });

      const textarea = wrapper.find("textarea");
      await textarea.setValue("some text");
      await flushPromises();

      const textareaWrapper = wrapper.find('[data-test="ftc-textarea"]');
      expect(textareaWrapper.text()).toContain("Invalid text");
    });

    it("Shows upload button in textarea mode", async () => {
      await mountWrapper({ startInText: true });

      const uploadBtn = wrapper.find('[data-test="ftc-textarea"] button[title="Return to file drop zone"]');

      expect(uploadBtn.exists()).toBe(true);
      expect(uploadBtn.find("i").classes()).toContain("mdi-upload");
    });
  });

  describe("Extensionless files", () => {
    it("Accepts extensionless file when allowExtensionless is true", async () => {
      const restore = mockFileReader("GOOD KEY");
      const file = new File(["GOOD KEY"], "id_ed25519", { type: "application/octet-stream" });

      await mountWrapper({
        allowExtensionless: true,
        validator: (t: string) => t.startsWith("GOOD"),
      });

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      const updates = wrapper.emitted("update:modelValue") ?? [];
      expect(updates[updates.length - 1]).toEqual(["GOOD KEY"]);
      expect(wrapper.emitted("file-name")?.[0]).toEqual(["id_ed25519"]);

      restore();
    });

    it("Accepts canonical ssh-keygen names when validator passes", async () => {
      const restore = mockFileReader("GOOD KEY");
      const file = new File(["GOOD KEY"], "id_rsa", { type: "" });

      await mountWrapper({
        allowExtensionless: false,
        validator: (t: string) => t === "GOOD KEY",
      });

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      const updates = wrapper.emitted("update:modelValue") ?? [];
      expect(updates[updates.length - 1]).toEqual(["GOOD KEY"]);

      restore();
    });

    it("Rejects random extensionless file when allowExtensionless is false", async () => {
      const restore = mockFileReader("WHATEVER");
      const file = new File(["WHATEVER"], "randomfile", { type: "" });

      await mountWrapper({ allowExtensionless: false });

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      expect(wrapper.vm.errorMessage).toBe("Unsupported file type.");

      restore();
    });
  });

  describe("Mode switching", () => {
    it("Switches mode and emits mode-changed event", async () => {
      const restore = mockFileReader("SSH KEY");
      const file = new File(["SSH KEY"], "file.pub", { type: "text/plain" });

      await mountWrapper({ startInText: true });

      await wrapper.vm.onFiles([file]);
      await flushPromises();

      await wrapper.vm.switchToFileMode();
      await flushPromises();

      expect(wrapper.emitted("mode-changed")?.[0]).toEqual(["file"]);
      expect(wrapper.emitted("file-name")?.[0]).toEqual(["file.pub"]);

      restore();
    });
  });

  describe("Disabled state", () => {
    it("Disables file upload when disabled is true", async () => {
      await mountWrapper({ disabled: true });

      const fileUpload = wrapper.findComponent({ name: "VFileUpload" });
      expect(fileUpload.attributes("disabled")).toBeDefined();
    });

    it("Disables textarea when disabled is true", async () => {
      await mountWrapper({
        startInText: true,
        disabled: true,
      });

      const textarea = wrapper.find("textarea");
      expect(textarea.attributes("disabled")).toBeDefined();
    });
  });
});
