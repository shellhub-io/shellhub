import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper, DOMWrapper } from "@vue/test-utils";
import axios from "axios";
import TerminalDialog from "@/components/Terminal/TerminalDialog.vue";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";
import { mountComponent } from "@tests/utils/mount";
import { createCleanRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";

vi.mock("@/utils/sshKeys", () => ({
  convertToFingerprint: vi.fn(() => "mock-fingerprint"),
  generateSignature: vi.fn(() => "mock-signature"),
}));

vi.mock("@xterm/xterm", () => ({
  Terminal: vi.fn().mockImplementation(() => ({
    open: vi.fn(),
    focus: vi.fn(),
    write: vi.fn(),
    onData: vi.fn(),
    onResize: vi.fn(),
    loadAddon: vi.fn(),
    cols: 80,
    rows: 24,
    options: {},
  })),
}));

vi.mock("@xterm/addon-fit", () => ({
  FitAddon: vi.fn().mockImplementation(() => ({
    fit: vi.fn(),
  })),
}));

describe("TerminalDialog", () => {
  let wrapper: VueWrapper<InstanceType<typeof TerminalDialog>>;
  let dialog: DOMWrapper<Element>;

  const router = createCleanRouter();

  const mountWrapper = (sshid?: string) => {
    wrapper = mountComponent(TerminalDialog, {
      props: {
        modelValue: true,
        deviceUid: "a582b47a42d",
        deviceName: "test-device",
        sshid,
      },
      global: {
        plugins: [router],
        stubs: {
          TerminalThemeDrawer: {
            template: "<div data-test='theme-drawer-stub' />",
            props: ["modelValue", "showDrawer"],
            emits: ["update:selectedTheme", "update:fontSettings"],
          },
        },
      },
      attachTo: document.body,
    });

    dialog = new DOMWrapper(document.body);
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering", () => {
    it("renders TerminalLoginForm initially", () => {
      expect(wrapper.findComponent({ name: "TerminalLoginForm" }).exists()).toBe(true);
      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(false);
    });

    it("passes correct props to TerminalLoginForm", () => {
      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      expect(loginForm.props("modelValue")).toBe(true);
      expect(loginForm.props("loading")).toBe(false);
    });

    it("passes sshid to TerminalLoginForm when provided", () => {
      wrapper.unmount();
      document.body.innerHTML = "";

      mountWrapper("test-sshid");

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      expect(loginForm.props("sshid")).toBe("test-sshid");
    });
  });

  describe("Login and terminal switching", () => {
    it("shows login form when showLoginForm is true", () => {
      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(true);
      expect(dialog.find('[data-test="terminal-container"]').exists()).toBe(false);
    });

    it("shows terminal when showLoginForm is false", async () => {
      const mockResponse = { data: { token: "fake-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      expect(dialog.find('[data-test="terminal-container"]').exists()).toBe(true);
      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(false);
    });

    it("passes token to Terminal component", async () => {
      const mockResponse = { data: { token: "test-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      const terminal = wrapper.findComponent({ name: "Terminal" });
      expect(terminal.props("token")).toBe("test-token");
    });

    it("passes deviceName to Terminal component", async () => {
      const mockResponse = { data: { token: "fake-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      const terminal = wrapper.findComponent({ name: "Terminal" });
      expect(terminal.props("deviceName")).toBe("test-device");
    });
  });

  describe("Password authentication", () => {
    it("connects with password successfully", async () => {
      const mockResponse = { data: { token: "fake-token" } };
      const postSpy = vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      expect(postSpy).toHaveBeenCalledWith("/ws/ssh", {
        device: "a582b47a42d",
        username: "test-user",
        password: "test-pass",
        authenticationMethod: TerminalAuthMethods.Password,
      });

      const terminal = wrapper.findComponent({ name: "Terminal" });
      expect(terminal.exists()).toBe(true);
      expect(terminal.props("token")).toBe("fake-token");
    });

    it("handles connection error with 401 status", async () => {
      const error = createAxiosError(401, "Unauthorized");
      vi.spyOn(axios, "post").mockRejectedValue(error);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "wrong-pass",
      });
      await flushPromises();

      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(true);
      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(false);
      expect(loginForm.props("loading")).toBe(false);
    });

    it("handles connection error with 403 status", async () => {
      const error = createAxiosError(403, "Forbidden");
      vi.spyOn(axios, "post").mockRejectedValue(error);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(true);
      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(false);
      expect(loginForm.props("loading")).toBe(false);
    });

    it("handles generic connection error", async () => {
      const error = createAxiosError(500, "Internal Server Error");
      vi.spyOn(axios, "post").mockRejectedValue(error);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(true);
      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(false);
      expect(loginForm.props("loading")).toBe(false);
    });
  });

  describe("Private key authentication", () => {
    it("connects with private key successfully", async () => {
      const { convertToFingerprint } = await import("@/utils/sshKeys");
      const mockResponse = { data: { token: "fake-token" } };
      const postSpy = vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.PrivateKey,
        username: "test-user",
        privateKey: "test-private-key",
        passphrase: "test-passphrase",
      });
      await flushPromises();

      expect(convertToFingerprint).toHaveBeenCalledWith("test-private-key", "test-passphrase");
      expect(postSpy).toHaveBeenCalledWith("/ws/ssh", {
        device: "a582b47a42d",
        username: "test-user",
        fingerprint: "mock-fingerprint",
      });

      const terminal = wrapper.findComponent({ name: "Terminal" });
      expect(terminal.exists()).toBe(true);
      expect(terminal.props("token")).toBe("fake-token");
    });

    it("passes private key to Terminal component", async () => {
      const mockResponse = { data: { token: "fake-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.PrivateKey,
        username: "test-user",
        privateKey: "test-private-key",
        passphrase: "test-passphrase",
      });
      await flushPromises();

      const terminal = wrapper.findComponent({ name: "Terminal" });
      expect(terminal.props("privateKey")).toBe("test-private-key");
      expect(terminal.props("passphrase")).toBe("test-passphrase");
    });

    it("handles private key connection error", async () => {
      const error = createAxiosError(401, "Authentication failed");
      vi.spyOn(axios, "post").mockRejectedValue(error);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.PrivateKey,
        username: "test-user",
        privateKey: "test-private-key",
      });
      await flushPromises();

      // Opens terminal with error message
      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(false);
      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(true);
    });
  });

  describe("Dialog close behavior", () => {
    it("closes dialog when TerminalLoginForm emits close", async () => {
      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("closes dialog when Terminal emits close", async () => {
      const mockResponse = { data: { token: "fake-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      const terminal = wrapper.findComponent({ name: "Terminal" });
      terminal.vm.$emit("close");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("resets state when dialog is closed", async () => {
      const mockResponse = { data: { token: "test-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.PrivateKey,
        username: "test-user",
        privateKey: "test-key",
        password: "test-pass",
      });
      await flushPromises();

      const terminal = wrapper.findComponent({ name: "Terminal" });

      // Terminal is shown
      expect(terminal.exists()).toBe(true);

      // Close the terminal
      terminal.vm.$emit("close");
      await flushPromises();

      // Unmount and remount to simulate reopening
      wrapper.unmount();
      document.body.innerHTML = "";
      mountWrapper();

      // Should show login form again (reset state)
      expect(dialog.find('[data-test="terminal-login-form"]').exists()).toBe(true);
      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(false);
    });
  });

  describe("Computed properties", () => {
    it("syncs showLoginDialog with showDialog and showLoginForm", async () => {
      expect(wrapper.findComponent({ name: "TerminalLoginForm" }).props("modelValue")).toBe(true);

      // Login successfully to hide login form
      const mockResponse = { data: { token: "fake-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      expect(wrapper.findComponent({ name: "TerminalLoginForm" }).exists()).toBe(false);
    });

    it("syncs showTerminalDialog with showDialog and showLoginForm", async () => {
      const mockResponse = { data: { token: "fake-token" } };
      vi.spyOn(axios, "post").mockResolvedValue(mockResponse);

      const loginForm = wrapper.findComponent({ name: "TerminalLoginForm" });
      loginForm.vm.$emit("submit", {
        authenticationMethod: TerminalAuthMethods.Password,
        username: "test-user",
        password: "test-pass",
      });
      await flushPromises();

      expect(wrapper.findComponent({ name: "Terminal" }).props("modelValue")).toBe(true);

      // Close terminal to show login form again
      const terminal = wrapper.findComponent({ name: "Terminal" });
      terminal.vm.$emit("close");
      await flushPromises();

      await wrapper.setProps({ modelValue: true });
      await flushPromises();

      expect(wrapper.findComponent({ name: "Terminal" }).exists()).toBe(false);
    });
  });
});
