import { describe, expect, it, beforeEach, vi, afterEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mountComponent, mockSnackbar } from "@tests/utils/mount";
import { createCleanAdminRouter } from "@tests/utils/router";
import { createAxiosError } from "@tests/utils/axiosError";
import useAnnouncementStore from "@admin/store/modules/announcement";
import NewAnnouncement from "@admin/views/NewAnnouncement.vue";

vi.mock("@admin/store/api/announcement");

// Mock TinyMCE Editor to make it testable
vi.mock("@tinymce/tinymce-vue", () => ({
  default: {
    name: "Editor",
    // eslint-disable-next-line vue/max-len
    template: "<textarea :value=\"modelValue\" @input=\"$emit('update:modelValue', $event.target.value)\" data-test=\"announcement-content\"></textarea>",
    props: ["modelValue", "apiKey", "init", "toolbar", "outputFormat"],
  },
}));

describe("NewAnnouncement", () => {
  let wrapper: VueWrapper<InstanceType<typeof NewAnnouncement>>;
  let announcementStore: ReturnType<typeof useAnnouncementStore>;

  const mountWrapper = async () => {
    const router = createCleanAdminRouter();
    await router.push({ name: "new-announcement" });
    await router.isReady();

    wrapper = mountComponent(NewAnnouncement, { global: { plugins: [router] } });

    announcementStore = useAnnouncementStore();
    await flushPromises();
  };

  beforeEach(() => mountWrapper());

  afterEach(() => {
    vi.clearAllMocks();
    wrapper?.unmount();
  });

  it("displays the page header with correct title", () => {
    expect(wrapper.text()).toContain("Create new Announcement");
    expect(wrapper.text()).toContain("Platform Messaging");
  });

  it("displays the page header description", () => {
    expect(wrapper.text()).toContain("Compose a system-wide update to keep every namespace informed about critical changes.");
  });

  it("displays the title input field", () => {
    const titleInput = wrapper.find('[data-test="announcement-title-field"] input');
    expect(titleInput.exists()).toBe(true);
    expect(wrapper.text()).toContain("Title");
  });

  it("displays the content editor", () => {
    expect(wrapper.find('[data-test="announcement-content"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Content");
  });

  it("displays the post button", () => {
    const postBtn = wrapper.find('[data-test="announcement-btn-post"]');
    expect(postBtn.exists()).toBe(true);
    expect(postBtn.text()).toBe("Post");
  });

  describe("when posting announcement", () => {
    it("shows error when title is empty", async () => {
      const postBtn = wrapper.find('[data-test="announcement-btn-post"]');
      await postBtn.trigger("click");
      await flushPromises();

      const titleInput = wrapper.find('[data-test="announcement-title-field"]');
      expect(titleInput.text()).toContain("Title cannot be empty!");
    });

    it("shows error when content is empty", async () => {
      const titleInput = wrapper.find('[data-test="announcement-title-field"] input');
      await titleInput.setValue("Test Title");
      await flushPromises();

      const postBtn = wrapper.find('[data-test="announcement-btn-post"]');
      await postBtn.trigger("click");
      await flushPromises();

      const errorAlert = wrapper.find('[data-test="announcement-error"]');
      expect(errorAlert.exists()).toBe(true);
      expect(errorAlert.text()).toContain("The announcement cannot be empty!");
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create announcement.");
    });

    it("creates announcement successfully with valid data", async () => {
      const titleInput = wrapper.find('[data-test="announcement-title-field"] input');
      await titleInput.setValue("Important Update");
      await flushPromises();

      const contentEditor = wrapper.find('[data-test="announcement-content"]');
      await contentEditor.setValue("<p>This is the announcement content</p>");
      await contentEditor.trigger("input");
      await flushPromises();

      vi.mocked(announcementStore.createAnnouncement).mockResolvedValueOnce(undefined);

      const postBtn = wrapper.find('[data-test="announcement-btn-post"]');
      await postBtn.trigger("click");
      await flushPromises();

      expect(announcementStore.createAnnouncement).toHaveBeenCalledWith(
        expect.objectContaining({
          title: "Important Update",
          content: expect.any(String),
        }),
      );
      expect(mockSnackbar.showSuccess).toHaveBeenCalledWith("Successfully created announcement.");
    });

    it("shows error when API call fails", async () => {
      const titleInput = wrapper.find('[data-test="announcement-title-field"] input');
      await titleInput.setValue("Failed Announcement");
      await flushPromises();

      const contentEditor = wrapper.find('[data-test="announcement-content"]');
      await contentEditor.setValue("<p>Content</p>");
      await contentEditor.trigger("input");
      await flushPromises();

      vi.mocked(announcementStore.createAnnouncement).mockRejectedValueOnce(
        createAxiosError(500, "Server Error"),
      );

      const postBtn = wrapper.find('[data-test="announcement-btn-post"]');
      await postBtn.trigger("click");
      await flushPromises();

      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to create announcement.");
    });
  });

  describe("when content changes", () => {
    it("clears announcement error when content is added", async () => {
      const titleInput = wrapper.find('[data-test="announcement-title-field"] input');
      await titleInput.setValue("Test Title");
      await flushPromises();

      // Trigger error first by posting without content
      const postBtn = wrapper.find('[data-test="announcement-btn-post"]');
      await postBtn.trigger("click");
      await flushPromises();

      let errorAlert = wrapper.find('[data-test="announcement-error"]');
      expect(errorAlert.exists()).toBe(true);

      // Add content to clear the error
      const contentEditor = wrapper.find('[data-test="announcement-content"]');
      await contentEditor.setValue("<p>New content</p>");
      await contentEditor.trigger("input");
      await flushPromises();

      errorAlert = wrapper.find('[data-test="announcement-error"]');
      expect(errorAlert.exists()).toBe(false);
    });
  });
});
