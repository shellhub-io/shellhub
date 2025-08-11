import { defineStore } from "pinia";
import { ref } from "vue";
import { IConnector, IConnectorPayload } from "@/interfaces/IConnector";
import * as connectorApi from "../api/connectors";

const useConnectorStore = defineStore("connectors", () => {
  const connectors = ref<IConnector[]>([]);
  const connector = ref<IConnector>({} as IConnector);
  const connectorInfo = ref<object>({});
  const connectorCount = ref(0);

  const fetchConnectorList = async (data: { page: number; perPage: number }) => {
    try {
      const res = await connectorApi.getConnectorList(data.page, data.perPage);
      if (res.data.length) {
        connectors.value = res.data as IConnector[];
        connectorCount.value = parseInt(res.headers["x-total-count"], 10);
        return;
      }
      connectors.value = [];
      connectorCount.value = 0;
    } catch (error) {
      connectors.value = [];
      connectorCount.value = 0;
      throw error;
    }
  };

  const fetchConnectorById = async (uid: string) => {
    const res = await connectorApi.getConnector(uid);
    connector.value = res.data as IConnector;
  };

  const getConnectorInfo = async (uid: string) => {
    const res = await connectorApi.getConnectorInfo(uid);
    connectorInfo.value = res.data;
  };

  const createConnector = async (data: Omit<IConnectorPayload, "uid">) => {
    await connectorApi.createConnector(data);
  };

  const updateConnector = async (data: IConnectorPayload) => {
    await connectorApi.updateConnector(data);
  };

  const deleteConnector = async (uid: string) => {
    await connectorApi.deleteConnector(uid);
  };

  return {
    connectors,
    connector,
    connectorInfo,
    connectorCount,
    fetchConnectorList,
    fetchConnectorById,
    getConnectorInfo,
    createConnector,
    updateConnector,
    deleteConnector,
  };
});

export default useConnectorStore;
