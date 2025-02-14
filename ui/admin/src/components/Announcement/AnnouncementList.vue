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

          <AnnouncementEdit :announcement="announcement" @update="refreshAnnouncements" />

          <AnnouncementDelete :uuid="announcement.uuid" @update="refreshAnnouncements" />
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import moment from "moment";
import { useStore } from "../../store";
import { INotificationsError } from "../../interfaces/INotifications";
import DataTable from "../DataTable.vue";
import displayOnlyTenCharacters from "../../hooks/string";
import AnnouncementDelete from "./AnnouncementDelete.vue";
import AnnouncementEdit from "./AnnouncementEdit.vue";
import { IAnnouncements } from "../../interfaces/IAnnouncements";

export default defineComponent({
  name: "AnnouncementList",
  setup() {
    const store = useStore();
    const router = useRouter();
    const itemsPerPage = ref(10);
    const loading = ref(false);
    const page = ref(1);

    const getAnnouncements = async (
      perPagaeValue: number,
      pageValue: number,
    ) => {
      try {
        loading.value = true;
        const hasAnnouncements = await store.dispatch(
          "announcement/getAnnouncements",
          {
            perPage: perPagaeValue,
            page: pageValue,
          },
        );
        if (!hasAnnouncements && page.value > 1) {
          page.value--;
        }
        loading.value = false;
      } catch (error) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.announcementList,
        );
      }
    };
    onMounted(async () => {
      try {
        loading.value = true;
        getAnnouncements(itemsPerPage.value, page.value);
      } catch (error) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.announcementList,
        );
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
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.announcementList,
        );
      }
    };
    const changeItemsPerPage = async (newItemsPerPage: number) => {
      itemsPerPage.value = newItemsPerPage;
    };
    watch(itemsPerPage, async () => {
      await getAnnouncements(itemsPerPage.value, page.value);
    });
    const announcements = computed(
      () => store.getters["announcement/announcements"],
    );
    const numberAnnouncements = computed(
      () => store.getters["announcement/numberAnnouncements"],
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

    return {
      headers: [
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
      ],
      announcements,
      numberAnnouncements,
      loading,
      itemsPerPage,
      page,
      next,
      prev,
      changeItemsPerPage,
      displayOnlyTenCharacters,
      formatDate,
      refreshAnnouncements,
      redirectToAnnouncement,
    };
  },
  components: { DataTable, AnnouncementDelete, AnnouncementEdit },
});
</script>
