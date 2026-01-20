import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import { DOMWrapper, flushPromises, VueWrapper } from "@vue/test-utils";
import { mountComponent } from "@tests/utils/mount";
import AnnouncementsModal from "@/components/Announcements/AnnouncementsModal.vue";
import type { IAnnouncement } from "@/interfaces/IAnnouncement";

const mockAnnouncement: IAnnouncement = {
  uuid: "announcement-123",
  title: "New Features Available",
  content: "**Hello** _world_!\n\nThis is a test announcement with:\n- Feature 1\n- Feature 2\n\n> Important note here",
  date: "2026-01-20T12:00:00Z",
};

const mockSimpleAnnouncement: IAnnouncement = {
  uuid: "announcement-456",
  title: "Simple Title",
  content: "Plain text content without markdown",
  date: "2025-12-15T08:30:00Z",
};

describe("AnnouncementsModal", () => {
  let wrapper: VueWrapper<InstanceType<typeof AnnouncementsModal>>;

  const getDialog = () => new DOMWrapper(document.body).find("[role='dialog']");

  const mountWrapper = (announcement = mockAnnouncement, modelValue = true) => {
    wrapper = mountComponent(AnnouncementsModal, {
      props: {
        announcement,
        modelValue,
      },
      attachTo: document.body,
    });
  };

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  afterEach(() => {
    wrapper?.unmount();
    document.body.innerHTML = "";
  });

  describe("rendering", () => {
    beforeEach(() => mountWrapper());

    it("renders the dialog when modelValue is true", () => {
      const dialog = getDialog();
      expect(dialog.exists()).toBe(true);
    });

    it("displays the announcement title in dialog titlebar", () => {
      const dialog = getDialog();
      const titleElement = dialog.find("[data-test='window-dialog-titlebar']");
      expect(titleElement.text()).toContain(mockAnnouncement.title);
    });

    it("renders markdown content as HTML", () => {
      const dialog = getDialog();
      const content = dialog.find('[data-test="announcement-title"]');
      expect(content.html()).toContain("<strong>Hello</strong>");
      expect(content.html()).toContain("<em>world</em>");
      expect(content.html()).toContain("<ul>");
      expect(content.html()).toContain("<li>Feature 1</li>");
      expect(content.html()).toContain("<blockquote>");
    });

    it("displays formatted date in readable format", () => {
      const dialog = getDialog();
      const dateElement = dialog.find('[data-test="announcement-date"]');
      expect(dateElement.text()).toContain("Posted in January 20, 2026");
    });

    it("shows the primary icon in dialog header", () => {
      const dialog = getDialog();
      const icon = dialog.find(".v-icon");
      expect(icon.classes()).toContain("mdi-bullhorn");
    });

    it("displays the Dismiss button in footer", () => {
      const dialog = getDialog();
      const btn = dialog.find('[data-test="announcement-close"]');
      expect(btn.exists()).toBe(true);
      expect(btn.text()).toBe("Dismiss");
    });
  });

  describe("markdown rendering", () => {
    it("renders plain text announcement without HTML tags", () => {
      mountWrapper(mockSimpleAnnouncement);
      const dialog = getDialog();
      const content = dialog.find('[data-test="announcement-title"]');
      expect(content.text()).toContain("Plain text content without markdown");
      expect(content.html()).not.toContain("<strong>");
      expect(content.html()).not.toContain("<em>");
    });

    it("preserves markdown formatting for complex content", () => {
      mountWrapper();
      const dialog = getDialog();
      const content = dialog.find('[data-test="announcement-title"]');
      expect(content.html()).toContain("<li>Feature 1</li>");
      expect(content.html()).toContain("<li>Feature 2</li>");
      expect(content.html()).toContain("<blockquote>");
    });
  });

  describe("date formatting", () => {
    it("formats different dates correctly", () => {
      mountWrapper(mockSimpleAnnouncement);
      const dialog = getDialog();
      const dateElement = dialog.find('[data-test="announcement-date"]');
      expect(dateElement.text()).toContain("Posted in December 15, 2025");
    });
  });

  describe("user interactions", () => {
    beforeEach(() => mountWrapper());

    it("closes the dialog when Dismiss button is clicked", async () => {
      const dialog = getDialog();
      const btn = dialog.find('[data-test="announcement-close"]');

      await btn.trigger("click");
      await flushPromises();

      expect(wrapper.emitted("update:modelValue")).toBeTruthy();
      expect(wrapper.emitted("update:modelValue")?.[0]).toEqual([false]);
    });

    it("stores announcement data in localStorage when dismissed", async () => {
      const dialog = getDialog();
      const btn = dialog.find('[data-test="announcement-close"]');

      await btn.trigger("click");
      await flushPromises();

      const stored = localStorage.getItem("announcement");
      expect(stored).toBeTruthy();

      const decoded = JSON.parse(atob(stored as string));
      expect(decoded).toMatchObject({
        uuid: mockAnnouncement.uuid,
        title: mockAnnouncement.title,
        content: mockAnnouncement.content,
        date: mockAnnouncement.date,
      });
    });

    it("does not show duplicate announcements after dismissal", async () => {
      const dialog = getDialog();
      const btn = dialog.find('[data-test="announcement-close"]');

      await btn.trigger("click");
      await flushPromises();

      // Verify localStorage has the announcement marked as seen
      const stored = localStorage.getItem("announcement");
      const decoded = JSON.parse(atob(stored as string));
      expect(decoded.uuid).toBe(mockAnnouncement.uuid);
    });
  });

  describe("dialog visibility", () => {
    it("hides dialog when modelValue is false", () => {
      mountWrapper(mockAnnouncement, false);
      const dialog = getDialog();
      expect(dialog.exists()).toBe(false);
    });

    it("shows dialog when modelValue changes from false to true", async () => {
      mountWrapper(mockAnnouncement, false);

      await wrapper.setProps({ modelValue: true });
      await flushPromises();

      const dialog = getDialog();
      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).not.toContain("display: none;");
    });

    it("hides dialog when modelValue changes from true to false", async () => {
      mountWrapper(mockAnnouncement, true);

      await wrapper.setProps({ modelValue: false });
      await flushPromises();

      const dialog = getDialog();
      const dialogContent = dialog.find(".v-overlay__content");
      expect(dialogContent.attributes("style")).toContain("display: none;");
    });
  });

  describe("dialog properties", () => {
    beforeEach(() => mountWrapper());

    it("renders as a scrollable dialog", () => {
      const dialog = getDialog();
      expect(dialog.find(".v-card-text").attributes("style")).toContain("max-height: 70vh");
    });

    it("displays footer with dismiss button", () => {
      const dialog = getDialog();
      const footer = dialog.find('[data-test="window-dialog-footer"]');
      expect(footer.exists()).toBe(true);
      expect(footer.find('[data-test="announcement-close"]').exists()).toBe(true);
    });
  });
});
