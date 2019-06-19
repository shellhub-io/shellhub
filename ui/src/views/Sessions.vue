<template>
<v-card>
    <v-toolbar card color="transparent">
        <v-toolbar-title>Sessions</v-toolbar-title>
    </v-toolbar>
    <v-divider></v-divider>
    <v-card-text class="pa-0">
    <v-data-table :headers="headers" :items="$store.getters['sessions/list']" item-key="uid" disable-initial-sort hide-actions>
        <template v-slot:items="props">
            <td>{{ props.item.device }}</td>
            <td>{{ props.item.username }}</td>
            <td>{{ props.item.started_at | moment("ddd D MMM YYYY HH:mm:ss") }}</td>
            <td>{{ props.item.finished_at | moment("ddd D MMM YYYY HH:mm:ss") }}</td>
        </template>
    </v-data-table>
    </v-card-text>
</v-card>
</template>

<script>
export default {
  data() {
    return {
      headers: [
        {
          text: "Device",
          value: "device"
        },
        {
          text: "Username",
          value: "username"
        },
        {
          text: "Started At",
          value: "started_at"
        },
        {
          text: "Finished At",
          value: "finished_at"
        }
      ]
    };
  },

  created() {
    this.$store.dispatch("sessions/fetch");
  }
};
</script>
