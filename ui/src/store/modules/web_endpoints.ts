import { defineStore } from "pinia";
import { ref } from "vue";
import * as webEndpointsApi from "../api/web_endpoints";
import { FetchWebEndpointsParams, IWebEndpoint, IWebEndpointsCreate } from "@/interfaces/IWebEndpoints";

const useWebEndpointsStore = defineStore("webEndpoints", () => {
  const webEndpoints = ref<Array<IWebEndpoint>>([]);
  const webEndpointCount = ref(0);
  const showWebEndpoints = ref(false);

  const fetchWebEndpointsList = async (data?: FetchWebEndpointsParams) => {
    try {
      const res = await webEndpointsApi.getWebEndpoints(
        data?.page || 1,
        data?.perPage || 10,
        data?.filter,
        data?.sortField,
        data?.sortOrder,
      );
      if (res.data.length) {
        showWebEndpoints.value = true;
      }
      webEndpoints.value = res.data as IWebEndpoint[];
      webEndpointCount.value = parseInt(res.headers["x-total-count"] as string, 10) || 0;
    } catch (error) {
      webEndpoints.value = [];
      webEndpointCount.value = 0;
      showWebEndpoints.value = false;
      throw error;
    }
  };

  const createWebEndpoint = async (data: IWebEndpointsCreate) => {
    const res = await webEndpointsApi.createWebEndpoint(data);
    return res;
  };

  const deleteWebEndpoint = async (address: string) => {
    const res = await webEndpointsApi.deleteWebEndpoint(address);
    return res;
  };

  return {
    webEndpoints,
    webEndpointCount,
    showWebEndpoints,
    fetchWebEndpointsList,
    deleteWebEndpoint,
    createWebEndpoint,
  };
});

export default useWebEndpointsStore;
