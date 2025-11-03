import { defineStore } from "pinia";
import { ref } from "vue";
import { IConnector, IConnectorPayload } from "@/interfaces/IConnector";
import * as connectorsApi from "../api/connectors";

const useConnectorStore = defineStore("connectors", () => {
  const connectors = ref<IConnector[]>([]);
  const connector = ref<IConnector>({} as IConnector);
  const connectorInfo = ref<object>({});
  const connectorCount = ref(0);

  const fetchConnectorList = async (data: { page: number; perPage: number }) => {
    try {
      const res = await connectorsApi.getConnectorList(data.page, data.perPage);
      connectors.value = res.data as IConnector[];
      connectorCount.value = parseInt(res.headers["x-total-count"] as string, 10);
    } catch (error) {
      connectors.value = [];
      connectorCount.value = 0;
      throw error;
    }
  };

  const fetchConnectorById = async (uid: string) => {
    const res = await connectorsApi.getConnector(uid);
    connector.value = res.data as IConnector;
  };

  const getConnectorInfo = async (uid: string) => {
    const res = await connectorsApi.getConnectorInfo(uid);
    connectorInfo.value = res.data;
  };

  const createConnector = async (data: Omit<IConnectorPayload, "uid">) => {
    await connectorsApi.createConnector(data);
  };

  const updateConnector = async (data: IConnectorPayload) => {
    await connectorsApi.updateConnector(data);
  };

  const deleteConnector = async (uid: string) => {
    await connectorsApi.deleteConnector(uid);
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
