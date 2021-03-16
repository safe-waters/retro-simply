<template>
  <div>
    <div
      v-show="errorMessage"
      class="alert alert-danger animate__animated animate__fadeIn"
      role="alert"
    >
      {{ errorMessage }}
    </div>
    <div class="container" style="height: 100vh">
      <span
        @click="sortByNumVotes"
        class="fas fa-sort text-dark ps-1"
        :class="{ grow: isSortable }"
        :style="{ cursor: isSortable ? 'pointer' : '' }"
        >&nbsp;&nbsp;Sort</span
      >
      <div class="row" style="height: 90%">
        <retro-column
          class="col-sm"
          v-for="column in columns"
          :key="column.id"
          :column="column"
        />
      </div>
    </div>
  </div>
</template>

<script>
import RetroColumn from "./components/RetroColumn.vue";
import { mapGetters } from "vuex";

const DISPLAY_ERROR_TIME = 2000;

function init(inst) {
  inst.$store.commit("connect");

  let self = inst;

  inst.$store.state.ws.onopen = function () {
    self.$store.commit("setConnected", true);
    self.$store.commit("setErrorMessage", "");
  };

  inst.$store.state.ws.onmessage = function (e) {
    let newState = JSON.parse(e.data);
    self.$store.commit("updateColumns", newState.columns);
  };

  inst.$store.state.ws.onclose = function () {
    self.$store.commit("setConnected", false);
    self.$store.commit(
      "setErrorMessage",
      "An error occurred - redirecting to the home page..."
    );

    setTimeout(function () {
      window.location = "/";
    }, DISPLAY_ERROR_TIME);

    return;
  };
}

export default {
  name: "App",
  created: function () {
    init(this);
  },
  computed: {
    columns: function () {
      return this.$store.state.columns;
    },
    isSortable: function () {
      for (let i = 0; i < this.columns.length; i++) {
        for (let j = 0; j < this.columns[i].groups.length; j++) {
          if (this.columns[i].groups[j].retroCards.length > 0) {
            return true;
          }
        }
      }

      return false;
    },
    ...mapGetters(["connected", "errorMessage"]),
  },
  methods: {
    sortByNumVotes: function () {
      if (this.isSortable) {
        this.$store.commit("sortByNumVotes");
      }
    },
  },
  components: {
    "retro-column": RetroColumn,
  },
};
</script>

<style scoped>
.grow {
  transition: all 0.1s;
}

.grow:hover {
  transform: scale(1.09);
}

.grow:active {
  transform: scale(1.09);
  opacity: 0.7;
}
</style>