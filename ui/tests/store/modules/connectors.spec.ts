import { createStore, Store } from "vuex";
import axios from "axios";
import MockAdapter from "axios-mock-adapter";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { namespacesApi } from "@/api/http";
import { connectors, ConnectorState } from "@/store/modules/connectors";
import { State } from "@/store";
import { IConnector } from "@/interfaces/IConnector";

// eslint-disable-next-line @typescript-eslint/no-unused-vars
let mockNamespaces: MockAdapter;

// Mock axios
const mock = new MockAdapter(axios);

// Define initial state
const initialState: ConnectorState = {
  connectors: [],
  connector: {} as IConnector,
  info: {},
  page: 1,
  perPage: 10,
  numberConnectors: 0,
};

const createVuexStore = (state = initialState) => createStore<State>({
  modules: {
    connectors: {
      ...connectors,
      state,
    },
  },
});

describe("Connectors Vuex Module", () => {
  let store: Store<State>;

  beforeEach(() => {
    store = createVuexStore();
    mockNamespaces = new MockAdapter(namespacesApi.getAxios());
  });

  afterEach(() => {
    mock.reset();
  });

  describe("state", () => {
    it("should have initial state", () => {
      expect(store.state.connectors.connectors).toEqual([]);
      expect(store.state.connectors.connector).toEqual({});
      expect(store.state.connectors.info).toEqual({});
      expect(store.state.connectors.page).toBe(1);
      expect(store.state.connectors.perPage).toBe(10);
      expect(store.state.connectors.numberConnectors).toBe(0);
    });
  });

  describe("getters", () => {
    it("should return connectors list", () => {
      expect(store.getters["connectors/list"]).toEqual([]);
    });

    it("should return a connector", () => {
      expect(store.getters["connectors/get"]).toEqual({});
    });

    it("should return connector info", () => {
      expect(store.getters["connectors/getInfo"]).toEqual({});
    });

    it("should return page number", () => {
      expect(store.getters["connectors/getPage"]).toBe(1);
    });

    it("should return perPage value", () => {
      expect(store.getters["connectors/getPerPage"]).toBe(10);
    });

    it("should return number of connectors", () => {
      expect(store.getters["connectors/getNumberConnectors"]).toBe(0);
    });
  });

  describe("mutations", () => {
    it("should set connectors", () => {
      const connectorsData = [
        { id: 1, name: "Connector 1" },
        { id: 2, name: "Connector 2" },
      ];
      const headers = { "x-total-count": "2" };

      store.commit("connectors/setConnectors", { data: connectorsData, headers });

      expect(store.state.connectors.connectors).toEqual(connectorsData);
      expect(store.state.connectors.numberConnectors).toBe(2);
    });

    it("should set page and perPage", () => {
      const data = { page: 2, perPage: 20 };

      store.commit("connectors/setPagePerpage", data);

      expect(store.state.connectors.page).toBe(2);
      expect(store.state.connectors.perPage).toBe(20);
    });

    it("should set a connector", () => {
      const connectorData = { id: 1, name: "Connector 1" };

      store.commit("connectors/setConnector", connectorData);

      expect(store.state.connectors.connector).toEqual(connectorData);
    });

    it("should set connector info", () => {
      const infoData = { info: "Some info" };

      store.commit("connectors/setInfoConnector", infoData);

      expect(store.state.connectors.info).toEqual(infoData);
    });

    it("should clear connectors list", () => {
      store.commit("connectors/clearListConnector");

      expect(store.state.connectors.connectors).toEqual([]);
      expect(store.state.connectors.numberConnectors).toBe(0);
    });

    it("should clear a connector", () => {
      store.commit("connectors/clearConnector");

      expect(store.state.connectors.connector).toEqual({});
      expect(store.state.connectors.info).toEqual({});
    });
  });
});
