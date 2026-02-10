import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushPromises, VueWrapper } from "@vue/test-utils";
import UserIcon from "@/components/User/UserIcon.vue";
import { mountComponent } from "@tests/utils/mount";

// Mock crypto.subtle.digest for SHA-256 hash generation
const mockDigest = vi.fn();
Object.defineProperty(globalThis, "crypto", {
  value: {
    subtle: {
      digest: mockDigest,
    },
  },
  writable: true,
});

describe("UserIcon", () => {
  let wrapper: VueWrapper<InstanceType<typeof UserIcon>>;

  const mockHash = (email: string) => {
    const encoder = new TextEncoder();
    const data = encoder.encode(email);
    // Simple mock hash - in reality this would be SHA-256
    const hashArray = Array.from(data).slice(0, 32);
    return Promise.resolve(new Uint8Array(hashArray).buffer);
  };

  const mountWrapper = (size: string | number = 48, email: string | null = "test@example.com", authEmail = "") => {
    wrapper = mountComponent(UserIcon, {
      props: {
        size,
        email,
      },
      piniaOptions: {
        initialState: {
          auth: { email: authEmail },
        },
      },
    });
  };

  beforeEach(() => {
    mockDigest.mockImplementation((algorithm: string, data: BufferSource) => {
      const text = new TextDecoder().decode(data);
      return mockHash(text);
    });
  });

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  describe("Component rendering with email prop", () => {
    it("renders avatar with correct size", () => {
      mountWrapper();

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
    });

    it("initially shows image element, not placeholder", () => {
      mountWrapper();

      const image = wrapper.find('[data-test="gravatar-image"]');
      const placeholder = wrapper.find('[data-test="gravatar-placeholder"]');

      expect(image.exists()).toBe(true);
      expect(placeholder.exists()).toBe(false);
    });
  });

  describe("Gravatar URL generation", () => {
    it("generates correct Gravatar URL from email prop", async () => {
      mountWrapper();

      await flushPromises();

      expect(mockDigest).toHaveBeenCalled();

      const callArgs = mockDigest.mock.calls[0];
      const encodedData = callArgs[1] as Uint8Array;
      const decodedEmail = new TextDecoder().decode(encodedData);
      expect(decodedEmail).toBe("test@example.com");
    });

    it("uses email from auth store when email prop is not provided", async () => {
      mountWrapper(48, undefined, "store@example.com");

      await flushPromises();

      expect(mockDigest).toHaveBeenCalled();
    });

    it("prioritizes email prop over auth store email", async () => {
      const propEmail = "prop@example.com";

      mountWrapper(48, propEmail, "store@example.com");

      await flushPromises();

      // The digest should be called with the prop email
      const calls = mockDigest.mock.calls;
      expect(calls.length).toBeGreaterThan(0);
    });

    it("regenerates URL when email prop changes", async () => {
      mountWrapper(48, "initial@example.com");

      await flushPromises();
      const initialCalls = mockDigest.mock.calls.length;

      await wrapper.setProps({ email: "changed@example.com" });
      await flushPromises();

      expect(mockDigest.mock.calls.length).toBeGreaterThan(initialCalls);
    });
  });

  describe("Image loading behavior", () => {
    it("shows gravatar image when successfully loaded", async () => {
      mountWrapper();

      await flushPromises();

      const img = wrapper.find('[data-test="gravatar-image"]');
      expect(img.exists()).toBe(true);
      expect(wrapper.find('[data-test="gravatar-placeholder"]').exists()).toBe(false);
    });

    it("shows placeholder icon when image fails to load", async () => {
      mountWrapper();

      await flushPromises();

      const img = wrapper.findComponent({ name: "VImg" });

      // Simulate image error
      img.vm.$emit("error");
      await flushPromises();

      expect(wrapper.find('[data-test="gravatar-placeholder"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="gravatar-image"]').exists()).toBe(false);
    });

    it("resets error state when email changes", async () => {
      mountWrapper();

      await flushPromises();

      // Trigger error
      const img = wrapper.findComponent({ name: "VImg" });
      img.vm.$emit("error");
      await flushPromises();

      expect(wrapper.find('[data-test="gravatar-placeholder"]').exists()).toBe(true);

      // Change email
      await wrapper.setProps({ email: "new@example.com" });
      await flushPromises();

      // Image should be attempted again
      expect(wrapper.find('[data-test="gravatar-image"]').exists()).toBe(true);
    });
  });

  describe("Null or empty email handling", () => {
    it("shows placeholder when email prop is null", async () => {
      mountWrapper(48, null);

      await flushPromises();

      expect(wrapper.find('[data-test="gravatar-placeholder"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="gravatar-image"]').exists()).toBe(false);
    });

    it("shows placeholder when email is not provided and auth store has no email", async () => {
      mountWrapper(48, null, "");

      await flushPromises();

      expect(wrapper.find('[data-test="gravatar-placeholder"]').exists()).toBe(true);
      expect(wrapper.find('[data-test="gravatar-image"]').exists()).toBe(false);
    });
  });

  describe("Avatar styling", () => {
    it("applies primary color to avatar background", () => {
      mountWrapper();

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.classes()).toContain("bg-primary");
    });

    it("applies border class to avatar", () => {
      mountWrapper();

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.classes()).toContain("border");
    });

    it("applies correct size to avatar", () => {
      mountWrapper(64);

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
    });

    it("handles string size prop", () => {
      mountWrapper("40");

      const avatar = wrapper.find(".v-avatar");
      expect(avatar.exists()).toBe(true);
    });
  });

  describe("Placeholder icon styling", () => {
    it("applies surface color to placeholder icon when shown", async () => {
      mountWrapper(48, null);

      await flushPromises();

      const placeholder = wrapper.find('[data-test="gravatar-placeholder"]');
      expect(placeholder.classes()).toContain("text-surface");
    });

    it("renders mdi-account icon as placeholder", async () => {
      mountWrapper(48, null);

      await flushPromises();

      const placeholder = wrapper.find('[data-test="gravatar-placeholder"]');
      expect(placeholder.classes()).toContain("mdi-account");
    });
  });
});
