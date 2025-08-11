import { defineStore } from "pinia";
import { ref } from "vue";
import * as containerApi from "../api/containers";
import { FetchContainerParams, IContainer, IContainerRename } from "@/interfaces/IContainer";

const useContainersStore = defineStore("containers", () => {
  const containers = ref<IContainer[]>([]);
  const container = ref<IContainer>({} as IContainer);
  const containerCount = ref(0);
  const showContainers = ref(false);

  const fetchContainerList = async (data?: FetchContainerParams) => {
    try {
      const res = await containerApi.fetchContainers(
        data?.page || 1,
        data?.perPage || 10,
        data?.status || "accepted",
        data?.filter,
        data?.sortField,
        data?.sortOrder,
      );

      containers.value = res.data as IContainer[];
      containerCount.value = parseInt(res.headers["x-total-count"], 10);
      showContainers.value = true;
    } catch (error) {
      containers.value = [];
      containerCount.value = 0;
      throw error;
    }
  };

  const acceptContainer = async (uid: string) => {
    await containerApi.acceptContainer(uid);
  };

  const rejectContainer = async (uid: string) => {
    await containerApi.rejectContainer(uid);
  };

  const removeContainer = async (uid: string) => {
    await containerApi.removeContainer(uid);
  };

  const renameContainer = async (data: IContainerRename) => {
    await containerApi.renameContainer(data);
    container.value.name = data.name.name as string;
  };

  const getContainer = async (uid: string) => {
    try {
      const res = await containerApi.getContainer(uid);
      container.value = res.data as IContainer;
    } catch (error) {
      container.value = {} as IContainer;
      throw error;
    }
  };

  return {
    containers,
    container,
    containerCount,
    showContainers,
    fetchContainerList,
    removeContainer,
    renameContainer,
    getContainer,
    acceptContainer,
    rejectContainer,
  };
});

export default useContainersStore;
