import { describe, expect, it, afterEach, vi, beforeEach } from "vitest";
import { VueWrapper, flushPromises } from "@vue/test-utils";
import { mockSnackbar, mountComponent } from "@tests/utils/mount";
import { createAxiosError } from "@tests/utils/axiosError";
import { mockNamespace } from "@tests/mocks";
import Namespace from "@/components/Namespace/Namespace.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import { envVariables } from "@/envVariables";
import { ref } from "vue";
import handleError from "@/utils/handleError";
import { Router } from "vue-router";
import { createCleanRouter } from "@tests/utils/router";

const mockNamespaceList = [
  mockNamespace,
  {
    ...mockNamespace,
    tenant_id: "tenant-2",
    name: "namespace-2",
  },
  {
    ...mockNamespace,
    tenant_id: "tenant-3",
    name: "namespace-3",
  },
];

vi.mock("@/envVariables", () => ({
  envVariables: {
    isEnterprise: false,
    isCloud: false,
  },
}));

const displayState = {
  smAndDown: ref(false),
  mdAndDown: ref(false),
  thresholds: ref({
    xs: 0,
    sm: 600,
    md: 960,
    lg: 1280,
    xl: 1920,
    xxl: 2560,
  }),
};

vi.mock("vuetify", async () => {
  const actual = await vi.importActual<typeof import("vuetify")>("vuetify");
  return {
    ...actual,
    useDisplay: () => displayState,
  };
});

const menuStub = {
  template: '<div><slot name="activator" :props="{}" /><slot /></div>',
};

