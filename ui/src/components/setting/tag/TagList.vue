<template>
  <fragment>
    <v-data-table
      :headers="headers"
      :items="getListTags"
      :server-items-length="getNumberTags"
      hide-default-footer
      data-test="tagListList-dataTable"
    >
      <template #[`item.name`]="{ item }">
        {{ item.name }}
      </template>

      <template #[`item.actions`]="{ item }">
        <TagFormDialog
          action="edit"
          :tag-name="item.name"
          data-test="tagFormDialog-component"
          @update="getTags()"
        />

        <TagDelete
          :tag-name="item.name"
          data-test="tagDelete-component"
          @update="getTags()"
        />
      </template>
    </v-data-table>
  </fragment>
</template>

<script>

import TagFormDialog from '@/components/setting/tag/TagFormDialog';
import TagDelete from '@/components/setting/tag/TagDelete';

export default {
  name: 'TagList',

  components: {
    TagFormDialog,
    TagDelete,
  },

  data() {
    return {
      headers: [
        {
          text: 'Name',
          value: 'name',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    getListTags() {
      return this.$store.getters['tags/list'].map((str) => ({ name: str }));
    },

    getNumberTags() {
      return this.$store.getters['tags/getNumberTags'];
    },
  },

  created() {
    this.getTags();
  },

  methods: {
    async getTags() {
      try {
        await this.$store.dispatch('tags/fetch');
      } catch (error) {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceTagList);
      }
    },
  },
};
</script>
