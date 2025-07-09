<template>
  <DataTable
    :headers
    :items="announcements"
    v-model:itemsPerPage="itemsPerPage"
    v-model:page="page"
    :totalCount="announcementCount"
    :loading
    :itemsPerPageOptions="[10, 20, 50, 100]"
    data-test="announcement-list"
  >
    <template v-slot:rows>
      <tr v-for="(announcement, index) in announcements" :key="index">
        <td data-test="announcement-uuid">
          <v-chip>
            {{ announcement.uuid }}
          </v-chip>
        </td>
        <td data-test="announcement-title">
          {{ announcement.title }}
        </td>
        <td>
          {{ formatDate(announcement.date) }}
        </td>
        <td data-test="announcement-actions">
          <v-tooltip bottom anchor="bottom">
            <template v-slot:activator="{ props }">
              <v-icon
                tag="a"
                dark
                v-bind="props"
                @click="redirectToAnnouncement(announcement)"
                @keyup.enter="redirectToAnnouncement(announcement)"
                tabindex="0"
                icon="mdi-information"
              />
            </template>
            <span>Info</span>
          </v-tooltip>

          <AnnouncementEdit
            :announcementItem="announcement"
            @update="refreshAnnouncements"
          />

          <AnnouncementDelete
            :uuid="announcement.uuid"
            @update="refreshAnnouncements"
          />
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import moment from "moment";
import useAnnouncementStore from "@admin/store/modules/announcement";
import { IAdminAnnouncementShort } from "@admin/interfaces/IAnnouncements";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/DataTable.vue";
import AnnouncementDelete from "./AnnouncementDelete.vue";
import AnnouncementEdit from "./AnnouncementEdit.vue";
import handleError from "@/utils/handleError";

const router = useRouter();
const announcementStore = useAnnouncementStore();
const snackbar = useSnackbar();
const page = ref(1);
const itemsPerPage = ref(10);
const loading = ref(false);
const announcements = computed(
  () => announcementStore.getAnnouncements as Array<IAdminAnnouncementShort>,
);
const announcementCount = computed(
  () => announcementStore.getNumberAnnouncements,
);
const headers = ref([
  {
    text: "Id",
    value: "uuid",
  },
  {
    text: "Title",
    value: "title",
  },
  {
    text: "Date",
    value: "date",
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

const fetchAnnouncements = async () => {
  try {
    loading.value = true;
    await announcementStore.fetchAnnouncements({ perPage: itemsPerPage.value, page: page.value, orderBy: "desc" });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch announcements.");
  }
  loading.value = false;
};

const refreshAnnouncements = async () => {
  await fetchAnnouncements();
};

const formatDate = (date: string) => moment(date).format("LL");

const redirectToAnnouncement = (announcement: IAdminAnnouncementShort) => {
  router.push({
    name: "announcementDetails",
    params: { uuid: announcement.uuid },
  });
};

watch([itemsPerPage, page], async () => {
  await fetchAnnouncements();
});

onMounted(async () => {
  await fetchAnnouncements();
});

defineExpose({ itemsPerPage, page, loading, numberAnnouncements: announcementCount, announcements });
</script>
