<template>
  <v-table v-bind="$attrs" data-test="tagListList-dataTable" class="bg-v-theme-surface">
    <thead>
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          :class="head.align ? `text-${head.align}` : 'text-center'"
        >
          <span> {{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody v-if="tags.length">
      <tr v-for="(tag, i) in tags" :key="i">
        <td class="text-center">{{ tag }}</td>
        <td class="text-center">
          <v-menu location="bottom" scrim eager>
            <template v-slot:activator="{ props }">
              <v-chip v-bind="props" density="comfortable" size="small">
                <v-icon>mdi-dots-horizontal</v-icon>
              </v-chip>
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorizationEdit()">
                <template v-slot:activator="{ props }">
                  <div v-bind="props">
                    <TagEdit
                      :tag="tag"
                      :not-has-authorization="!hasAuthorizationEdit()"
                      @update="getTags()"
                    />
                  </div>
                </template>
                <span> You don't have this kind of authorization. </span>
              </v-tooltip>

              <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorizationRemove()">
                <template v-slot:activator="{ props }">
                  <div v-bind="props">
                    <TagRemove
                      :tag="tag"
                      :not-has-authorization="!hasAuthorizationRemove()"
                      @update="getTags()"
                    />
                  </div>
                </template>
                <span> You don't have this kind of authorization. </span>
              </v-tooltip>
            </v-list>
          </v-menu>
        </td>
      </tr>
    </tbody>
    <div v-else class="text-start mt-2 text-medium-emphasis">
      <p>No data avaliabe</p>
    </div>
  </v-table>
</template>

<script lang="ts">
import { defineComponent, computed, onMounted } from "vue";
import { useStore } from "../../store";
import { actions, authorizer } from "../../authorizer";
import hasPermission from "../../utils/permission";
import TagRemove from "./TagRemove.vue";
import TagEdit from "./TagEdit.vue";
import { INotificationsError } from "../../interfaces/INotifications";

export default defineComponent({
  inheritAttrs: true,
  setup() {
    const store = useStore();
    const tags = computed(() => store.getters["tags/list"]);
    const hasAuthorizationEdit = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.tag.edit);
      }
      return false;
    };
    const hasAuthorizationRemove = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(authorizer.role[role], actions.tag.remove);
      }
      return false;
    };
    const getTags = async () => {
      try {
        await store.dispatch("tags/fetch");
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.deviceTagList,
        );
        throw new Error(error);
      }
    };
    onMounted(() => {
      getTags();
    });
    const updateTags = () => {
      getTags();
    };
    return {
      headers: [
        {
          text: "Name",
          value: "name",
          align: "center",
          sortable: false,
        },
        {
          text: "Actions",
          value: "actions",
          align: "center",
          sortable: false,
        },
      ],
      tags,
      hasAuthorizationEdit,
      hasAuthorizationRemove,
      updateTags,
      getTags,
    };
  },
  components: { TagRemove, TagEdit },
});
</script>
