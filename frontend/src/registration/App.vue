<template>
  <div>
    <div
      v-show="alert.message"
      class="alert alert-danger animate__animated animate__fadeIn"
      role="alert"
    >
      {{ alert.message }}
    </div>
    <div class="container">
      <div class="bg-light p-5 rounded-lg m-3">
        <h1 class="display-2 text-center">Retro Simply</h1>
        <p class="lead text-center">
          <i
            >A simple,
            <a href="https://github.com/safe-waters/retro-simply">open source</a
            >, free tool to run retrospectives for your team.</i
          >
        </p>
        <hr class="my-4" />
        <div class="row justify-content-center">
          <label
            class="col-md-3 m-3 btn btn-primary btn-lg"
            @click="showCreateRoom"
            >Create room</label
          >
          <label
            class="col-md-3 m-3 btn btn-primary btn-lg"
            @click="showJoinRoom"
            >Join room</label
          >
        </div>
        <create-join-room
          class="animate__animated animate__fadeIn"
          v-show="create.show"
          type="create"
          title="Create room"
        ></create-join-room>
        <create-join-room
          class="animate__animated animate__fadeIn"
          v-show="join.show"
          type="join"
          title="Join room"
        ></create-join-room>
      </div>
    </div>
  </div>
</template>

<script>
import CreateJoinRoom from "./components/CreateJoinRoom.vue";
import { mapGetters } from "vuex";

export default {
  name: "App",
  mounted: function () {
    if (!navigator.cookieEnabled) {
      this.$store.commit(
        "setAlertMessage",
        "cookies are disabled - enable cookies for authentication and authorization"
      );
    }
  },
  methods: {
    showCreateRoom: function () {
      this.$store.commit("showCreateRoom");
    },
    showJoinRoom: function () {
      this.$store.commit("showJoinRoom");
    },
  },
  components: {
    "create-join-room": CreateJoinRoom,
  },
  computed: {
    ...mapGetters(["alert", "create", "join"]),
  },
};
</script>