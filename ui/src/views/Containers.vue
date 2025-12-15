<template>
  <PageHeader
    icon="mdi-docker"
    title="Containers"
    overline="Container Management"
    :description="description"
    icon-color="primary"
    data-test="device-title"
  >
    <template #actions>
      <ContainerAdd />
    </template>
  </PageHeader>
  <div
    v-if="showContainers"
    class="mt-2"
    data-test="device-table-component"
  >
    <Containers />
  </div>

  <NoItemsMessage
    v-else
    class="mt-2"
    item="Containers"
    icon="mdi-server"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>In order to register a container on ShellHub, you need to configure a Docker Connector.</p>
      <p>
        To view and connect to your containers in ShellHub, please add a Docker Engine connector.
        This will allow you to connect to your Docker Engine and see all your containers here.
      </p>
    </template>
    <template #action>
      <ContainerAdd />
    </template>
  </NoItemsMessage>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Containers from "../components/Containers/Container.vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import ContainerAdd from "../components/Containers/ContainerAdd.vue";
import PageHeader from "../components/PageHeader.vue";
import useContainersStore from "@/store/modules/containers";

const description
  = "View and manage Docker containers from connected Docker Engine connectors. "
    + "Access and monitor your containerized applications.";

const containersStore = useContainersStore();
const showContainers = computed(() => containersStore.showContainers);
</script>
