<template>
<v-card>
    <v-app-bar flat color="transparent">
        <v-toolbar-title>Sessions</v-toolbar-title>
    </v-app-bar>
    <v-divider></v-divider>
    <v-card-text class="pa-0">
    <v-data-table :headers="headers" :items="$store.getters['sessions/list']" item-key="uid" hide-default-footer>
        <template v-slot:items="props">
            <td>{{ props.item.device }}</td>
            <td>{{ props.item.username }}</td>
            <td>{{ props.item.active }}</td>
            <td>{{ props.item.ip_address }}</td>
            <td>{{ props.item.started_at | moment("ddd D MMM YYYY HH:mm:ss") }}</td>
            <td>{{ props.item.last_seen | moment("ddd D MMM YYYY HH:mm:ss") }}</td>
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
          text: "Active",
          value: "active"
        },
        {
          text: "IP Address",
          value: "ip_address"
        },
        {
          text: "Started At",
          value: "started_at"
        },
        {
          text: "Last Seen",
          value: "last_seen"
        }
      ]
    };
  },

  created() {
    this.$store.dispatch("sessions/fetch");
  }
};
</script>
