import { describe, expect, it } from "vitest";
import { store } from "../../../src/store";

describe('Announcement', () => {
  const announcements = 
  [
    {
      uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
      title: "This is a announcement",
      content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
      date: "2022-12-15T19:45:45.618Z"
    },
    {
      uuid: "52188548-2b99-4f38-ac09-3a8f8988476f",
      title: "This is a new announcement",
      content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
      date: "2022-12-15T19:46:45.618Z"
    },
  ];

  const announcement = {
    uuid: "52088548-2b99-4f38-ac09-3a8f8988476f",
    title: "This is a announcement",
    content: "## ShellHub new features \n - New feature 1 \n - New feature 2 \n - New feature 3",
    date: "2022-12-15T19:45:45.618Z"
  };

  it('Return device default variables', () => {
    expect(store.getters["announcement/list"]).toEqual([]);
    expect(store.getters["announcement/get"]).toEqual({});
    expect(store.getters["announcement/getPage"]).toEqual(1);
    expect(store.getters["announcement/getPerPage"]).toEqual(10);
    expect(store.getters["announcement/getOrderBy"]).toEqual("asc");
  });

  it('Verify initial states change for mutation setAnnouncements', () => {
    store.commit("announcement/setAnnouncements", announcements);
    expect(store.getters["announcement/list"]).toEqual(announcements);
  });

  it('Verify initial state change for mutation setAnnouncement', () => {
    store.commit("announcement/setAnnouncement", announcement);
    expect(store.getters["announcement/get"]).toEqual(announcement);
  });

  it('Verify initial state change for mutation setPageAndPerPage', () => {
    store.commit("announcement/setPageAndPerPage", { page: 2, perPage: 20 });
    expect(store.getters["announcement/getPage"]).toEqual(2);
    expect(store.getters["announcement/getPerPage"]).toEqual(20);
  });

  it('Verify initial state change for mutation setOrderBy', () => {
    store.commit("announcement/setOrderBy", "desc");
    expect(store.getters["announcement/getOrderBy"]).toEqual("desc");
  });

  it('Verify initial state change for mutation clearAnnouncements', () => {
    store.commit("announcement/clearAnnouncements");
    expect(store.getters["announcement/list"]).toEqual([]);
  });

  it('Verify initial state change for mutation clearAnnouncement', () => {
    store.commit("announcement/clearAnnouncement");
    expect(store.getters["announcement/get"]).toEqual({});
  });
});