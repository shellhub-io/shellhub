import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import { ref } from "vue";
import Namespace from "@/components/Namespace/Namespace.vue";
import { mountComponent } from "@tests/utils/mount";
import { mockNamespace } from "@tests/views/mocks/namespace";

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

vi.mock("vue-router", () => ({
  useRouter: () => ({ push: vi.fn() }),
}));

const menuStub = {
  template: "<div><slot name=\"activator\" :props=\"{}\" /><slot /></div>",
};

const mountNamespace = (smAndDownValue: boolean) => {
  displayState.smAndDown.value = smAndDownValue;
  displayState.mdAndDown.value = smAndDownValue;
  localStorage.setItem("tenant", mockNamespace.tenant_id);

  return mountComponent(Namespace, {
    piniaOptions: {
      initialState: {
        namespaces: {
          currentNamespace: mockNamespace,
          namespaceList: [mockNamespace],
        },
      },
    },
    global: {
      stubs: {
        VMenu: menuStub,
        "v-menu": menuStub,
        NamespaceAdd: true,
        NamespaceInstructions: true,
        NamespaceChip: true,
        NamespaceListItem: true,
        AdminConsoleItem: true,
        CopyWarning: true,
      },
    },
  });
};

describe("Namespace Component", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    localStorage.clear();
  });

  it("Renders the namespace name with default max width", () => {
    const wrapper = mountNamespace(false);
    const nameSpan = wrapper.find(".text-truncate");

    expect(nameSpan.exists()).toBe(true);
    expect(nameSpan.text()).toBe(mockNamespace.name);
    expect(nameSpan.attributes("style")).toContain("max-width: 220px");
  });

  it("Uses compact max width on small screens", () => {
    const wrapper = mountNamespace(true);
    const nameSpan = wrapper.find(".text-truncate");

    expect(nameSpan.exists()).toBe(true);
    expect(nameSpan.attributes("style")).toContain("max-width: 4rem");
  });
});
