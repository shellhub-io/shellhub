<template>
  <DataTable
    :headers="headers"
    :items="announcements"
    :itemsPerPage="itemsPerPage"
    :nextPage="next"
    :previousPage="prev"
    :loading="loading"
    :totalCount="numberAnnouncements"
    :page="page"
    :actualPage="page"
    @changeItemsPerPage="changeItemsPerPage"
    @clickNextPage="next"
    @clickPreviousPage="prev"
    data-test="announcement-list"
  >
    <template v-slot:rows>
      <tr v-for="(announcement, index) in announcements" :key="index">
        <td class="text-left" data-test="announcement-uuid">
          <v-chip>
            {{ announcement.uuid }}
          </v-chip>
        </td>
        <td class="text-left" data-test="announcement-title">
          {{ announcement.title }}
        </td>
        <td class="text-left">
          {{ formatDate(announcement.date) }}
        </td>
        <td class="text-left" data-test="announcement-actions">
          <v-tooltip bottom anchor="bottom">
            <template v-slot:activator="{ props }">
              <v-icon
                tag="a"
                dark
                v-bind="props"
                @click="redirectToAnnouncement(announcement)"
                @keyup.enter="redirectToAnnouncement(announcement)"
                tabindex="0"
              >mdi-information
              </v-icon>
            </template>
            <span>Info</span>
          </v-tooltip>

          <AnnouncementEdit :announcementItem="announcement" @update="refreshAnnouncements" />

          <AnnouncementDelete :uuid="announcement.uuid" @update="refreshAnnouncements" />
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
import useSnackbarStore from "@admin/store/modules/snackbar";
import { INotificationsError } from "../../interfaces/INotifications";
import DataTable from "../DataTable.vue";
import AnnouncementDelete from "./AnnouncementDelete.vue";
import AnnouncementEdit from "./AnnouncementEdit.vue";
import { IAnnouncements } from "../../interfaces/IAnnouncements";

const router = useRouter();
const announcementStore = useAnnouncementStore();
const snackbarStore = useSnackbarStore();
const itemsPerPage = ref(10);
const loading = ref(false);
const page = ref(1);
const headers = ref([
  {
    text: "Id",
    value: "uuid",
    align: "left",
  },
  {
    text: "Title",
    value: "title",
    align: "left",
  },
  {
    text: "Date",
    value: "date",
    align: "left",
  },
  {
    text: "Actions",
    value: "actions",
    align: "left",
  },
]);

const getAnnouncements = async (
  perPagaeValue: number,
  pageValue: number,
) => {
  try {
    loading.value = true;
    const hasAnnouncements = await announcementStore.fetchAnnouncements({ perPage: perPagaeValue, page: pageValue, orderBy: "desc" });
    if (!hasAnnouncements && page.value > 1) {
      page.value--;
    }
    loading.value = false;
  } catch (error) {
    snackbarStore.showSnackbarErrorAction(INotificationsError.announcementList);
  }
};
onMounted(async () => {
  try {
    loading.value = true;
    getAnnouncements(itemsPerPage.value, page.value);
  } catch (error) {
    snackbarStore.showSnackbarErrorAction(INotificationsError.announcementList);
  } finally {
    loading.value = false;
  }
});
const next = async () => {
  await getAnnouncements(itemsPerPage.value, ++page.value);
};
const prev = async () => {
  try {
    if (page.value > 1) await getAnnouncements(itemsPerPage.value, --page.value);
  } catch (error) {
    snackbarStore.showSnackbarErrorAction(INotificationsError.announcementList);
  }
};
const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};
watch(itemsPerPage, async () => {
  await getAnnouncements(itemsPerPage.value, page.value);
});
const announcements = computed(
  () => announcementStore.getAnnouncements as Array<IAnnouncements>,
);
const numberAnnouncements = computed(
  () => announcementStore.getNumberAnnouncements,
);

const refreshAnnouncements = async () => {
  await getAnnouncements(itemsPerPage.value, page.value);
};

const formatDate = (date: string) => moment(date).format("LL");

const redirectToAnnouncement = (announcement: IAnnouncements) => {
  router.push({
    name: "announcementDetails",
    params: { uuid: announcement.uuid },
  });
};
</script>