describe("Namespace", () => {
  let wrapper: VueWrapper<InstanceType<typeof Namespace>>;
  let namespacesStore: ReturnType<typeof useNamespacesStore>;
  let router: Router;

  const mountWrapper = (isAdminContext = false, hasNamespaces = true, isAdmin = false) => {
    localStorage.setItem("tenant", mockNamespace.tenant_id);
    localStorage.setItem("id", "user-1");
    router = createCleanRouter();

    wrapper = mountComponent(Namespace, {
      global: {
        plugins: [router],
        stubs: {
          VMenu: menuStub,
          "v-menu": menuStub,
        },
      },
      props: { isAdminContext },
      piniaOptions: {
        initialState: {
          namespaces: {
            currentNamespace: hasNamespaces ? mockNamespace : {},
            namespaceList: hasNamespaces ? mockNamespaceList : [],
          },
          auth: {
            id: "user-1",
            isAdmin,
          },
        },
      },
    });

    namespacesStore = useNamespacesStore();
  };

  afterEach(() => {
    wrapper?.unmount();
    vi.clearAllMocks();
    localStorage.clear();
  });

  describe("Menu display", () => {
    beforeEach(() => mountWrapper());

    it("Renders menu button", () => {
      const button = wrapper.find("button");
      expect(button.exists()).toBe(true);
    });

    it("Shows namespace name in button", () => {
      expect(wrapper.text()).toContain(mockNamespace.name);
    });

    it("Renders NamespaceChip component", () => {
      const chip = wrapper.find('[data-test="menu-namespace-chip"]');
      expect(chip.exists()).toBe(true);
    });

    it("Shows chevron down icon", () => {
      const icons = wrapper.findAll(".v-icon");
      const hasChevronIcon = icons.some((icon) => icon.classes().includes("mdi-chevron-down"));
      expect(hasChevronIcon).toBe(true);
    });
  });

  describe("Menu display - Admin context", () => {
    beforeEach(() => mountWrapper(true));

    it("Shows AdminConsoleItem in admin context", () => {
      const adminItem = wrapper.findComponent({ name: "AdminConsoleItem" });
      expect(adminItem.exists()).toBe(true);
    });

    it("Does not show namespace chip in admin context", () => {
      const chip = wrapper.find('[data-test="menu-namespace-chip"]');
      expect(chip.exists()).toBe(false);
    });
  });

  describe("Current namespace section", () => {
    beforeEach(() => mountWrapper());

    it("Shows Active Namespace subheader", () => {
      expect(wrapper.text()).toContain("Active Namespace");
    });

    it("Shows Settings button", () => {
      const settingsBtn = wrapper.find('[data-test="namespace-settings-btn"]');
      expect(settingsBtn.exists()).toBe(true);
      expect(settingsBtn.text()).toContain("Settings");
    });

    it("Shows Tenant ID section", () => {
      expect(wrapper.text()).toContain("Tenant ID");
      expect(wrapper.text()).toContain(mockNamespace.tenant_id);
    });

    it("Renders CopyWarning component for Tenant ID", () => {
      const copyWarning = wrapper.findComponent({ name: "CopyWarning" });
      expect(copyWarning.exists()).toBe(true);
    });
  });

  describe("Available namespaces section", () => {
    beforeEach(() => mountWrapper());

    it("Shows Switch Namespace subheader", () => {
      expect(wrapper.text()).toContain("Switch Namespace");
    });

    it("Renders NamespaceListItem for available namespaces", () => {
      const listItems = wrapper.findAllComponents({ name: "NamespaceListItem" });
      expect(listItems.length).toBeGreaterThan(0);
    });

    it("Does not show current namespace in available list", () => {
      const listItems = wrapper.findAllComponents({ name: "NamespaceListItem" });
      const currentNamespaceItem = listItems.find(
        (item) => item.props("namespace")?.tenant_id === mockNamespace.tenant_id && !item.props("active"),
      );
      expect(currentNamespaceItem).toBeUndefined();
    });
  });

  describe("Available namespaces - Admin context", () => {
    beforeEach(() => mountWrapper(true));
    it("Shows Available Namespaces subheader in admin context", () => {
      expect(wrapper.text()).toContain("Available Namespaces");
    });

    it("Shows all namespaces in admin context", () => {
      const listItems = wrapper.findAllComponents({ name: "NamespaceListItem" });
      expect(listItems.length).toBeGreaterThan(0);
    });
  });

  describe("Create Namespace button", () => {
    beforeEach(() => mountWrapper());
    it("Shows Create Namespace button", () => {
      const createBtn = wrapper.find('[data-test="create-namespace-btn"]');
      expect(createBtn.exists()).toBe(true);
    });

    it("Opens NamespaceAdd dialog when clicked", async () => {
      const createBtn = wrapper.find('[data-test="create-namespace-btn"]');
      await createBtn.trigger("click");
      await flushPromises();

      const namespaceAdd = wrapper.findComponent({ name: "NamespaceAdd" });
      expect(namespaceAdd.props("modelValue")).toBe(true);
    });

    it("Does not show Create Namespace button in admin context", () => {
      wrapper?.unmount();
      mountWrapper(true);

      const createBtn = wrapper.find('[data-test="create-namespace-btn"]');
      expect(createBtn.exists()).toBe(false);
    });
  });

  describe("Admin panel button", () => {
    it("Shows admin button when user is admin in Enterprise", () => {
      envVariables.isEnterprise = true;
      envVariables.isCloud = false;
      mountWrapper(false, true, true);

      const adminItem = wrapper.findAllComponents({ name: "AdminConsoleItem" });
      expect(adminItem.length).toBeGreaterThan(0);
    });

    it("Does not show admin button in Cloud version", () => {
      envVariables.isEnterprise = true;
      envVariables.isCloud = true;
      mountWrapper();
      const adminItems = wrapper.findAllComponents({ name: "AdminConsoleItem" });
      expect(adminItems.length).toBeLessThanOrEqual(1);
    });
  });

  describe("Responsive design", () => {
    it("Uses default max width for namespace name", () => {
      displayState.smAndDown.value = false;
      mountWrapper();
      const nameSpan = wrapper.find(".text-truncate");
      expect(nameSpan.attributes("style")).toContain("max-width: 220px");
    });

    it("Uses compact max width on small screens", () => {
      displayState.smAndDown.value = true;
      displayState.mdAndDown.value = true;

      mountWrapper();

      const nameSpan = wrapper.find(".text-truncate");
      expect(nameSpan.attributes("style")).toContain("max-width: 4rem");
    });
  });

  describe("Settings navigation", () => {
    it("Navigates to namespace settings when Settings button is clicked", async () => {
      mountWrapper();
      const routerPushSpy = vi.spyOn(router, "push").mockResolvedValue();
      const settingsBtn = wrapper.find('[data-test="namespace-settings-btn"]');
      await settingsBtn.trigger("click");
      await flushPromises();

      expect(routerPushSpy).toHaveBeenCalledWith({ name: "SettingNamespace" });
    });
  });

  describe("Namespace switching", () => {
    it("Calls switchNamespace when namespace is selected", async () => {
      mountWrapper();
      const listItems = wrapper.findAllComponents({ name: "NamespaceListItem" });
      const inactiveItem = listItems.find((item) => !item.props("active"));

      inactiveItem?.vm.$emit("select", "tenant-2");
      await flushPromises();

      expect(namespacesStore.switchNamespace).toHaveBeenCalledWith("tenant-2");
    });
  });

  describe("Instructions dialog", () => {
    it("Shows NamespaceInstructions when user has no namespaces", () => {
      mountWrapper(false, false);

      const instructions = wrapper.findComponent({ name: "NamespaceInstructions" });
      expect(instructions.exists()).toBe(true);
    });

    it("Does not show NamespaceInstructions in admin context", () => {
      mountWrapper(true, false);

      const instructions = wrapper.findComponent({ name: "NamespaceInstructions" });
      expect(instructions.exists()).toBe(true);
    });
  });

  describe("Error handling", () => {
    it("Calls switchNamespace on 404 error", async () => {
      mountWrapper();
      vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(createAxiosError(404, "Not Found"));
      await flushPromises();

      expect(namespacesStore.fetchNamespace).toHaveBeenCalled();
      expect(namespacesStore.switchNamespace).toHaveBeenCalledWith("fake-tenant-data");
    });

    it("Handles 500 error when loading namespace with tenant ID", async () => {
      const error = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(error);
      await flushPromises();

      expect(namespacesStore.fetchNamespace).toHaveBeenCalled();
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to load namespace");
      expect(handleError).toHaveBeenCalledWith(error);
    });

    it("Handles error when switching namespace", async () => {
      const switchNamespaceError = createAxiosError(500, "Internal Server Error");

      mountWrapper();
      vi.mocked(namespacesStore.fetchNamespace).mockRejectedValueOnce(createAxiosError(404, "Not Found"));
      vi.mocked(namespacesStore.switchNamespace).mockRejectedValueOnce(switchNamespaceError);
      await flushPromises();

      expect(namespacesStore.switchNamespace).toHaveBeenCalledWith("fake-tenant-data");
      expect(mockSnackbar.showError).toHaveBeenCalledWith("Failed to switch namespace");
      expect(handleError).toHaveBeenCalledWith(switchNamespaceError);
    });
  });
});
