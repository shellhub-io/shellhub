import { createVuetify } from "vuetify";
import { mount, VueWrapper } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DataTable from "../../../src/components/DataTable.vue";

const headersTable = [
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Email",
    value: "email",
  },
  {
    text: "Phone",
    value: "phone",
  },
  {
    text: "Address",
    value: "address",
  },
  {
    text: "City",
    value: "city",
  },
];

const itemsTable = [
  {
    id: 1,
    name: "John Doe",
    email: "jhon@email.com",
    phone: "123456789",
    address: "123 Street",
    city: "New York",
  },
  {
    id: 2,
    name: "Sophia Liss",
    email: "sohpia@email.com",
    phone: "123456789",
    address: "123 Street",
    city: "New York",
  },
];

describe("DataTable", () => {
  let wrapper: VueWrapper<InstanceType<typeof DataTable>>;

  beforeEach(() => {
    const vuetify = createVuetify();

    wrapper = mount(DataTable, {
      props: {
        headers: [],
        items: [],
        itemsPerPage: 10,
        comboboxOptions: [10, 20, 50, 100],
        loading: false,
        actualPage: 1,
        totalCount: 2,
        nextPage: vi.fn(),
        previousPage: vi.fn(),
      },
      global: {
        plugins: [vuetify],
      },
    });
  });

  it("Is a Vue instance", () => {
    expect(wrapper).toBeTruthy();
  });
  it("Renders the component", () => {
    expect(wrapper.html()).toMatchSnapshot();
  });

  it("Should show in the table 'No data avaliable' with pass prop items empty", () => {
    expect(wrapper.vm.items).toEqual([]);
    expect(wrapper.find(".pa-4.text-subtitle-2 > p").exists()).toBe(true);
    expect(wrapper.find(".pa-4.text-subtitle-2 > p").text()).toBe("No data available");
  });

  it("Should show the items data with pass prop items", async () => {
    await wrapper.setProps({
      headers: headersTable,
      items: itemsTable,
    });
    expect(wrapper.vm.headers).toEqual(headersTable);
    expect(wrapper.vm.items).toEqual(itemsTable);
  });

  it("Should be emmited 'clickNextPage' function when click on next page button", async () => {
    const emmited = wrapper.emitted();
    await wrapper.find(".mdi-chevron-right").trigger("click");
    expect(emmited.clickNextPage.length).toBeTruthy();
    expect(emmited.clickNextPage.length).toBe(1);
  });

  it("Should be emmited 'clickPrevPage' function when click on next page button", async () => {
    const emmited = wrapper.emitted();
    await wrapper.find(".mdi-chevron-right").trigger("click");
    expect(emmited.clickNextPage.length).toBeTruthy();
    expect(emmited.clickNextPage.length).toBe(1);
  });

  it("Should be emmited 'changeItemsPerPage' function when click on next page button", async () => {
    const emmited = wrapper.emitted();
    await wrapper.find("div.v-combobox").trigger("click");
    await wrapper.setProps({
      itemsPerPage: 20,
    });
    expect(emmited.changeItemsPerPage.length).toBeTruthy();
  });
});
