<template>
  <fragment>
    <div class="mr-4">
      <v-menu
        offset-y
        :close-on-content-click="false"
      >
        <template #activator="{ on, attrs }">
          <v-btn
            color="primary"
            dark
            v-bind="attrs"
            data-test="tags-btn"
            v-on="on"
            @click="getTags"
          >
            Tags

            <v-icon class="ml-3">
              mdi-chevron-down
            </v-icon>
          </v-btn>
        </template>

        <v-list>
          <v-list-item
            v-for="(item, index) in getListTags"
            :key="index"
            :data-test="item + '-item'"
          >
            <v-list-item-action>
              <v-checkbox
                v-model="selectedTags"
                :value="item"
              />
            </v-list-item-action>

            <v-list-item-title
              :data-test="item + '-title'"
              v-text="item"
            />
          </v-list-item>
        </v-list>
      </v-menu>
    </div>
  </fragment>
</template>

<script>

export default {
  name: 'TagSelector',

  data() {
    return {
      selectedTags: [],
    };
  },

  computed: {
    getListTags() {
      return this.$store.getters['tags/list'];
    },
  },

  watch: {
    selectedTags(item) {
      this.getDevices(item);
    },
  },

  created() {
    this.getTags();
  },

  methods: {
    async getTags() {
      await this.$store.dispatch('tags/fetch');
    },

    async getDevices(item) {
      let encodedFilter = null;

      if (item.length > 0) {
        const filter = [{ type: 'property', params: { name: 'tags', operator: 'contains', value: item } }];
        encodedFilter = btoa(JSON.stringify(filter));
      }
      await this.$store.dispatch('devices/setFilter', encodedFilter);

      try {
        this.$store.dispatch('devices/refresh');
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorDefault');
        }
      }
    },
  },
};
</script>
